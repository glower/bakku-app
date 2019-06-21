package storage

import "github.com/glower/bakku-app/pkg/storage/boltdb"

// Storage is an interface for a permanent storage for a files meta data
type Storage interface {
	Exist() bool
	Add(string, string, []byte) error
	Get(string, string) (string, error)
	GetAll(string) (map[string]string, error)
	Remove(string, string) error
}

// New returns new snapshot storage implementation
func New(path string) Storage {
	return &boltdb.Storage{
		DBFilePath: path,
	}
}
