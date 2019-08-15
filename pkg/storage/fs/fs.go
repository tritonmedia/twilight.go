package fs

import (
	"io"
	"os"

	"github.com/tritonmedia/twilight.go/pkg/storage"
)

// Provider implements a storage provider using the filesystem
type Provider struct{}

// NewProvider returns a new fs provider
func NewProvider() *Provider {
	return &Provider{}
}

// Exists checks to see if a file exists on the local file system
func (p *Provider) Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}

	return true
}

// Unlink deletes a file on the local fs
func (p *Provider) Unlink(path string) error {
	return os.Remove(path)
}

// Create a new file on the local filesystem
func (p *Provider) Create(r io.Reader, path string) error {
	if p.Exists(path) {
		return storage.ErrorIsExists
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}
