package s3

import (
	"io"
	"net/url"

	"github.com/minio/minio-go/v6"
	"github.com/pkg/errors"
	"github.com/tritonmedia/twilight.go/pkg/storage"
)

// Provider is a s3 comptaible client storage provider
type Provider struct {
	m      *minio.Client
	bucket string
}

// NewProvider returns an s3 comptaible client storage provider
func NewProvider(accessKey, secretKey, endpoint, bucket string) (*Provider, error) {
	var ssl bool
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse minio endpoint")
	}

	if u.Scheme == "https" {
		ssl = true
	}

	m, err := minio.New(
		u.Host,
		accessKey,
		secretKey,
		ssl,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap minio")
	}

	if _, err := m.ListBuckets(); err != nil {
		return nil, errors.Wrap(err, "failed to test s3 connection")
	}

	// TODO(jaredallard): feels hacky
	m.MakeBucket(bucket, "us-west-2")

	return &Provider{
		m:      m,
		bucket: bucket,
	}, nil
}

// Exists checks if a key exists
func (p *Provider) Exists(path string) bool {
	if _, err := p.m.StatObject(p.bucket, path, minio.StatObjectOptions{}); err != nil {
		return false
	}

	return true
}

// Unlink removes an object from the bucket
func (p *Provider) Unlink(path string) error {
	return p.m.RemoveObject(p.bucket, path)
}

// Create uploads an object to an s3 compatible storage
func (p *Provider) Create(r io.Reader, destPath string) error {
	if p.Exists(destPath) {
		return storage.ErrorIsExists
	}

	_, err := p.m.PutObject(p.bucket, destPath, r, -1, minio.PutObjectOptions{})
	return err
}
