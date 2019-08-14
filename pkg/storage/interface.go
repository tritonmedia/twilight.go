package storage

import (
	"errors"
	"io"
)

var (
	// ErrorIsExists is thrown when a file already exists on create
	ErrorIsExists = errors.New("file already exists")
)

// Provider is a storage provider for storing media
type Provider interface {
	// Exists checks to see if a path exists
	Exists(path string) bool

	// Unlink removes a file path
	Unlink(path string) error

	// Create a file on the remote store
	Create(r io.Reader, destPath string) error
}
