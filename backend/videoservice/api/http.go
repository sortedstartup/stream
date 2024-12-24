package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid" // Use UUID for unique IDs
	"sortedstartup.com/stream/videoservice/db"
)

// Helper function to check if a string is in a list of strings
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// uploadHandler handles the file upload and saves video metadata to the database
func uploadHandler(w http.ResponseWriter, r *http.Request, dbQueries *db.Queries) {
	// Parse the form data
	err := r.ParseMultipartForm(10 << 20) // Max upload size: 10MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve file
	file, fileHeader, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))              // Normalize extension to lowercase
	allowedExtensions := []string{".mp4", ".avi", ".mov", ".webm", ".mkv"} // Add common formats
	if !contains(allowedExtensions, ext) {
		http.Error(w, "Invalid file type. Allowed types are: .mp4, .avi, .mov, .webm, .mkv", http.StatusBadRequest)
		return
	}

	// Generate a unique file name
	fileName := uuid.New().String() + ext

	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		http.Error(w, "Failed to create uploads directory", http.StatusInternalServerError)
		return
	}

	// Save the file
	savePath := "./uploads/" + fileName
	out, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	// Prepare the CreateVideoParams for the database
	createVideoParams := db.CreateVideoParams{
		ID:          uuid.New().String(),                          // Unique ID for the video
		Title:       strings.TrimSuffix(fileHeader.Filename, ext), // Video title (filename without extension)
		Description: r.FormValue("description"),                   // Video description from form
		Url:         "/uploads/" + fileName,                       // URL to the uploaded video
	}

	// Save metadata to database
	err = dbQueries.CreateVideo(context.Background(), createVideoParams)
	if err != nil {
		http.Error(w, "Failed to save metadata", http.StatusInternalServerError)
		return
	}

	// Respond with success
	response := map[string]interface{}{
		"message":  "File uploaded successfully",
		"filename": fileName,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// listVideosHandler handles the request to list all videos
func listVideosHandler(w http.ResponseWriter, r *http.Request, dbQueries *db.Queries) {
	// Fetch videos from the database
	videos, err := dbQueries.GetAllVideo(context.Background())
	if err != nil {
		http.Error(w, "Failed to fetch videos", http.StatusInternalServerError)
		return
	}

	// Convert videos to a suitable response structure
	videoList := make([]map[string]interface{}, len(videos))
	for i, video := range videos {
		videoList[i] = map[string]interface{}{
			"id":          video.ID,
			"title":       video.Title,
			"description": video.Description,
			"url":         video.Url,
			"created_at":  video.CreatedAt,
		}
	}

	// Respond with the list of videos
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"videos": videoList,
	})
}

// Serve video files with the correct MIME type and handle CORS
func serveVideo(w http.ResponseWriter, r *http.Request) {
	// Allow cross-origin requests
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins, or specify a particular domain

	// Get the file path from the URL
	filePath := "./uploads" + r.URL.Path[len("/uploads/"):]
	ext := strings.ToLower(filepath.Ext(filePath))

	log.Println("Serving video from path:", filePath)

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		log.Println("File not found:", filePath)
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	// Determine MIME type based on file extension
	var contentType string
	switch ext {
	case ".mp4":
		contentType = "video/mp4"
	case ".avi":
		contentType = "video/x-msvideo"
	case ".mov":
		contentType = "video/quicktime"
	case ".webm":
		contentType = "video/webm"
	case ".mkv":
		contentType = "video/x-matroska"
	default:
		contentType = "application/octet-stream"
	}

	// Set the appropriate Content-Type header
	w.Header().Set("Content-Type", contentType)

	// Serve the file
	http.ServeFile(w, r, filePath)
}
