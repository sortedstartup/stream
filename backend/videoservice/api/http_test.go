package api

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUploadHandlerContentLengthExceedsLimit(t *testing.T) {
	t.Log("Start TestUploadHandlerContentLengthExceedsLimit:", time.Now())
	reqBody := bytes.NewReader(make([]byte, maxUploadSize+1)) // maxUploadSize + 1
	req := httptest.NewRequest(http.MethodPost, "/upload", reqBody)
	req.Header.Set("Content-Length", "104857601") // maxUploadSize + 1
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status code %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}
	expectedBody := "File size exceeds the 100 MB limit\n"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rec.Body.String())
	}
	t.Log("End TestUploadHandlerContentLengthExceedsLimit:", time.Now())
}

func TestUploadHandlerNoContentLengthHeader(t *testing.T) {
	t.Log("Start TestUploadHandlerNoContentLengthHeader:", time.Now())
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("video", "test.webm")
	part.Write([]byte("testdata")) // small dummy data
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	absUploadDir, _ := filepath.Abs(uploadDir)
	expectedPath := filepath.Join(absUploadDir, "test.webm")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file to be saved at %s, but it does not exist", expectedPath)
	} else {
		os.Remove(expectedPath) // Clean up after the test
	}
	t.Log("End TestUploadHandlerNoContentLengthHeader:", time.Now())
}

func TestUploadHandlerMaxBytesReader(t *testing.T) {
	t.Log("Start TestUploadHandlerMaxBytesReader:", time.Now())
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

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

	absUploadDir, _ := filepath.Abs(uploadDir)
	expectedPath := filepath.Join(absUploadDir, "largefile.webm")
	if _, err := os.Stat(expectedPath); err == nil {
		t.Errorf("File should not exist at %s, but it does", expectedPath)
	}
	t.Log("End TestUploadHandlerMaxBytesReader:", time.Now())
}
