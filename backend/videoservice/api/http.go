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
	maxUploadSize = 500 << 20 // Maximum file size limit: 500 MB (increased for large videos)
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
		http.Error(w, "File size exceeds the 500 MB limit", http.StatusRequestEntityTooLarge)
		slog.Error("File size exceeds the 500 MB limit")
		return
	}

	// Limit the request body size for memory efficiency
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse the multipart form with increased memory limit
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

	// Retrieve the title, description, and optional space_id from the form
	title := r.FormValue("title")
	description := r.FormValue("description")
	spaceID := r.FormValue("space_id") // Optional parameter

	// Validate file type - support more video formats
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	supportedFormats := []string{".mp4", ".mov", ".avi", ".webm", ".mkv", ".flv", ".wmv", ".m4v", ".3gp", ".ogv"}
	isValidFormat := false
	for _, format := range supportedFormats {
		if ext == format {
			isValidFormat = true
			break
		}
	}

	if !isValidFormat {
		http.Error(w, "Unsupported file format. Supported formats: mp4, mov, avi, webm, mkv, flv, wmv, m4v, 3gp, ogv", http.StatusBadRequest)
		slog.Error("Unsupported file format", "ext", ext, "filename", fileHeader.Filename)
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

	// Stream the file content directly to disk for large files
	_, err = io.Copy(outFile, file)
	if err != nil {
		// Check for MaxBytesError
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			slog.Error("File size exceeds the 500 MB limit", "err", err)
			http.Error(w, "File size exceeds the 500 MB limit", http.StatusRequestEntityTooLarge)
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
	now := time.Now()
	err = api.dbQueries.CreateVideoUploaded(r.Context(), db.CreateVideoUploadedParams{
		ID:             uid,
		Title:          title,
		Description:    description,
		Url:            outputPath,
		UploadedUserID: userID,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		slog.Error("Failed to add video to the database", "err", err)
		http.Error(w, "Failed to add video to the library", http.StatusInternalServerError)
		return
	}

	// If space_id is provided, add the video to that space
	if spaceID != "" {
		// First verify the user owns the space
		_, err = api.dbQueries.GetSpaceByID(r.Context(), db.GetSpaceByIDParams{
			ID:     spaceID,
			UserID: userID,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				// Space doesn't exist or user doesn't own it - log warning but don't fail upload
				slog.Warn("Space not found or user doesn't own it, skipping space assignment", "space_id", spaceID, "user_id", userID)
			} else {
				slog.Warn("Error verifying space ownership, skipping space assignment", "err", err, "space_id", spaceID)
			}
		} else {
			// Add video to space
			err = api.dbQueries.AddVideoToSpace(r.Context(), db.AddVideoToSpaceParams{
				VideoID:   uid,
				SpaceID:   spaceID,
				CreatedAt: now,
				UpdatedAt: now,
			})
			if err != nil {
				slog.Warn("Failed to add video to space, but upload was successful", "err", err, "video_id", uid, "space_id", spaceID)
			} else {
				slog.Info("Video successfully added to space", "video_id", uid, "space_id", spaceID)
			}
		}
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "File uploaded successfully", "filename": "%s", "video_id": "%s"}`, fileName, uid)))
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
	video, err := api.dbQueries.GetVideoByID(r.Context(), db.GetVideoByIDParams{
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
	file, err := os.Open(video.Url) // Use the URL field from the database
	if err != nil {
		api.log.Error("Failed to open video file", "error", err, "path", video.Url)
		http.Error(w, "Video file not found", http.StatusNotFound)
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

	// Determine content type based on file extension
	ext := strings.ToLower(filepath.Ext(video.Url))
	contentType := getVideoContentType(ext)

	// Set appropriate headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	w.Header().Set("Accept-Ranges", "bytes") // Enable range requests for video seeking

	// Handle range requests for video seeking
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		api.handleRangeRequest(w, r, file, fileInfo.Size(), contentType)
		return
	}

	// Stream the full file to the response
	if _, err := io.Copy(w, file); err != nil {
		api.log.Error("Failed to stream video file", "error", err)
		// Can't send error response here as we've already started writing the response
	}
}

// getVideoContentType returns the appropriate MIME type for video files
func getVideoContentType(ext string) string {
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".mkv":
		return "video/x-matroska"
	case ".flv":
		return "video/x-flv"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".m4v":
		return "video/x-m4v"
	case ".3gp":
		return "video/3gpp"
	case ".ogv":
		return "video/ogg"
	default:
		return "video/mp4" // Default fallback
	}
}

// handleRangeRequest handles HTTP range requests for video seeking
func (api *VideoAPI) handleRangeRequest(w http.ResponseWriter, r *http.Request, file *os.File, fileSize int64, contentType string) {
	rangeHeader := r.Header.Get("Range")

	// Parse range header (simple implementation for "bytes=start-end")
	var start, end int64
	if _, err := fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end); err != nil {
		// Try parsing "bytes=start-" format
		if _, err := fmt.Sscanf(rangeHeader, "bytes=%d-", &start); err != nil {
			http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
			return
		}
		end = fileSize - 1
	}

	// Validate range
	if start >= fileSize || end >= fileSize || start > end {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		http.Error(w, "Requested range not satisfiable", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// Seek to start position
	if _, err := file.Seek(start, 0); err != nil {
		api.log.Error("Failed to seek file", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set headers for partial content
	contentLength := end - start + 1
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusPartialContent)

	// Stream the requested range
	if _, err := io.CopyN(w, file, contentLength); err != nil {
		api.log.Error("Failed to stream video range", "error", err)
	}
}
