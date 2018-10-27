package storage

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"log"

	"github.com/glower/bakku-app/pkg/types"
	"github.com/r3labs/sse"
)

// Storage ...
type Storage interface {
	Setup(chan *Progress) bool
	Start(ctx context.Context) error
	FileChangeNotification() chan *types.FileChangeNotification
	SyncLocalFilesToBackup()
	SyncSnapshot(string, string)
}

// Manager ...
type Manager struct {
	ctx                           context.Context
	FileChangeNotificationChannel chan *types.FileChangeNotification
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
		log.Printf("[ERROR] storage.Register(): called twice for " + name)
		return
	}

	log.Printf("storage.Register(): storage provider [%s] registered\n", name)
	storages[name] = s
}

// UnregisterStorage removes a Storage with a particular name from the list.
func UnregisterStorage(name string) {
	storagesM.Lock()
	defer storagesM.Unlock()
	delete(storages, name)
}

// SetupManager runs all implemented storages
func SetupManager(ctx context.Context, sseServer *sse.Server, notifications []chan types.FileChangeNotification) *Manager {
	m := &Manager{
		FileChangeNotificationChannel: make(chan *types.FileChangeNotification),
		ProgressChannel:               make(chan *Progress),
		SSEServer:                     sseServer,
		ctx:                           ctx,
	}
	for name, storage := range storages {
		ok := storage.Setup(m.ProgressChannel)
		if ok {
			m.SetupStorage(name, storage)
		} else {
			log.Printf("SetupManager(): storage [%s] is not configured\n", name)
			UnregisterStorage(name)
		}
	}

	go m.ProcessFileChangeNotifications(ctx, notifications)

	return m
}

// SetupStorage ...
func (m *Manager) SetupStorage(name string, storage Storage) {
	ctx, cancel := context.WithCancel(context.Background())
	go storage.SyncLocalFilesToBackup()
	err := storage.Start(ctx)
	if err != nil {
		cancel()
		log.Printf("[ERROR] SetupStorage: failed to setup storage [%s]\n", name)
	} else {
		// store cancelling context for each storage
		teardowns[name] = func() { cancel() }
		// TODO: dose it make sence to start it for each storage???
		// If so, we don't need a for loop over all storages
		go m.ProcessProgressCallback(ctx)
	}
}

// ProcessFileChangeNotifications sends file change notofocations to all registerd storages
func (m *Manager) ProcessFileChangeNotifications(ctx context.Context, notifications []chan types.FileChangeNotification) {
	for _, watcher := range notifications {
		for {
			select {
			case <-ctx.Done():
				return
			case change := <-watcher:
				log.Printf("storage.ProcessFileChangeNotifications(): file=[%s]\n", change.AbsolutePath)
				for name, storage := range storages {
					log.Printf("storage.ProcessFileChangeNotifications(): send notification to [%s] storage provider\n", name)
					storage.FileChangeNotification() <- &change
				}
			}
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
			// log.Printf("ProcessProgressCallback(): [%s] [%s]\t%.2f%%\n", progress.StorageName, progress.FileName, progress.Percent)
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
		log.Printf("storage.Stop(): Teardown storage [%s]\n", name)
		teardown()
		UnregisterStorage(name)
	}
	log.Println("storage.Stop(): eveything is stoped")
}
