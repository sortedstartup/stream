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
	"sortedstartup.com/stream/userservice/proto"
	"sortedstartup.com/stream/videoservice/config"
	"sortedstartup.com/stream/videoservice/db"
	"sortedstartup.com/stream/videoservice/db/mocks"
)

// fakeLargeReader streams 'a' bytes for N bytes without large memory allocation
type fakeLargeReader struct {
	N    int64 // total bytes to simulate
	read int64 // bytes read so far
}

func (r *fakeLargeReader) Read(p []byte) (int, error) {
	if r.read >= r.N {
		return 0, io.EOF
	}
	toRead := int64(len(p))
	if r.read+toRead > r.N {
		toRead = r.N - r.read
	}
	for i := int64(0); i < toRead; i++ {
		p[i] = 'a'
	}
	r.read += toRead
	return int(toRead), nil
}

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)

	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{
				{Tenant: &proto.Tenant{Id: "tenant-1"}},
			},
		}, nil).
		AnyTimes()

	api := createTestVideoAPI()

	api.userServiceClient = mockUser

	reqBody := bytes.NewReader(make([]byte, maxUploadSize+1)) // maxUploadSize + 1
	req := httptest.NewRequest(http.MethodPost, "/upload", reqBody)
	req.Header.Set("Content-Length", "524288001") // maxUploadSize + 1 (500MB + 1)
	req.Header.Set("x-tenant-id", "tenant-1")     // match mocked tenant
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

// prepareMultipartBodyStream streams multipart body with large file content on-the-fly
func prepareMultipartBodyStream(t *testing.T, title, description, channelID, fileName string, fileSize int64) (*io.PipeReader, string) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

		if err := writer.WriteField("title", title); err != nil {
			pw.CloseWithError(err)
			return
		}
		if err := writer.WriteField("description", description); err != nil {
			pw.CloseWithError(err)
			return
		}
		if err := writer.WriteField("channel_id", channelID); err != nil {
			pw.CloseWithError(err)
			return
		}

		part, err := writer.CreateFormFile("video", fileName)
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		fakeReader := &fakeLargeReader{N: fileSize}
		if _, err := io.Copy(part, fakeReader); err != nil {
			pw.CloseWithError(err)
			return
		}
	}()

	return pr, writer.FormDataContentType()
}

func TestUploadHandlerMaxBytesReader(t *testing.T) {
	t.Log("Start TestUploadHandlerMaxBytesReader:", time.Now())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)
	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{
				{Tenant: &proto.Tenant{Id: "tenant-1"}},
			},
		}, nil).
		AnyTimes()

	api := createTestVideoAPI()
	api.userServiceClient = mockUser

	// Set file size to maxUploadSize + 1 to exceed limit
	var fileSize int64 = maxUploadSize + 1
	body, contentType := prepareMultipartBodyStream(t, "Test Title", "Test Description", "test-channel", "largefile.mp4", fileSize)

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-tenant-id", "tenant-1")
	req = req.WithContext(createAuthContext())

	rec := httptest.NewRecorder()
	api.uploadHandler(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status code %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}

	// Flexible substring match for error message:
	if !strings.Contains(rec.Body.String(), "File size exceeds") && !strings.Contains(rec.Body.String(), "request body too large") {
		t.Errorf("Expected error message mentioning size limit, got: %s", rec.Body.String())
	}

	// Clean up test uploads dir if any
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
func prepareMultipartBody(t *testing.T, title, description, channelID, fileName string, fileContent []byte) (*bytes.Buffer, string) {
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
	err = writer.WriteField("channel_id", channelID) // <-- added channel_id field
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)
	// Return empty tenant list simulating user NOT in tenant
	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{},
		}, nil).
		Times(1)

	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	api.userServiceClient = mockUser

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req = req.WithContext(authCtx())
	req.Header.Set("x-tenant-id", "tenant-1")
	rec := httptest.NewRecorder()

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)
	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{
				{Tenant: &proto.Tenant{Id: "tenant-1"}},
			},
		}, nil).
		Times(1)

	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()
	api.userServiceClient = mockUser

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)
	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{
				{Tenant: &proto.Tenant{Id: "tenant-1"}},
			},
		}, nil).
		Times(1)

	api, mockDB, teardown := createTestAPIWithMockDB(t) // capture mockDB
	defer teardown()

	api.userServiceClient = mockUser

	// Set expectation on mockDB because handler will call CreateVideoUploaded
	mockDB.EXPECT().
		CreateVideoUploaded(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)

	body, contentType := prepareMultipartBody(t, "   ", "desc", "channel-1", "test.mp4", []byte("dummy"))
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-tenant-id", "tenant-1")
	req = req.WithContext(authCtx())

	rec := httptest.NewRecorder()
	api.uploadHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200 OK because title is autogenerated, got %d", rec.Code)
	}
}

func TestUploadHandler_UnsupportedFileExtension(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)
	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{
				{Tenant: &proto.Tenant{Id: "tenant-1"}},
			},
		}, nil).
		Times(1)

	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	api.userServiceClient = mockUser

	body, contentType := prepareMultipartBody(t, "title", "desc", "channel-1", "test.exe", []byte("dummy"))
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)
	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{
				{Tenant: &proto.Tenant{Id: "tenant-1"}},
			},
		}, nil).
		Times(1)

	api, mockDB, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	api.userServiceClient = mockUser

	// Setup mockDB CreateVideoUploaded to succeed
	mockDB.EXPECT().
		CreateVideoUploaded(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(0)

	// Create upload directory path but simulate Mkdir failure by setting directory to an invalid path
	api.config.FileStoreDir = "/root/invalid/dir" // permission denied

	body, contentType := prepareMultipartBody(t, "title", "desc", "channel-1", "test.mp4", []byte("dummy"))
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)
	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{
				{Tenant: &proto.Tenant{Id: "tenant-1"}},
			},
		}, nil).
		Times(1)
	api, mockDB, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	api.userServiceClient = mockUser

	// Setup mockDB CreateVideoUploaded to return error
	mockDB.EXPECT().
		CreateVideoUploaded(gomock.Any(), gomock.Any()).
		Return(errors.New("db failure")).
		Times(1)

	body, contentType := prepareMultipartBody(t, "title", "desc", "channel-1", "test.mp4", []byte("dummy"))
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)
	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{},
		}, nil).
		Times(1)

	api, _, teardown := createTestAPIWithMockDB(t)
	defer teardown()
	api.userServiceClient = mockUser

	req := httptest.NewRequest(http.MethodGet, "/video/someid?tenant=test-tenant", nil)
	req = req.WithContext(authCtx())
	rec := httptest.NewRecorder()

	api.serveVideoHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden, got %d", rec.Code)
	}
}

func TestServeVideoHandler_VideoNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)
	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{
				{Tenant: &proto.Tenant{Id: "test-tenant"}},
			},
		}, nil).
		Times(1)
	api, mockDB, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	api.userServiceClient = mockUser

	mockDB.EXPECT().
		GetVideoByVideoIDAndTenantID(gomock.Any(), gomock.Any()).
		Return(db.VideoserviceVideo{}, sql.ErrNoRows).
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)
	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{
				{Tenant: &proto.Tenant{Id: "test-tenant"}},
			},
		}, nil).
		Times(1)

	api, mockDB, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	api.userServiceClient = mockUser

	mockDB.EXPECT().
		GetVideoByVideoIDAndTenantID(gomock.Any(), gomock.Any()).
		Return(db.VideoserviceVideo{}, errors.New("db error")).
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)
	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{
				{Tenant: &proto.Tenant{Id: "test-tenant"}},
			},
		}, nil).
		Times(1)

	api, mockDB, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	api.userServiceClient = mockUser

	mockDB.EXPECT().
		GetVideoByVideoIDAndTenantID(gomock.Any(), gomock.Any()).
		Return(db.VideoserviceVideo{Url: "nonexistent.mp4"}, nil).
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUser := proto.NewMockUserServiceClient(ctrl)
	mockUser.EXPECT().
		GetTenants(gomock.Any(), gomock.Any()).
		Return(&proto.GetTenantsResponse{
			TenantUsers: []*proto.TenantUser{
				{Tenant: &proto.Tenant{Id: "test-tenant"}},
			},
		}, nil).
		Times(1)

	api, mockDB, teardown := createTestAPIWithMockDB(t)
	defer teardown()

	api.userServiceClient = mockUser

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
		Return(db.VideoserviceVideo{Url: videoFileName}, nil).
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
