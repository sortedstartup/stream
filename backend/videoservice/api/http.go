package api

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
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
	maxUploadSize = 1024 << 20 // Maximum file size limit: 1024 MB
	maxFormParts  = 4          // Maximum number of multipart form parts allowed
)

func (api *VideoAPI) getVideoDir() string {
	uploadDir := ""
	if strings.TrimSpace(api.config.FileStoreDir) == "" {
		// get absolute path for current working dir and store in current working dir
		cwdDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		uploadDir = cwdDir
	} else {
		uploadDir = api.config.FileStoreDir
	}
	return uploadDir
}

// uploadHandler handles file uploads with streaming to prevent memory issues
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
		http.Error(w, "File size exceeds the 1024 MB limit", http.StatusRequestEntityTooLarge)
		slog.Error("File size exceeds the 1024 MB limit")
		return
	}

	// Limit the request body size for memory efficiency
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Get the multipart reader for streaming
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		slog.Error("Invalid multipart form", "err", err)
		return
	}

	var title, description string
	var videoPart *multipart.Part
	var originalFilename string
	var partCount int

	// Cleanup function for videoPart
	defer func() {
		if videoPart != nil {
			videoPart.Close()
		}
	}()

	// Helper function to read text field
	readTextField := func(part *multipart.Part, fieldName string, currentValue string) (string, error) {
		if currentValue != "" {
			return "", fmt.Errorf("duplicate %s field", fieldName)
		}
		data, err := io.ReadAll(part)
		if err != nil {
			return "", fmt.Errorf("error reading %s: %w", fieldName, err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	// Process all multipart fields
	for {
		partCount++
		if partCount > maxFormParts {
			http.Error(w, "Too many form parts", http.StatusBadRequest)
			slog.Error("Too many form parts", "count", partCount)
			return
		}

		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "Error reading multipart form", http.StatusBadRequest)
			slog.Error("Error reading multipart form", "err", err)
			return
		}

		formName := part.FormName()

		switch formName {
		case "title":
			title, err = readTextField(part, "title", title)
			if err != nil {
				part.Close()
				http.Error(w, err.Error(), http.StatusBadRequest)
				slog.Error("Title field error", "err", err)
				return
			}

		case "description":
			description, err = readTextField(part, "description", description)
			if err != nil {
				part.Close()
				http.Error(w, err.Error(), http.StatusBadRequest)
				slog.Error("Description field error", "err", err)
				return
			}

		case "video":
			if videoPart != nil {
				part.Close()
				http.Error(w, "Multiple video files not allowed", http.StatusBadRequest)
				slog.Error("Multiple video files not allowed")
				return
			}

			originalFilename = part.FileName()
			if originalFilename == "" {
				part.Close()
				http.Error(w, "No file selected", http.StatusBadRequest)
				slog.Error("No file selected")
				return
			}

			// Validate file type
			ext := strings.ToLower(filepath.Ext(originalFilename))
			if ext != ".mp4" && ext != ".webm" && ext != ".ogg" && ext != ".ogv" {
				part.Close()
				http.Error(w, "Unsupported file format. Only .mp4, .webm, .ogg, .ogv are allowed", http.StatusBadRequest)
				slog.Error("Unsupported file format", "ext", ext)
				return
			}

			videoPart = part // Don't close - will be handled by defer

		default:
			part.Close()
			http.Error(w, "Unknown form field: "+formName, http.StatusBadRequest)
			slog.Error("Unknown form field", "field", formName)
			return
		}

		// Close part unless it's the video part
		if formName != "video" {
			part.Close()
		}
	}

	// Validate required fields
	if title == "" || description == "" {
		http.Error(w, "Title and description are required", http.StatusBadRequest)
		slog.Error("Missing required fields", "hasTitle", title != "", "hasDescription", description != "")
		return
	}

	if videoPart == nil {
		http.Error(w, "No video file found", http.StatusBadRequest)
		slog.Error("No video file found")
		return
	}

	// Process the video file
	ext := strings.ToLower(filepath.Ext(originalFilename))
	uid := uuid.New().String()
	fileName := uid + ext

	absUploadDir, err := filepath.Abs(api.getVideoDir())
	if err != nil {
		http.Error(w, "Failed to resolve upload directory", http.StatusInternalServerError)
		slog.Error("Failed to resolve upload directory", "err", err)
		return
	}

	if _, err := os.Stat(absUploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(absUploadDir, 0755); err != nil {
			http.Error(w, "Error creating uploads directory", http.StatusInternalServerError)
			slog.Error("Error creating uploads directory", "err", err)
			return
		}
	}

	outputPath := filepath.Join(absUploadDir, fileName)
	outFile, err := os.Create(outputPath)
	if err != nil {
		http.Error(w, "Unable to create file", http.StatusInternalServerError)
		slog.Error("Error creating file", "err", err)
		return
	}
	defer outFile.Close()

	// Stream the file content directly to disk
	_, err = io.Copy(outFile, videoPart)
	if err != nil {
		os.Remove(outputPath) // Cleanup on error

		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "File size exceeds the limit", http.StatusRequestEntityTooLarge)
			slog.Error("File size exceeds limit", "err", err)
		} else {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			slog.Error("Failed to save file", "err", err)
		}
		return
	}

	slog.Info("File streamed successfully", "filename", fileName, "original", originalFilename)

	// Save video details to the database
	err = api.dbQueries.CreateVideoUploaded(r.Context(), db.CreateVideoUploadedParams{
		ID:             uid,
		Title:          title,
		Description:    description,
		Url:            fileName,
		UploadedUserID: userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})
	if err != nil {
		// Clean up the uploaded file if database operation fails
		absUploadDir, _ := filepath.Abs(api.getVideoDir())
		outputPath := filepath.Join(absUploadDir, fileName)
		os.Remove(outputPath)

		slog.Error("Failed to add video to the database", "err", err)
		http.Error(w, "Failed to add video to the library", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "File uploaded successfully", "filename": "%s"}`, fileName)))
}

func getMimeTypeFromExtension(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".ogg":
		return "video/ogg"
	case ".ogv":
		return "video/ogg"
	default:
		return "application/octet-stream" // fallback
	}
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

	_, err := interceptors.AuthFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get video details from database (now allows any authenticated user to access any video)
	video, err := api.dbQueries.GetVideoByIDForAllUsers(r.Context(), videoID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Video not found", http.StatusNotFound)
			return
		}
		api.log.Error("Failed to get video from database", "error", err, "videoID", videoID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Build the file path
	videoFileName := filepath.Base(video.Url)
	absVideoPath := filepath.Join(api.getVideoDir(), videoFileName)

	// Open the video file
	file, err := os.Open(absVideoPath)
	if err != nil {
		api.log.Error("Failed to open video file", "error", err, "path", absVideoPath)
		http.Error(w, "Video file not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Get file info for ServeContent
	fileInfo, err := file.Stat()
	if err != nil {
		api.log.Error("Failed to get file info", "error", err, "path", absVideoPath)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set the appropriate content type
	contentType := getMimeTypeFromExtension(video.Url)
	w.Header().Set("Content-Type", contentType)

	// Cache video files for a longer period since they rarely change
	// 7 days = 604800 seconds
	w.Header().Set("Cache-Control", "public, max-age=604800") // Cache for 7 days

	// Use http.ServeContent to handle range requests, caching, and proper HTTP semantics
	http.ServeContent(w, r, videoFileName, fileInfo.ModTime(), file)
}
