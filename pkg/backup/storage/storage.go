package storage

import (
	"context"
	"os"
	"sync"

	"github.com/glower/file-change-notification/watch"
	log "github.com/sirupsen/logrus"
)

// Storage ...
type Storage interface {
	Setup(chan *FileChangeNotification) bool
	Start(ctx context.Context) error
}

// FileChangeNotification ...
type FileChangeNotification struct {
	File   os.FileInfo
	Action watch.Action
}

// Manager ...
type Manager struct {
	fileChangeNotificationChannel chan *FileChangeNotification
}

type teardown func()

var (
	storagesM sync.RWMutex
	storages  = make(map[string]Storage)
	teardowns = make(map[string]teardown)
)

// Register a storage implementation by name.
func Register(name string, s Storage) {
	if name == "" {
		panic("Register(): could not register a StorageFactory with an empty name")
	}

	if s == nil {
		panic("Register(): could not register a nil Storage interface")
	}

	storagesM.Lock()
	defer storagesM.Unlock()

	if _, dup := storages[name]; dup {
		panic("Register(): called twice for " + name)
	}

	log.WithFields(log.Fields{
		"name": name,
	}).Info("Register(): registered")

	storages[name] = s
}

// UnregisterStorage removes a Storage with a particular name from the list.
func UnregisterStorage(name string) {
	storagesM.Lock()
	defer storagesM.Unlock()
	delete(storages, name)
}

// Run all implemented storages
func Run() {
	m := &Manager{
		fileChangeNotificationChannel: make(chan *FileChangeNotification),
	}
	for name, storage := range storages {
		ok := storage.Setup(m.fileChangeNotificationChannel)
		if ok {
			m.SetupStorage(name, storage)
		} else {
			log.Errorf("Run(): can not get configuration for storage [%s]", name)
		}
	}
}

// SetupStorage ...
func (m *Manager) SetupStorage(name string, storage Storage) {
	ctx, cancel := context.WithCancel(context.Background())
	err := storage.Start(ctx)
	if err != nil {
		cancel()
		log.WithFields(log.Fields{
			"error": err,
		}).Fatalf("main: failed to setup %s bot\n", name)
	} else {
		// store cancelling context for each storage
		teardowns[name] = func() { cancel() }
	}
}

// Stop eveything
func Stop() {
	for name, teardown := range teardowns {
		log.Infof("Teardown %s storage", name)
		teardown()
		UnregisterStorage(name)
	}
}
