package s3

import (
	"github.com/minio/minio-go"
	"github.com/pkg/errors"
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

	// TODO(jaredallard): feel hacky
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
