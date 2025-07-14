package api

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"context"

	"github.com/golang/mock/gomock"
	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/videoservice/config"
	"sortedstartup.com/stream/videoservice/db"
	"sortedstartup.com/stream/videoservice/db/mocks"
)

// Helper function to create a test VideoAPI instance
func createTestVideoAPI() *VideoAPI {
	cfg := config.VideoServiceConfig{
		FileStoreDir: "./test_uploads",
	}

	// Add a dummy slog.Logger that discards all logs in tests
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	
	api := &VideoAPI{
		config: cfg,
		log:    logger,
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

	// Mock the global IsUserInTenantFunc to always allow
	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
		return nil
	}

	reqBody := bytes.NewReader(make([]byte, maxUploadSize+1)) // maxUploadSize + 1
	req := httptest.NewRequest(http.MethodPost, "/upload", reqBody)
	req.Header.Set("Content-Length", "524288001") // maxUploadSize + 1 (500MB + 1)
	req.Header.Set("x-tenant-id", "test-tenant-id")
	req = req.WithContext(createAuthContext())
	rec := httptest.NewRecorder()

	api.uploadHandler(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status code %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}
	expectedBody := "File size exceeds the 1024 MB limit\n"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rec.Body.String())
	}
	t.Log("End TestUploadHandlerContentLengthExceedsLimit:", time.Now())
}

func TestUploadHandlerMaxBytesReader(t *testing.T) {
	t.Log("Start TestUploadHandlerMaxBytesReader:", time.Now())

	api := createTestVideoAPI()

	// Mock IsUserInTenantFunc to always allow access
	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
		return nil
	}

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
	req.Header.Set("x-tenant-id", "test-tenant-id")
	req = req.WithContext(createAuthContext())

	rec := httptest.NewRecorder()
	api.uploadHandler(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status code %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}

	if !bytes.Contains(rec.Body.Bytes(), []byte("File size exceeds the 1024 MB limit")) {
		t.Errorf("Expected error message to mention 1024 MB limit, got: %s", rec.Body.String())
	}

	// Clean up test directory
	os.RemoveAll("./test_uploads")
	t.Log("End TestUploadHandlerMaxBytesReader:", time.Now())
}

// helper: create a test API with mock dbQueries
func createTestAPIWithMockDB(t *testing.T) (*VideoAPI, *mocks.MockDBQuerier, func()) {
	ctrl := gomock.NewController(t)
	mockDB := mocks.NewMockDBQuerier(ctrl)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	cfg := config.VideoServiceConfig{
		FileStoreDir: "./test_uploads",
	}
	api := &VideoAPI{
		config:    cfg,
		dbQueries: mockDB,
		log:       logger,
	}
	return api, mockDB, ctrl.Finish
}

// helper: authenticated context with fixed userID
func authCtx() context.Context {
	authUser := &auth.AuthContext{
		User: &auth.User{
			ID: "test-user-id",
		},
	}
	return context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, authUser)
}

// helper: prepare a multipart body with fields title, description, and video file
func prepareMultipartBody(t *testing.T, title, description, fileName string, fileContent []byte) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err := writer.WriteField("title", title)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.WriteField("description", description)
	if err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile("video", fileName)
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write(fileContent)
	if err != nil {
		t.Fatal(err)
	}
	writer.Close()
	return body, writer.FormDataContentType()
}

func TestUploadHandler_Unauthorized(t *testing.T) {
	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	rec := httptest.NewRecorder()

	// No auth context added
	api.uploadHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", rec.Code)
	}
}

func TestUploadHandler_MissingTenantHeader(t *testing.T) {
	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req = req.WithContext(authCtx()) // auth context present
	rec := httptest.NewRecorder()

	// No x-tenant-id header set
	api.uploadHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestUploadHandler_UserNotInTenant(t *testing.T) {
	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req = req.WithContext(authCtx())
	req.Header.Set("x-tenant-id", "tenant-1")
	rec := httptest.NewRecorder()

	// Override isUserInTenant to return error (simulate user not in tenant)
	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
        return errors.New("not a member")
    }

	api.uploadHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden, got %d", rec.Code)
	}
}

func TestUploadHandler_MethodNotAllowed(t *testing.T) {
	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/upload", nil) // GET not allowed
	req = req.WithContext(authCtx())
	rec := httptest.NewRecorder()

	api.uploadHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 Method Not Allowed, got %d", rec.Code)
	}
}

func TestUploadHandler_MissingTitlePart(t *testing.T) {
	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
		return nil
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Intentionally skip title field
	writer.WriteField("description", "desc")
	part, _ := writer.CreateFormFile("video", "test.mp4")
	part.Write([]byte("dummy"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("x-tenant-id", "tenant-1")
	req = req.WithContext(authCtx())

	rec := httptest.NewRecorder()
	api.uploadHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request due to missing title part, got %d", rec.Code)
	}
}

func TestUploadHandler_EmptyTitle(t *testing.T) {
	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
		return nil
	}

	body, contentType := prepareMultipartBody(t, "   ", "desc", "test.mp4", []byte("dummy"))
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-tenant-id", "tenant-1")
	req = req.WithContext(authCtx())

	rec := httptest.NewRecorder()
	api.uploadHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request due to empty title, got %d", rec.Code)
	}
}

func TestUploadHandler_UnsupportedFileExtension(t *testing.T) {
	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
		return nil
	}

	body, contentType := prepareMultipartBody(t, "title", "desc", "test.exe", []byte("dummy"))
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-tenant-id", "tenant-1")
	req = req.WithContext(authCtx())

	rec := httptest.NewRecorder()
	api.uploadHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request due to unsupported extension, got %d", rec.Code)
	}
}

func TestUploadHandler_FileCreationFailure(t *testing.T) {
	api, mockDB, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	// Simulate isUserInTenant success
	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
    return nil 
	}

	// Setup mockDB CreateVideoUploaded to succeed
	mockDB.EXPECT().
		CreateVideoUploaded(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(0)

	// Create upload directory path but simulate Mkdir failure by setting directory to an invalid path
	api.config.FileStoreDir = "/root/invalid/dir" // permission denied

	body, contentType := prepareMultipartBody(t, "title", "desc", "test.mp4", []byte("dummy"))
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-tenant-id", "tenant-1")
	req = req.WithContext(authCtx())
	rec := httptest.NewRecorder()

	api.uploadHandler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 Internal Server Error due to directory creation failure, got %d", rec.Code)
	}
}

func TestUploadHandler_DBError(t *testing.T) {
	api, mockDB, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
    return nil 
	}

	// Setup mockDB CreateVideoUploaded to return error
	mockDB.EXPECT().
		CreateVideoUploaded(gomock.Any(), gomock.Any()).
		Return(errors.New("db failure")).
		Times(1)

	body, contentType := prepareMultipartBody(t, "title", "desc", "test.mp4", []byte("dummy"))
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-tenant-id", "tenant-1")
	req = req.WithContext(authCtx())
	rec := httptest.NewRecorder()

	api.uploadHandler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 Internal Server Error due to DB failure, got %d", rec.Code)
	}

	// Clean up created test directory if any
	os.RemoveAll("./test_uploads")
}

func TestServeVideoHandler_MethodNotAllowed(t *testing.T) {
	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodPost, "/video/someid", nil)
	rec := httptest.NewRecorder()

	api.serveVideoHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", rec.Code)
	}
}

func TestServeVideoHandler_MissingVideoID(t *testing.T) {
	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/video/", nil)
	req = req.WithContext(authCtx())
	rec := httptest.NewRecorder()

	api.serveVideoHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestServeVideoHandler_Unauthorized(t *testing.T) {
	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/video/someid", nil)
	rec := httptest.NewRecorder()

	// No auth context
	api.serveVideoHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", rec.Code)
	}
}

func TestServeVideoHandler_MissingTenantQuery(t *testing.T) {
	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/video/someid", nil)
	req = req.WithContext(authCtx())
	rec := httptest.NewRecorder()

	// No ?tenant param
	api.serveVideoHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestServeVideoHandler_UserNotInTenant(t *testing.T) {
	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	// Simulate IsUserInTenantFunc failing
	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
		return errors.New("not in tenant")
	}

	req := httptest.NewRequest(http.MethodGet, "/video/someid?tenant=test-tenant", nil)
	req = req.WithContext(authCtx())
	rec := httptest.NewRecorder()

	api.serveVideoHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", rec.Code)
	}
}

func TestServeVideoHandler_VideoNotFound(t *testing.T) {
	api, mockDB, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
		return nil
	}

	mockDB.EXPECT().
		GetVideoByVideoIDAndTenantID(gomock.Any(), gomock.Any()).
		Return(db.Video{}, sql.ErrNoRows).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/video/someid?tenant=test-tenant", nil)
	req = req.WithContext(authCtx())
	rec := httptest.NewRecorder()

	api.serveVideoHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestServeVideoHandler_DBError(t *testing.T) {
	api, mockDB, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
		return nil
	}

	mockDB.EXPECT().
		GetVideoByVideoIDAndTenantID(gomock.Any(), gomock.Any()).
		Return(db.Video{}, errors.New("db error")).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/video/someid?tenant=test-tenant", nil)
	req = req.WithContext(authCtx())
	rec := httptest.NewRecorder()

	api.serveVideoHandler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", rec.Code)
	}
}

func TestServeVideoHandler_FileNotFound(t *testing.T) {
	api, mockDB, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
		return nil
	}

	mockDB.EXPECT().
		GetVideoByVideoIDAndTenantID(gomock.Any(), gomock.Any()).
		Return(db.Video{Url: "nonexistent.mp4"}, nil).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/video/someid?tenant=test-tenant", nil)
	req = req.WithContext(authCtx())
	rec := httptest.NewRecorder()

	api.serveVideoHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestServeVideoHandler_Success(t *testing.T) {
	api, mockDB, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	IsUserInTenantFunc = func(ctx context.Context, tenantID, userID string) error {
		return nil
	}

	// Create dummy file to serve
	uploadDir := "./test_uploads"
	os.MkdirAll(uploadDir, 0755)
	defer os.RemoveAll(uploadDir)

	videoFileName := "test-video.mp4"
	filePath := filepath.Join(uploadDir, videoFileName)
	err := os.WriteFile(filePath, []byte("dummy video content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	api.config.FileStoreDir = uploadDir

	mockDB.EXPECT().
		GetVideoByVideoIDAndTenantID(gomock.Any(), gomock.Any()).
		Return(db.Video{Url: videoFileName}, nil).
		Times(1)

	req := httptest.NewRequest(http.MethodGet, "/video/someid?tenant=test-tenant", nil)
	req = req.WithContext(authCtx())
	rec := httptest.NewRecorder()

	api.serveVideoHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Header().Get("Content-Type"), "video/mp4") {
		t.Errorf("Expected video/mp4 Content-Type, got %s", rec.Header().Get("Content-Type"))
	}
}
