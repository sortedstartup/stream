package api

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"context"

	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/videoservice/config"
)

// Helper function to create a test VideoAPI instance
func createTestVideoAPI() *VideoAPI {
	cfg := config.VideoServiceConfig{
		FileStoreDir: "./test_uploads",
	}

	api := &VideoAPI{
		config: cfg,
	}

	return api
}

// Helper function to create authenticated context
func createAuthContext() context.Context {
	authUser := &auth.AuthContext{
		User: &auth.User{
			ID:    "test-user-id",
			Name:  "Test User",
			Email: "test@example.com",
		},
	}
	ctx := context.Background()
	return context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, authUser)
}

func TestUploadHandlerContentLengthExceedsLimit(t *testing.T) {
	t.Log("Start TestUploadHandlerContentLengthExceedsLimit:", time.Now())

	api := createTestVideoAPI()
	reqBody := bytes.NewReader(make([]byte, maxUploadSize+1)) // maxUploadSize + 1
	req := httptest.NewRequest(http.MethodPost, "/upload", reqBody)
	req.Header.Set("Content-Length", "524288001") // maxUploadSize + 1 (500MB + 1)
	req = req.WithContext(createAuthContext())
	rec := httptest.NewRecorder()

	api.uploadHandler(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status code %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}
	expectedBody := "File size exceeds the 500 MB limit\n"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rec.Body.String())
	}
	t.Log("End TestUploadHandlerContentLengthExceedsLimit:", time.Now())
}

func TestUploadHandlerMaxBytesReader(t *testing.T) {
	t.Log("Start TestUploadHandlerMaxBytesReader:", time.Now())

	api := createTestVideoAPI()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("video", "largefile.webm")
	io.CopyN(part, bytes.NewReader(make([]byte, maxUploadSize+1)), maxUploadSize+1)

	// Add required form fields
	writer.WriteField("title", "Test Title")
	writer.WriteField("description", "Test Description")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = req.WithContext(createAuthContext())

	rec := httptest.NewRecorder()
	api.uploadHandler(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status code %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}

	// Verify the error message contains the new 500 MB limit
	if !bytes.Contains(rec.Body.Bytes(), []byte("File size exceeds the 500 MB limit")) {
		t.Errorf("Expected error message to mention 500 MB limit, got: %s", rec.Body.String())
	}

	// Clean up test directory
	os.RemoveAll("./test_uploads")
	t.Log("End TestUploadHandlerMaxBytesReader:", time.Now())
}
