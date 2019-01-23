package storage

import (
	"fmt"
	"log"
	"sync"
)

// Storage is an interface for a permanent storage for a files meta data
type Storage interface {
	Add(string, string, []byte) error
	Exist() bool
	FilePath() string
	FileName() string
	Get(string, string) (string, error)
	GetAll(string) (map[string]string, error)
	Remove(string, string) error
}

var (
	snapshotStoragesM sync.RWMutex
	snapshotStorages  = make(map[string]Storage)
)

// Register a snapshot storage implementation by name.
func Register(path string, s Storage) {
	if path == "" {
		panic("storage.Register(): could not register a StorageFactory with an empty path")
	}

	if s == nil {
		panic("storage.Register(): could not register a nil Storage interface")
	}

	snapshotStoragesM.Lock()
	defer snapshotStoragesM.Unlock()

	if _, dup := snapshotStorages[path]; dup {
		log.Printf("[ERROR] storage.Register(): called twice for " + path)
		return
	}

	log.Printf("storage.Register(): snapshot storage for the path [%s] registered\n", path)
	snapshotStorages[path] = s
}

// GetByPath returs a snapshot storage for a given path
func GetByPath(path string) (Storage, error) {
	snapshotStorage, ok := snapshotStorages[path]
	if ok {
		return snapshotStorage, nil
	}
	return nil, fmt.Errorf("cannot find snapshot for a given path: [%s]", path)
}
