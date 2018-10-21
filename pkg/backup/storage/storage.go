package storage

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/glower/bakku-app/pkg/watchers/watch"
	"github.com/r3labs/sse"
	log "github.com/sirupsen/logrus"
)

// Storage ...
type Storage interface {
	Setup(chan *Progress) bool
	Start(ctx context.Context) error
	FileChangeNotification() chan *FileChangeNotification
}

// FileChangeNotification ...
type FileChangeNotification struct {
	File   os.FileInfo
	Action watch.Action
}

// Manager ...
type Manager struct {
	FileChangeNotificationChannel chan *FileChangeNotification
	ProgressChannel               chan *Progress
	SSEServer                     *sse.Server
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
		panic("storage.Register(): could not register a StorageFactory with an empty name")
	}

	if s == nil {
		panic("storage.Register(): could not register a nil Storage interface")
	}

	storagesM.Lock()
	defer storagesM.Unlock()

	if _, dup := storages[name]; dup {
		panic("storageRegister(): called twice for " + name)
	}

	log.WithFields(log.Fields{
		"name": name,
	}).Info("storage.Register(): registered")

	storages[name] = s
}

// UnregisterStorage removes a Storage with a particular name from the list.
func UnregisterStorage(name string) {
	storagesM.Lock()
	defer storagesM.Unlock()
	delete(storages, name)
}

// SetupManager runs all implemented storages
func SetupManager(sseServer *sse.Server) *Manager {
	log.Println("strorage.Run()")
	m := &Manager{
		FileChangeNotificationChannel: make(chan *FileChangeNotification),
		ProgressChannel:               make(chan *Progress),
		SSEServer:                     sseServer,
	}
	for name, storage := range storages {
		ok := storage.Setup(m.ProgressChannel)
		if ok {
			m.SetupStorage(name, storage)
		} else {
			log.Errorf("Run(): can not get configuration for storage [%s]", name)
		}
	}
	return m
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

		go m.ProcessProgressCallback(ctx)
		go m.ProcessFileChangeNotifications(ctx)
	}
}

// ProcessFileChangeNotifications sends file change notofocations to all registerd storages
func (m *Manager) ProcessFileChangeNotifications(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case change := <-m.FileChangeNotificationChannel:
		for name, storage := range storages {
			log.Printf("storage.ProcessFileChangeNotifications(): send notification to [%s] storage provider\n", name)
			storage.FileChangeNotification() <- change
		}
	}
}

// ProcessProgressCallback ...
func (m *Manager) ProcessProgressCallback(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case progress := <-m.ProgressChannel:
			log.Printf("[%s] [%s]\t%.2f%%\n", progress.StorageName, progress.FileName, progress.Percent)
			progressJSON, _ := json.Marshal(progress)
			// file fotification for the frontend client over the SSE
			m.SSEServer.Publish("files", &sse.Event{
				Data: []byte(progressJSON),
			})
		}
	}
}

// Stop eveything
func Stop() {
	// block here untill all files are transferd
	for {
		select {
		case <-time.After(1 * time.Second):
			if TotalFilesInProgres() == 0 {
				teardownAll()
				return
			}
		}
	}

}

func teardownAll() {
	for name, teardown := range teardowns {
		log.Infof("Teardown %s storage", name)
		teardown()
		UnregisterStorage(name)
	}
	log.Println("storage.Stop(): eveything is stoped")
}
