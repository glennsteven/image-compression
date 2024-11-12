package storage

import (
	"context"
	"fmt"
)

var ErrNotFound = fmt.Errorf("storage object not found")

// Storage defines the minimum interface for a blob storage system.
type Storage interface {
	// Put creates or overwrites an object in the storage system.
	// If contentType is blank, the default for the chosen storage implementation is used.
	Put(ctx context.Context, bucket, name string, contents []byte, contentType string) error

	// Get fetches the object's contents.
	Get(ctx context.Context, bucket, name string) ([]byte, error)
}
