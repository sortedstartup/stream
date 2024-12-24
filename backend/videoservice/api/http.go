package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
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
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExtensions := []string{".mp4", ".avi", ".mov", ".webm", ".mkv"}
	if !contains(allowedExtensions, ext) {
		http.Error(w, "Invalid file type. Allowed types are: .mp4, .avi, .mov, .webm, .mkv", http.StatusBadRequest)
		return
	}

	fileName := uuid.New().String() + ext

	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		http.Error(w, "Failed to create uploads directory", http.StatusInternalServerError)
		return
	}

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

	createVideoParams := db.CreateVideoParams{
		ID:          uuid.New().String(),
		Title:       strings.TrimSuffix(fileHeader.Filename, ext),
		Description: r.FormValue("description"),
		Url:         "/uploads/" + fileName,
	}

	err = dbQueries.CreateVideo(context.Background(), createVideoParams)
	if err != nil {
		http.Error(w, "Failed to save metadata", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":  "File uploaded successfully",
		"filename": fileName,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
