package storage

import (
	"sync"

	"log"

	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/file-watcher/notification"
)

const defultFolderName = "bakku-app"

// DefultFolderName returns a name for a folder where all backups should be stored
func DefultFolderName() string {
	return defultFolderName
}

// BackupStorage represents an interface for a backup storage provider
type BackupStorage interface {
	Setup(chan types.BackupProgress) bool
	Store(*notification.Event)
}

var (
	storagesM sync.RWMutex
	storages  = make(map[string]BackupStorage)
)

// Register a storage implementation by name.
func Register(name string, s BackupStorage) {
	if name == "" {
		panic("storage.Register(): could not register a StorageFactory with an empty name")
	}

	if s == nil {
		panic("storage.Register(): could not register a nil Storage interface")
	}

	storagesM.Lock()
	defer storagesM.Unlock()

	if _, dup := storages[name]; dup {
		log.Printf("[ERROR] storage.Register(): called twice for " + name)
		return
	}

	log.Printf("storage.Register(): storage provider [%s] registered\n", name)
	storages[name] = s
}

// GetAll returns a map of all registered backup storages
func GetAll() map[string]BackupStorage {
	return storages
}

// Unregister ...
func Unregister(name string) {
	storagesM.Lock()
	defer storagesM.Unlock()
	delete(storages, name)
}
