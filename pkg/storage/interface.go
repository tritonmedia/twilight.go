package storage

// Provider is a storage provider for storing media
type Provider interface {
	// Exists checks to see if a path exists
	Exists(path string) bool

	// BucketExists checks to see if a bucket exists
	BucketExists(name string) bool

	// Unlink removes a file path
	Unlink(path string) error

	// Create a file on the remote store
	Create(contents []byte, path string) error
}
