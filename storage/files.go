package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	storagemodel "go.proteos.ai/model/storage"
	storageapi "go.proteos.ai/model/storage/api"
	sdk "go.proteos.ai/sdk"
	"go.proteos.ai/sdk/internal/httpx"
)

const filesBasePath = "/storage/v1/files"

// FileServiceAPI is the contract a FileService satisfies. Provided so
// callers can inject mocks.
type FileServiceAPI interface {
	List(opts *ListFilesOptions) *sdk.PageIterator[storagemodel.File, ListFilesOptions]
	ListPage(ctx context.Context, opts *ListFilesOptions) (sdk.ListResult[storagemodel.File], error)
	Get(ctx context.Context, id string) (storagemodel.File, error)
	Create(ctx context.Context, request storageapi.CreateFileRequest, upload *FileUpload) (storagemodel.File, error)
	Update(ctx context.Context, id string, request storageapi.UpdateFileRequest, upload *FileUpload) (storagemodel.File, error)
	Download(ctx context.Context, id string) (io.ReadCloser, error)
	Delete(ctx context.Context, id string) error
	DeleteVersion(ctx context.Context, id string, versionId string) error
	GenerateDownloadUrl(ctx context.Context, id string) (AccessUrl, error)
	GenerateUploadUrl(ctx context.Context, id string) (AccessUrl, error)
}

// FileService manages files and their one-time access URLs via the
// storage-service. Single-object responses arrive in a {data: ...} envelope
// and are unwrapped before returning.
type FileService struct{ c *sdk.Client }

var _ FileServiceAPI = (*FileService)(nil)

// dataEnvelope is the storage-service single-object response wrapper.
type dataEnvelope[T any] struct {
	Data T `json:"data"`
}

// List returns a PageIterator over files matching opts.
func (s *FileService) List(opts *ListFilesOptions) *sdk.PageIterator[storagemodel.File, ListFilesOptions] {
	o := ListFilesOptions{}
	if opts != nil {
		o = *opts
	}
	if o.PageSize == 0 {
		o.PageSize = sdk.DefaultPageSize
	}
	return sdk.NewPageIterator(func(ctx context.Context, page int, in ListFilesOptions) (sdk.ListResult[storagemodel.File], error) {
		in.Page = page
		return s.ListPage(ctx, &in)
	}, o)
}

// ListPage fetches a single page of files.
func (s *FileService) ListPage(ctx context.Context, opts *ListFilesOptions) (sdk.ListResult[storagemodel.File], error) {
	var out sdk.ListResult[storagemodel.File]
	err := s.c.DoWithQuery(ctx, http.MethodGet, filesBasePath, opts, nil, &out)
	return out, err
}

// Get returns the file with the given id.
func (s *FileService) Get(ctx context.Context, id string) (storagemodel.File, error) {
	var out dataEnvelope[storagemodel.File]
	err := s.c.Do(ctx, http.MethodGet, filesBasePath+"/"+id, nil, &out)
	return out.Data, err
}

// Create creates a file. With upload nil only the metadata row is created
// (is_persisted=false) — the two-step flow for minting an upload URL
// afterwards. With upload set, the bytes are streamed in the same request
// and the file comes back persisted with its first version.
func (s *FileService) Create(ctx context.Context, request storageapi.CreateFileRequest, upload *FileUpload) (storagemodel.File, error) {
	metaJSON, err := json.Marshal(request)
	if err != nil {
		return storagemodel.File{}, fmt.Errorf("storage: marshal create file metadata: %w", err)
	}

	var files []httpx.MultipartFile
	if upload != nil && upload.Reader != nil {
		contentType := upload.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		files = append(files, httpx.MultipartFile{
			FieldName:   "file",
			Filename:    upload.Filename,
			ContentType: contentType,
			Reader:      upload.Reader,
		})
	}

	var out dataEnvelope[storagemodel.File]
	err = s.c.DoMultipartJSON(ctx, http.MethodPost, filesBasePath, "metadata", string(metaJSON), files, &out)
	return out.Data, err
}

// Update patches a file's metadata — name, is_locked, public_access (only
// ["read"] is honored; an empty set makes the file private again). Nil fields
// are left untouched. With upload set, the bytes are streamed in the same
// request as a new version.
func (s *FileService) Update(ctx context.Context, id string, request storageapi.UpdateFileRequest, upload *FileUpload) (storagemodel.File, error) {
	metaJSON, err := json.Marshal(request)
	if err != nil {
		return storagemodel.File{}, fmt.Errorf("storage: marshal update file metadata: %w", err)
	}

	var files []httpx.MultipartFile
	if upload != nil && upload.Reader != nil {
		contentType := upload.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		files = append(files, httpx.MultipartFile{
			FieldName:   "file",
			Filename:    upload.Filename,
			ContentType: contentType,
			Reader:      upload.Reader,
		})
	}

	var out dataEnvelope[storagemodel.File]
	err = s.c.DoMultipartJSON(ctx, http.MethodPatch, filesBasePath+"/"+id, "metadata", string(metaJSON), files, &out)
	return out.Data, err
}

// Download streams the file's current-version bytes through the
// authenticated endpoint. The caller must Close the returned reader.
func (s *FileService) Download(ctx context.Context, id string) (io.ReadCloser, error) {
	body, _, err := s.c.DoRaw(ctx, http.MethodGet, filesBasePath+"/"+id+"/download", nil)
	return body, err
}

// Delete permanently deletes a file: every version's stored bytes, its
// access tokens and parsed content. Irreversible. Locked files are refused
// (409 delete_locked_file); unlock via Update first.
func (s *FileService) Delete(ctx context.Context, id string) error {
	return s.c.Do(ctx, http.MethodDelete, filesBasePath+"/"+id, nil, nil)
}

// DeleteVersion permanently deletes a single non-current version and its
// stored bytes. Deleting the current version is refused
// (409 delete_current_version), as are locked files (409 delete_locked_file).
func (s *FileService) DeleteVersion(ctx context.Context, id string, versionId string) error {
	return s.c.Do(ctx, http.MethodDelete, filesBasePath+"/"+id+"/versions/"+versionId, nil, nil)
}

// GenerateDownloadUrl mints a short-lived (5 min TTL), one-time-use public
// download URL for the file's current version. The file must be persisted.
func (s *FileService) GenerateDownloadUrl(ctx context.Context, id string) (AccessUrl, error) {
	var out dataEnvelope[AccessUrl]
	err := s.c.Do(ctx, http.MethodPost, filesBasePath+"/"+id+"/generate-download-url", nil, &out)
	return out.Data, err
}

// GenerateUploadUrl mints a short-lived (3 min TTL), one-time-use public
// upload URL for the file. Create the file first (metadata-only Create);
// locked files are rejected.
func (s *FileService) GenerateUploadUrl(ctx context.Context, id string) (AccessUrl, error) {
	var out dataEnvelope[AccessUrl]
	err := s.c.Do(ctx, http.MethodPost, filesBasePath+"/"+id+"/generate-upload-url", nil, &out)
	return out.Data, err
}
