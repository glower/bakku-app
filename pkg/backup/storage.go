package backup

import (
	"sync"

	"log"
)

const defultFolderName = "bakku-app"

// DefultFolderName returns a name for a folder where all backups should be stored
func DefultFolderName() string {
	return defultFolderName
}

var (
	storagesM sync.RWMutex
	storages  = make(map[string]Storage)
)

// Register a storage implementation by name.
func Register(name string, s Storage) {
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
func GetAll() map[string]Storage {
	return storages
}

// Unregister ...
func Unregister(name string) {
	storagesM.Lock()
	defer storagesM.Unlock()
	delete(storages, name)
}
