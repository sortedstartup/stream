package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const testUploadDir = "test_uploads"

func TestMain(m *testing.M) {
	// Ensure test uploads directory is cleaned up before and after tests
	os.RemoveAll(testUploadDir)
	os.Mkdir(testUploadDir, 0755)
	code := m.Run()
	os.RemoveAll(testUploadDir)
	os.Exit(code)
}

func TestUploadHandlerContentLengthExceedsLimit(t *testing.T) {
	reqBody := bytes.NewBuffer(make([]byte, maxUploadSize+1))
	req := httptest.NewRequest(http.MethodPost, "/upload", reqBody)
	req.Header.Set("Content-Length", "104857601")

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status code %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}
}

func TestUploadHandlerNoContentLengthHeader(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("video", "test.webm")
	part.Write([]byte("dummy content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// Check if the file exists in the uploads directory
	absUploadDir, _ := filepath.Abs(uploadDir)
	expectedPath := filepath.Join(absUploadDir, "test.webm")
	if _, err := os.Stat(expectedPath); err == nil {
		t.Errorf("File should not exist at %s, but it does", expectedPath)
	}
}

func TestUploadHandlerMaxBytesReader(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a large file part exceeding the limit
	part, _ := writer.CreateFormFile("video", "largefile.webm")
	io.CopyN(part, bytes.NewReader(make([]byte, maxUploadSize+1)), maxUploadSize+1)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status code %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}

	// Check that the file is not created
	absUploadDir, _ := filepath.Abs(uploadDir)
	expectedPath := filepath.Join(absUploadDir, "largefile.webm")
	if _, err := os.Stat(expectedPath); err == nil {
		t.Errorf("File should not exist at %s, but it does", expectedPath)
	}
}

func TestUploadHandlerSuccessfulUpload(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("video", "success.webm")
	part.Write([]byte("valid content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	// Check if the file was saved correctly
	absUploadDir, _ := filepath.Abs(uploadDir)
	files, err := os.ReadDir(absUploadDir)
	if err != nil || len(files) == 0 {
		t.Errorf("Expected file in uploads directory, but none found")
	}
}
