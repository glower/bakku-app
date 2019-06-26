package storage

import (
	"github.com/glower/bakku-app/pkg/storage/boltdb"
)

// Storager is an interface for a permanent storage for a files meta data
type Storager interface {
	Exist() bool
	Add(string, string, []byte) error
	Get(string, string) ([]byte, error)
	GetAll(string) (map[string]string, error)
	Remove(string, string) error
}

// New returns new snapshot storage implementation
func New(path string) Storager {
	return &boltdb.BoltDB{
		DBFilePath: path,
	}
}
