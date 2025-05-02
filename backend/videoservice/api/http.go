package api

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"sortedstartup.com/stream/common/interceptors"
	"sortedstartup.com/stream/videoservice/db"
)

const (
	uploadDir     = "uploads" // Directory to store uploaded files
	maxUploadSize = 100 << 20 // Maximum file size limit: 100 MB
)

// uploadHandler handles file uploads
func (api *VideoAPI) uploadHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("uploadHandler")

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		slog.Error("Only POST method is allowed")
		return
	}

	authContext, err := interceptors.AuthFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		slog.Error("Unauthorized", "err", err)
		return
	}
	userID := authContext.User.ID

	// Enforce Content-Length header if provided
	if r.ContentLength > maxUploadSize {
		http.Error(w, "File size exceeds the 100 MB limit", http.StatusRequestEntityTooLarge)
		slog.Error("File size exceeds the 100 MB limit")
		return
	}

	// Limit the request body size for memory efficiency
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse the multipart form
	err = r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		slog.Error("Invalid form data", "err", err)
		return
	}

	// Retrieve the uploaded file
	file, fileHeader, err := r.FormFile("video") // "video" is the form field name
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		slog.Error("Error retrieving file", "err", err)
		return
	}
	defer file.Close()

	// Retrieve the title and description from the form
	title := r.FormValue("title")
	description := r.FormValue("description")

	// Validate file type
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".mp4" && ext != ".mov" && ext != ".avi" && ext != ".webm" {
		http.Error(w, "Unsupported file format. Only .mp4, .mov, .avi, .webm are allowed", http.StatusBadRequest)
		slog.Error("Unsupported file format. Only .mp4, .mov, .avi, .webm are allowed")
		return
	}

	// Generate a unique filename with the original extension
	uid := uuid.New().String()
	fileName := uid + ext

	// Resolve absolute path for the uploads directory
	absUploadDir, err := filepath.Abs(uploadDir)
	if err != nil {
		http.Error(w, "Failed to resolve upload directory", http.StatusInternalServerError)
		slog.Error("Failed to resolve upload directory", "err", err)
		return
	}
	outputPath := filepath.Join(absUploadDir, fileName)

	// Ensure the uploads directory exists
	if _, err := os.Stat(absUploadDir); os.IsNotExist(err) {
		err := os.Mkdir(absUploadDir, 0755)
		if err != nil {
			http.Error(w, "Error creating uploads directory", http.StatusInternalServerError)
			slog.Error("Error creating uploads directory", "err", err)
			return
		}
	}

	// Create the destination file for writing
	outFile, err := os.Create(outputPath)
	if err != nil {
		http.Error(w, "Unable to create file", http.StatusInternalServerError)
		slog.Error("Error creating file", "err", err)
		return
	}
	defer outFile.Close()

	// Stream the file content directly to disk
	_, err = io.Copy(outFile, file)
	if err != nil {
		// Check for MaxBytesError
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			slog.Error("File size exceeds the 100 MB limit", "err", err)
			http.Error(w, "File size exceeds the 100 MB limit", http.StatusRequestEntityTooLarge)
		} else {
			slog.Error("Failed to save file", "err", err)
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
		}

		// Delete the partially written file
		os.Remove(outputPath)
		slog.Error("Error saving file, partial file deleted", "err", err)
		return
	}

	// Save video details to the database, including title and description
	err = api.DBQueries.CreateVideoUploaded(r.Context(), db.CreateVideoUploadedParams{
		ID:             uid,
		Title:          title,
		Description:    description,
		Url:            outputPath,
		UploadedUserID: userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})
	if err != nil {
		slog.Error("Failed to add video to the database", "err", err)
		http.Error(w, "Failed to add video to the library", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "File uploaded successfully", "filename": "%s"}`, fileName)))
}

func (api *VideoAPI) serveVideoHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract video ID from URL path
	videoID := r.URL.Path[len("/video/"):]
	if videoID == "" {
		http.Error(w, "Video ID is required", http.StatusBadRequest)
		return
	}

	authContext, err := interceptors.AuthFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		slog.Error("Unauthorized", "err", err)
		return
	}
	userID := authContext.User.ID

	// Get video details from database
	video, err := api.DBQueries.GetVideoByID(r.Context(), db.GetVideoByIDParams{
		ID:     videoID,
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Video not found", http.StatusNotFound)
			return
		}
		api.log.Error("Failed to get video from database", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Open the video file
	// videoPath := filepath.Join(api.config.Storage.Path, video.ID)
	// file, err := os.Open(videoPath)

	file, err := os.Open(video.Url) // Use the URL field from the database
	if err != nil {
		api.log.Error("Failed to open video file", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Get file info for Content-Length header
	fileInfo, err := file.Stat()
	if err != nil {
		api.log.Error("Failed to get file info", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", "video/webm") // Adjust content type based on your video format
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Stream the file to the response
	// TODO: make it efficient and streaming
	if _, err := io.Copy(w, file); err != nil {
		api.log.Error("Failed to stream video file", "error", err)
		// Can't send error response here as we've already started writing the response
	}
}
