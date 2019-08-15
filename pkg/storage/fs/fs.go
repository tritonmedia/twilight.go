package fs

import (
	"io"
	"os"
	"path/filepath"

	"github.com/tritonmedia/twilight.go/pkg/storage"
)

// Provider implements a storage provider using the filesystem
type Provider struct {
	base string
}

// NewProvider returns a new fs provider
func NewProvider(basePath string) *Provider {
	if !filepath.IsAbs(basePath) {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		basePath = filepath.Join(wd, basePath)
	}

	return &Provider{
		base: basePath,
	}
}

// Exists checks to see if a file exists on the local file system
func (p *Provider) Exists(path string) bool {
	_, err := os.Stat(filepath.Join(p.base, path))
	if err != nil {
		return false
	}

	return true
}

// Unlink deletes a file on the local fs
func (p *Provider) Unlink(path string) error {
	return os.Remove(filepath.Join(p.base, path))
}

// Create a new file on the local filesystem
func (p *Provider) Create(r io.Reader, path string) error {
	if p.Exists(path) {
		return storage.ErrorIsExists
	}

	baseDir := filepath.Dir(filepath.Join(p.base, path))
	if err := os.MkdirAll(baseDir, 0777); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(p.base, path))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}
