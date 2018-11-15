package storage

import (
	"log"
	"sync"

	"github.com/glower/bakku-app/pkg/config/snapshot"
)

// Snapshot ...
type Snapshot interface {
	Register() bool
	Path(path string) Snapshot
	SnapshotStoragePath() string
	Exist() bool
	Add(string, []byte) error
	GetAll() (map[string]string, error)
	Get(string) (string, error)
	Remove(string) error
}

var (
	snapshotStoragesM sync.RWMutex
	snapshotStorages  = make(map[string]Snapshot)
)

// Register a snapshot storage implementation by name.
func Register(name string, s Snapshot) {
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

	log.Printf("storage.Register(): storage provider [%s] registered\n", name)
	snapshotStorages[name] = s
}

// GetDefault a snapshot storage implementation
func GetDefault() Snapshot {
	defaultSnapshotStorage := snapshot.DefaultStorage()
	snapshotStorage, ok := snapshotStorages[defaultSnapshotStorage]
	if ok {
		return snapshotStorage
	}
	log.Panicf("default snapshot storage [%s] is not implemented", defaultSnapshotStorage)
	return nil
}
