package s3

import (
	"io"

	"github.com/minio/minio-go"
	"github.com/pkg/errors"
	"github.com/tritonmedia/twilight.go/pkg/storage"
)

// Client is a s3 client
type Client struct {
	m      *minio.Client
	bucket string
}

// NewClient retruns a new S3 client
func NewClient(accessKey, secretKey, endpoint, bucket string) (*Client, error) {
	m, err := minio.New(
		endpoint,
		accessKey,
		secretKey,

		// TODO(jaredallard): calculate from endpoint
		false,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wrap minio")
	}

	if _, err := m.ListBuckets(); err != nil {
		return nil, errors.Wrap(err, "failed to test s3 connection")
	}

	// TODO(jaredallard): feels hacky
	m.MakeBucket(bucket, "us-west-2")

	return &Client{
		m:      m,
		bucket: bucket,
	}, nil
}

// Exists checks if a key exists
func (c *Client) Exists(path string) bool {
	if _, err := c.m.StatObject(c.bucket, path, minio.StatObjectOptions{}); err != nil {
		return false
	}

	return true
}

// Unlink removes an object from the bucket
func (c *Client) Unlink(path string) error {
	return c.m.RemoveObject(c.bucket, path)
}

// Create uploads an object to an s3 compatible storage
func (c *Client) Create(r io.Reader, destPath string) error {
	exists := c.Exists(destPath)
	if exists {
		return storage.ErrorIsExists
	}

	_, err := c.m.PutObject(c.bucket, destPath, r, -1, minio.PutObjectOptions{})
	return err
}
