package storage

import (
	"io"
	"time"
)

// ListFilesOptions filters and paginates GET /storage/v1/files.
type ListFilesOptions struct {
	Id            string `query:"id,omitempty"             json:"id,omitempty"`
	Name          string `query:"name,omitempty"           json:"name,omitempty"`
	Page          int    `query:"page"                     json:"page"`
	PageSize      int    `query:"page_size"                json:"page_size"`
	SortDirection string `query:"sort_direction,omitempty" json:"sort_direction,omitempty"`
	SortBy        string `query:"sort_by,omitempty"        json:"sort_by,omitempty"`
}

// FileUpload carries the byte stream for Create. ContentType defaults to
// application/octet-stream on the wire when empty.
type FileUpload struct {
	Filename    string
	ContentType string
	Reader      io.Reader
}

// AccessUrl is a minted short-lived, one-time download or upload URL. The
// URL embeds the one-time token — it is the sole credential at redeem time,
// so treat the value like a secret until it is spent or expired.
type AccessUrl struct {
	Url       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}
