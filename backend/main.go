package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Directory to save uploaded files
const uploadDir = "uploads"

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set file size limit to 100MB
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		fmt.Println("Error parsing form:", err)
		return
	}

	file, fileHeader, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		fmt.Println("Error retrieving file:", err)
		return
	}
	defer file.Close()

	// Get the file extension from the uploaded file
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		http.Error(w, "Invalid file extension", http.StatusBadRequest)
		fmt.Println("Error: file extension missing")
		return
	}

	// Generate a unique filename
	fileName := uuid.New().String() + ext
	outputPath := filepath.Join(uploadDir, fileName)

	// Create the uploads directory if it doesn't exist
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.Mkdir(uploadDir, 0755)
		if err != nil {
			http.Error(w, "Error creating uploads directory", http.StatusInternalServerError)
			fmt.Println("Error creating uploads directory:", err)
			return
		}
	}

	// Save the uploaded file
	outFile, err := os.Create(outputPath)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		fmt.Println("Error saving file:", err)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		fmt.Println("Error writing file:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "File uploaded successfully", "filename": "%s"}`, fileName)))
}

func main() {
	// Set up the /upload route
	http.HandleFunc("/upload", uploadHandler)

	// Start the server
	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
