package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestUploadHandlerContentLengthExceedsLimit(t *testing.T) {
	t.Log("Testing Content-Length Exceeds Limit...")
	// Create a request that exceeds the 100MB limit
	req := createFakeFileUploadRequest("file", "largefile.txt", 1024*1024*1024) // 1GB file
	rr := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(rr, req)

	// Check if the response code is 413 (Payload Too Large)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status code %d, got %d", http.StatusRequestEntityTooLarge, rr.Code)
	}

	// Check if the error message is correct
	expectedBody := "File size exceeds the 100 MB limit\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rr.Body.String())
	}
}

func TestUploadHandlerNoContentLengthHeader(t *testing.T) {
	t.Log("Testing No Content-Length Header...")

	// Create a request without the Content-Length header
	req := createFakeFileUploadRequest("file", "test.mp4", 100*1024*1024) // 100MB file
	rr := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(rr, req)

	// Check if the response code is 413 (Payload Too Large) or 200 OK
	// If the file exceeds the limit, the server should reject with 413
	if rr.Code != http.StatusRequestEntityTooLarge && rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200 or 413, got %d", rr.Code)
	}

	// If it's 200 OK, check if the file exists
	if rr.Code == http.StatusOK {
		if _, err := os.Stat("uploads/test.mp4"); os.IsNotExist(err) {
			t.Errorf("Expected file to be saved at uploads/test.mp4, but it does not exist")
		}
	}
}

func TestUploadHandlerMaxBytesReader(t *testing.T) {
	t.Log("Testing MaxBytesReader...")

	// Create a request with a file smaller than the limit
	req := createFakeFileUploadRequest("file", "test.mp4", 50*1024*1024) // 50MB file
	rr := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(rr, req)

	// Check if the response code is 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	// Check if the file exists
	if _, err := os.Stat("uploads/test.mp4"); os.IsNotExist(err) {
		t.Errorf("Expected file to be saved at uploads/test.mp4, but it does not exist")
	}
}

// Helper function to create a fake file upload request
func createFakeFileUploadRequest(fieldName, fileName string, fileSize int) *http.Request {
	// Create a dummy file with the specified size
	var buf bytes.Buffer
	for i := 0; i < fileSize; i++ {
		buf.WriteByte('a') // Write dummy data to the file
	}

	// Create a multipart writer and add the file field
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile(fieldName, fileName)
	part.Write(buf.Bytes()) // Add the fake file content

	// Close the writer and prepare the request
	writer.Close()
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}
