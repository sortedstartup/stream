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
	paymentProto "sortedstartup.com/stream/paymentservice/proto"
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

	// Get tenant ID from header
	tenantID := r.Header.Get("x-tenant-id")
	if tenantID == "" {
		http.Error(w, "x-tenant-id header is required", http.StatusBadRequest)
		slog.Error("x-tenant-id header is required")
		return
	}

	// Validate user has access to this tenant
	err = isUserInTenant(r.Context(), api.userServiceClient, api.log, tenantID, userID)
	if err != nil {
		http.Error(w, "Access denied: you are not a member of this tenant", http.StatusForbidden)
		slog.Error("Tenant access denied", "tenantID", tenantID, "userID", userID, "err", err)
		return
	}

	// Check payment access for storage upload
	slog.Info("Checking storage access for upload", "userID", userID)
	accessResp, err := api.paymentServiceClient.CheckUserAccess(r.Context(), &paymentProto.CheckUserAccessRequest{
		UserId:         userID,
		UsageType:      "storage",
		RequestedUsage: int64(r.ContentLength), // Check if the upload size would exceed limits
	})
	if err != nil {
		http.Error(w, "Payment service unavailable", http.StatusServiceUnavailable)
		slog.Error("Payment service error", "err", err, "userID", userID)
		return
	}

	if !accessResp.HasAccess {
		var errorMessage string
		switch accessResp.Reason {
		case "storage_limit_exceeded":
			errorMessage = "Upload failed: Storage limit exceeded. Please upgrade your plan to continue uploading."
		case "subscription_inactive":
			errorMessage = "Upload failed: Your subscription is inactive. Please reactivate to continue uploading."
		default:
			errorMessage = "Upload failed: Access denied. Please check your subscription status."
		}

		http.Error(w, errorMessage, http.StatusPaymentRequired) // 402 Payment Required
		slog.Warn("Upload blocked due to payment restrictions", "userID", userID, "reason", accessResp.Reason)
		return
	}

	// Log usage warning if near limit
	if accessResp.IsNearLimit && accessResp.WarningMessage != "" {
		slog.Warn("User approaching storage limit", "userID", userID, "warning", accessResp.WarningMessage)
	}

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

	// Read title (part 1)
	titlePart, err := reader.NextPart()
	if err != nil {
		http.Error(w, "Missing title field", http.StatusBadRequest)
		slog.Error("Missing title field", "err", err)
		return
	}
	defer titlePart.Close()

	if titlePart.FormName() != "title" {
		http.Error(w, "Expected title field first", http.StatusBadRequest)
		slog.Error("Expected title field first, got", "field", titlePart.FormName())
		return
	}

	titleData, err := io.ReadAll(titlePart)
	if err != nil {
		http.Error(w, "Error reading title", http.StatusBadRequest)
		slog.Error("Error reading title", "err", err)
		return
	}
	title := strings.TrimSpace(string(titleData))

	// Read description (part 2)
	descPart, err := reader.NextPart()
	if err != nil {
		http.Error(w, "Missing description field", http.StatusBadRequest)
		slog.Error("Missing description field", "err", err)
		return
	}
	defer descPart.Close()

	if descPart.FormName() != "description" {
		http.Error(w, "Expected description field second", http.StatusBadRequest)
		slog.Error("Expected description field second, got", "field", descPart.FormName())
		return
	}

	descData, err := io.ReadAll(descPart)
	if err != nil {
		http.Error(w, "Error reading description", http.StatusBadRequest)
		slog.Error("Error reading description", "err", err)
		return
	}
	description := strings.TrimSpace(string(descData))

	// Read channel_id (part 3)
	channelPart, err := reader.NextPart()
	if err != nil {
		http.Error(w, "Missing channel_id field", http.StatusBadRequest)
		slog.Error("Missing channel_id field", "err", err)
		return
	}
	defer channelPart.Close()

	if channelPart.FormName() != "channel_id" {
		http.Error(w, "Expected channel_id field third", http.StatusBadRequest)
		slog.Error("Expected channel_id field third, got", "field", channelPart.FormName())
		return
	}

	channelData, err := io.ReadAll(channelPart)
	if err != nil {
		http.Error(w, "Error reading channel_id", http.StatusBadRequest)
		slog.Error("Error reading channel_id", "err", err)
		return
	}
	channelID := strings.TrimSpace(string(channelData))

	// Read video file (part 4)
	videoPart, err := reader.NextPart()
	if err != nil {
		http.Error(w, "Missing video file", http.StatusBadRequest)
		slog.Error("Missing video file", "err", err)
		return
	}
	defer videoPart.Close()

	if videoPart.FormName() != "video" {
		http.Error(w, "Expected video field fourth", http.StatusBadRequest)
		slog.Error("Expected video field fourth, got", "field", videoPart.FormName())
		return
	}

	originalFilename := videoPart.FileName()
	if originalFilename == "" {
		http.Error(w, "No file selected", http.StatusBadRequest)
		slog.Error("No file selected")
		return
	}

	// Auto-generate title if not provided
	if title == "" {
		// Remove extension and use filename as title
		title = strings.TrimSuffix(originalFilename, filepath.Ext(originalFilename))
		if title == "" {
			// Fallback to timestamp if filename is empty
			title = "Recording " + time.Now().Format("2006-01-02 15:04")
		}
	}

	// Description is optional, can be empty

	// Validate file type
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext != ".mp4" && ext != ".webm" && ext != ".ogg" && ext != ".ogv" {
		http.Error(w, "Unsupported file format. Only .mp4, .webm, .ogg, .ogv are allowed", http.StatusBadRequest)
		slog.Error("Unsupported file format", "ext", ext)
		return
	}

	// Generate unique filename
	uid := uuid.New().String()
	fileName := uid + ext

	// Prepare upload directory
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

	// Create and stream to file immediately
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

	// Get file size to store in database
	fileInfo, err := outFile.Stat()
	if err != nil {
		os.Remove(outputPath)
		slog.Error("Failed to get file size for database storage", "err", err, "filename", fileName)
		http.Error(w, "Failed to process file", http.StatusInternalServerError)
		return
	}
	fileSizeForDB := fileInfo.Size()

	// Save video details to the database
	err = api.dbQueries.CreateVideoUploaded(r.Context(), db.CreateVideoUploadedParams{
		ID:             uid,
		Title:          title,
		Description:    description,
		Url:            fileName,
		UploadedUserID: userID,
		TenantID:       sql.NullString{String: tenantID, Valid: true},
		ChannelID:      sql.NullString{String: channelID, Valid: channelID != ""},
		IsPrivate:      sql.NullBool{Bool: true, Valid: true},  // All videos are private by default
		IsDeleted:      sql.NullBool{Bool: false, Valid: true}, // All videos start as not deleted
		FileSizeBytes:  sql.NullInt64{Int64: fileSizeForDB, Valid: true},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})
	if err != nil {
		// Clean up the uploaded file if database operation fails
		os.Remove(outputPath)
		slog.Error("Failed to add video to the database", "err", err)
		http.Error(w, "Failed to add video to the library", http.StatusInternalServerError)
		return
	}

	// Update storage usage in payment service
	slog.Info("Updating storage usage", "userID", userID, "fileSize", fileSizeForDB)

	_, err = api.paymentServiceClient.UpdateUserUsage(r.Context(), &paymentProto.UpdateUserUsageRequest{
		UserId:      userID,
		UsageType:   "storage",
		UsageChange: fileSizeForDB,
	})
	if err != nil {
		slog.Error("Failed to update storage usage in payment service", "err", err, "userID", userID, "fileSize", fileSizeForDB)
		// Don't fail the upload, just log the error
	} else {
		slog.Info("Storage usage updated successfully", "userID", userID, "fileSize", fileSizeForDB)
	}

	// Success! Respond and exit
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

	// Authentication is handled by the cookie middleware
	authContext, err := interceptors.AuthFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get tenant ID from query parameter (since HTML5 video elements can include query params)
	tenantID := r.URL.Query().Get("tenant")
	if tenantID == "" {
		http.Error(w, "tenant query parameter is required", http.StatusBadRequest)
		return
	}

	// Validate user has access to this tenant
	err = isUserInTenant(r.Context(), api.userServiceClient, api.log, tenantID, authContext.User.ID)
	if err != nil {
		http.Error(w, "Access denied: you are not a member of this tenant", http.StatusForbidden)
		return
	}

	// Get video details from database with tenant validation
	video, err := api.dbQueries.GetVideoByVideoIDAndTenantID(r.Context(), db.GetVideoByVideoIDAndTenantIDParams{
		ID:       videoID,
		TenantID: sql.NullString{String: tenantID, Valid: true},
	})
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
