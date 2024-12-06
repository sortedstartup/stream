package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request for debugging purpose
	fmt.Println("Upload request received")

	// Parse the form data
	err := r.ParseMultipartForm(10 << 20) // Limit upload to 10MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		fmt.Println("Error parsing form:", err)
		return
	}

	// Retrieve the file from the form
	file, _, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		fmt.Println("Error retrieving file:", err)
		return
	}
	defer file.Close()

	// Create a file to store the uploaded video
	outFile, err := os.Create(filepath.Join("uploads", "recording.webm"))
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		fmt.Println("Error saving file:", err)
		return
	}
	defer outFile.Close()

	// Copy the file data into the server file
	_, err = outFile.ReadFrom(file)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		fmt.Println("Error copying file:", err)
		return
	}

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "File uploaded successfully"}`))
}

func main() {
	// Ensure the uploads directory exists
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		err := os.Mkdir("uploads", 0755)
		if err != nil {
			fmt.Println("Error creating uploads directory:", err)
			return
		}
	}

	http.HandleFunc("/upload", uploadHandler)
	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
