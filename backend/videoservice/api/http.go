package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	uploadDir     = "uploads"
	maxUploadSize = 100 << 20 // Increased limit to 100 MB
)

// uploadHandler handles file uploads
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Enforce Content-Length header if provided
	if r.ContentLength > maxUploadSize {
		http.Error(w, "File size exceeds the 100 MB limit", http.StatusRequestEntityTooLarge)
		return
	}

	// Limit the request body size for memory efficiency
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse the multipart form
	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		fmt.Println("Error parsing form:", err)
		return
	}

	// Retrieve the uploaded file
	file, fileHeader, err := r.FormFile("video") 
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		fmt.Println("Error retrieving file:", err)
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".mp4" && ext != ".mov" && ext != ".avi" {
		http.Error(w, "Unsupported file format. Only .mp4, .mov, .avi are allowed", http.StatusBadRequest)
		return
	}

	// Generate a unique filename with the original extension
	fileName := uuid.New().String() + ext

	// Resolve absolute path for the uploads directory
	absUploadDir, err := filepath.Abs(uploadDir)
	if err != nil {
		http.Error(w, "Failed to resolve upload directory", http.StatusInternalServerError)
		fmt.Println("Error resolving upload directory:", err)
		return
	}
	outputPath := filepath.Join(absUploadDir, fileName)

	// Ensure the uploads directory exists
	if _, err := os.Stat(absUploadDir); os.IsNotExist(err) {
		err := os.Mkdir(absUploadDir, 0755)
		if err != nil {
			http.Error(w, "Error creating uploads directory", http.StatusInternalServerError)
			fmt.Println("Error creating uploads directory:", err)
			return
		}
	}

	// Create the destination file for writing
	outFile, err := os.Create(outputPath)
	if err != nil {
		http.Error(w, "Unable to create file", http.StatusInternalServerError)
		fmt.Println("Error creating file:", err)
		return
	}
	defer outFile.Close()

	// Stream the file content directly to disk
	_, err = io.Copy(outFile, file)
	if err != nil {
		// Check for MaxBytesError
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "File size exceeds the 100 MB limit", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
		}

		// Delete the partially written file
		os.Remove(outputPath)
		fmt.Println("Error saving file, partial file deleted:", err)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "File uploaded successfully", "filename": "%s"}`, fileName)))
}
