package storage

import (
	"log"
	"sync"
)

// Storager ...
type Storager interface {
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
	snapshotStorages  = make(map[string]Storager)
)

// Register a snapshot storage implementation by name.
func Register(name string, s Storager) {
	if name == "" {
		panic("storage.Register(): could not register a StorageFactory with an empty name")
	}

	if s == nil {
		panic("storage.Register(): could not register a nil Storage interface")
	}

	snapshotStoragesM.Lock()
	defer snapshotStoragesM.Unlock()

	if _, dup := snapshotStorages[name]; dup {
		log.Printf("[ERROR] storage.Register(): called twice for " + name)
		return
	}

	log.Printf("storage.Register(): snapshot storage for the path [%s] registered\n", name)
	snapshotStorages[name] = s
}

// GetByPath returs a snapshot storage for a given path
func GetByPath(path string) Storager {
	snapshotStorage, ok := snapshotStorages[path]
	if ok {
		return snapshotStorage
	}
	log.Panicf("[ERROR] storage.GetByPath(): can't find snapshot for [%s]", path)
	return nil
}
