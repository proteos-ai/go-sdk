package storage_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	storagemodel "go.proteos.ai/model/storage"
	storageapi "go.proteos.ai/model/storage/api"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/storage"
)

func newClient(t *testing.T, handler http.HandlerFunc) *storage.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := sdk.NewClient(sdk.WithBaseURL(srv.URL), sdk.WithToken("t"))
	require.NoError(t, err)
	return storage.New(c)
}

func TestFilesGet_UnwrapsDataEnvelope(t *testing.T) {
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/storage/v1/files/f-1", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": storagemodel.File{Id: "f-1", Name: "report.pdf", IsPersisted: true},
		})
	})

	file, err := s.Files.Get(context.Background(), "f-1")
	require.NoError(t, err)
	require.Equal(t, "f-1", file.Id)
	require.Equal(t, "report.pdf", file.Name)
	require.True(t, file.IsPersisted)
}

func TestFilesListPage(t *testing.T) {
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/storage/v1/files", r.URL.Path)
		require.Equal(t, "report", r.URL.Query().Get("name"))
		require.Equal(t, "25", r.URL.Query().Get("page_size"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]any{"items_total": 1, "pages_total": 1},
			"data": []storagemodel.File{{Id: "f-1", Name: "report.pdf"}},
		})
	})

	page, err := s.Files.ListPage(context.Background(), &storage.ListFilesOptions{Name: "report", PageSize: 25})
	require.NoError(t, err)
	require.Len(t, page.Data, 1)
	require.Equal(t, "f-1", page.Data[0].Id)
}

func TestFilesCreate_MetadataOnly(t *testing.T) {
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/storage/v1/files", r.URL.Path)

		mr, err := r.MultipartReader()
		require.NoError(t, err)

		part, err := mr.NextPart()
		require.NoError(t, err)
		require.Equal(t, "metadata", part.FormName())
		require.Equal(t, "application/json", part.Header.Get("Content-Type"))
		var req storageapi.CreateFileRequest
		require.NoError(t, json.NewDecoder(part).Decode(&req))
		require.Equal(t, "report.pdf", req.Name)

		_, err = mr.NextPart()
		require.Equal(t, io.EOF, err, "metadata-only create must carry no file part")

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": storagemodel.File{Id: "f-1", Name: "report.pdf", IsPersisted: false},
		})
	})

	file, err := s.Files.Create(context.Background(), storageapi.CreateFileRequest{Name: "report.pdf", ContentType: "application/pdf"}, nil)
	require.NoError(t, err)
	require.Equal(t, "f-1", file.Id)
	require.False(t, file.IsPersisted)
}

func TestFilesCreate_WithUpload(t *testing.T) {
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		mr, err := r.MultipartReader()
		require.NoError(t, err)

		part, err := mr.NextPart()
		require.NoError(t, err)
		require.Equal(t, "metadata", part.FormName())

		filePart, err := mr.NextPart()
		require.NoError(t, err)
		require.Equal(t, "file", filePart.FormName())
		require.Equal(t, "text/plain", filePart.Header.Get("Content-Type"))
		body, err := io.ReadAll(filePart)
		require.NoError(t, err)
		require.Equal(t, "hello", string(body))

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": storagemodel.File{Id: "f-1", Name: "hello.txt", IsPersisted: true},
		})
	})

	file, err := s.Files.Create(context.Background(),
		storageapi.CreateFileRequest{Name: "hello.txt", ContentType: "text/plain"},
		&storage.FileUpload{Filename: "hello.txt", ContentType: "text/plain", Reader: strings.NewReader("hello")})
	require.NoError(t, err)
	require.True(t, file.IsPersisted)
}

func TestFilesDownload_StreamsBody(t *testing.T) {
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/storage/v1/files/f-1/download", r.URL.Path)
		_, _ = w.Write([]byte("file bytes"))
	})

	body, err := s.Files.Download(context.Background(), "f-1")
	require.NoError(t, err)
	defer body.Close()
	raw, err := io.ReadAll(body)
	require.NoError(t, err)
	require.Equal(t, "file bytes", string(raw))
}

func TestFilesGenerateDownloadUrl(t *testing.T) {
	expires := time.Date(2026, 6, 10, 12, 5, 0, 0, time.UTC)
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/storage/v1/files/f-1/generate-download-url", r.URL.Path)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": storage.AccessUrl{Url: "https://api.proteos.ai/storage/v1/files/f-1/download/url/tok", ExpiresAt: expires},
		})
	})

	got, err := s.Files.GenerateDownloadUrl(context.Background(), "f-1")
	require.NoError(t, err)
	require.Equal(t, "https://api.proteos.ai/storage/v1/files/f-1/download/url/tok", got.Url)
	require.Equal(t, expires, got.ExpiresAt)
}

func TestFilesGenerateUploadUrl_SurfacesApiError(t *testing.T) {
	s := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/storage/v1/files/f-1/generate-upload-url", r.URL.Path)
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "file_locked", "message": "attempt to mint an upload url for a locked file: f-1"})
	})

	_, err := s.Files.GenerateUploadUrl(context.Background(), "f-1")
	require.Error(t, err)
	var sdkErr *sdk.Error
	require.ErrorAs(t, err, &sdkErr)
	require.Equal(t, http.StatusConflict, sdkErr.HTTPStatus)
	require.Equal(t, sdk.ErrCode("file_locked"), sdkErr.Code)
}
