package storage

import (
	"fmt"
	"log"
	"sync"
)

// Storage is an interface for a permanent storage for a files meta data
type Storage interface {
	Path() string
	FilePath() string
	FileName() string

	Add(string, string, []byte) error
	Exist() bool
	Get(string, string) (string, error)
	GetAll(string) (map[string]string, error)
	Remove(string, string) error
}

var (
	snapshotStoragesM sync.RWMutex
	snapshotStorages  = make(map[string]Storage)
)

// Register a snapshot storage implementation by name.
func Register(s Storage) error {

	if s == nil {
		return fmt.Errorf("storage.Register(): could not register a nil Storage interface")
	}

	path := s.Path()
	if path == "" {
		return fmt.Errorf("storage.Register(): could not register a StorageFactory with an empty path")
	}

	snapshotStoragesM.Lock()
	defer snapshotStoragesM.Unlock()

	if _, dup := snapshotStorages[path]; dup {
		return fmt.Errorf("storage.Register(): called twice for [%s]", path)
	}

	log.Printf("storage.Register(): snapshot storage for the path [%s] registered\n", path)
	snapshotStorages[path] = s
	return nil
}

// GetByPath returs a snapshot storage for a given path
func GetByPath(path string) (Storage, error) {
	if len(snapshotStorages) == 0 {
		return nil, fmt.Errorf("snapshot storage is empty")
	}
	snapshotStorage, ok := snapshotStorages[path]
	log.Printf("storage.GetByPath(): path=%s, found=%v\n", path, ok)
	if ok {
		return snapshotStorage, nil
	}
	return nil, fmt.Errorf("cannot find snapshot for a given path: [%s]", path)
}
