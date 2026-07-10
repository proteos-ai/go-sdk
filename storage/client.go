// Package storage provides the files service backed by the storage-service.
package storage

import sdk "go.proteos.ai/sdk"

// Client groups the storage services. Construct with New, then access the
// resource services via the public fields:
//
//	s := storage.New(c)
//	file, err := s.Files.Get(ctx, "f-1")
type Client struct {
	Files *FileService
}

// New builds a Client backed by the given *sdk.Client.
func New(c *sdk.Client) *Client {
	return &Client{
		Files: &FileService{c: c},
	}
}
