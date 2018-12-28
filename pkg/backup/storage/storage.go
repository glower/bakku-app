package storage

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"log"

	"github.com/glower/bakku-app/pkg/backup"
	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/r3labs/sse"
)

// Storage represents an interface for a backup storage provider
type Storage interface {
	Setup(chan *backup.Progress) bool
	SyncSnapshot(*types.FileChangeNotification)
	Store(*types.FileChangeNotification)
	SyncLocalFilesToBackup()
}

// Manager ...
type Manager struct {
	ctx                           context.Context
	FileChangeNotificationChannel chan *types.FileChangeNotification
	ProgressChannel               chan *backup.Progress
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
func SetupManager(ctx context.Context, sseServer *sse.Server, notifications []types.Notifications) *Manager {
	m := &Manager{
		FileChangeNotificationChannel: make(chan *types.FileChangeNotification),
		ProgressChannel:               make(chan *backup.Progress),
		SSEServer:                     sseServer,
		ctx:                           ctx,
	}
	for name, storage := range storages {
		ok := storage.Setup(m.ProgressChannel)
		if ok {
			m.SetupStorage(name, storage)
		} else {
			log.Printf("storage.SetupManager(): storage [%s] is not configured\n", name)
			UnregisterStorage(name)
		}
	}

	go m.ProcessNotifications(ctx, notifications)

	return m
}

// SetupStorage ...
func (m *Manager) SetupStorage(name string, storage Storage) {
	log.Printf("SetupStorage(): [%s]\n", name)
	ctx, cancel := context.WithCancel(context.Background())
	teardowns[name] = func() { cancel() }
	go m.ProcessProgressCallback(ctx)
}

func processeFileChangeNotifications(ctx context.Context, watcher <-chan types.FileChangeNotification) {
	for {
		select {
		case <-ctx.Done():
			return
		case file := <-watcher:
			switch change.Action {
			case types.FileRemoved:
				log.Printf("storage.ProcessFileChangeNotifications(): file=[%s] was deleted\n", change.AbsolutePath)
				snapshot.RemoveSnapshotEntry(change.DirectoryPath, change.AbsolutePath) // TODO: is here a  good place?
			case types.FileAdded, types.FileModified, types.FileRenamedNewName:
				if len(file.BackupToStorages) > 0 {
					for _, storageName := range file.BackupToStoragesar {
						if storageProvider, ok := [storages]; ok {
							go storages(&file, storageProvider, storageName)
						}
					}
					return
				}
				sendFileToAllStorages(file)

			default:
				log.Printf("[ERROR] ProcessFileChangeNotifications(): unknown file change notification: %#v\n", change)
			}
		}
	}
}

func sendFileToAllStorages(file *types.FileChangeNotification) {
	for storageName, storageProvider := range storages {
		log.Printf("storage.ProcessFileChangeNotifications(): send notification to [%s] storage provider\n", name)
		go storages(&file, storageProvider, storageName)
	}
}

func sendFileToStorage(fileChange *types.FileChangeNotification, s Storage, storageName string) {
	log.Printf("handleFileChanges(): File [%s] has been changed\n", fileChange.AbsolutePath)
	if !backup.InProgress(fileChange, storageName) {
		backup.Start(fileChange, storageName)
		s.Store(fileChange)
		backup.Finished(fileChange, storageName)
		snapshot.UpdateEntry(fileChange, storageName)
		s.SyncSnapshot(fileChange)
	}
}

func processeFilesScanDoneNotifications(ctx context.Context, done <-chan bool) {
	log.Println("processeFilesScanDoneNotifications(): setup channels")
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			for name := range storages {
				// compare local files with remote and copy local files to the backup
				log.Printf("storage.processeFilesScanDoneNotifications(): sync local files to backups for [%s]\n", name)
				// go storage.SyncLocalFilesToBackup() //
			}
		}
	}
}

// ProcessNotifications sends file change notofocations to all registerd storages
func (m *Manager) ProcessNotifications(ctx context.Context, notifications []types.Notifications) {
	log.Println("backup.ProcessNotifications(): setup channels")
	for _, notification := range notifications {
		go processeFileChangeNotifications(ctx, notification.FileChangeChan)
		go processeFilesScanDoneNotifications(ctx, notification.DoneChan)
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
			if backup.TotalFilesInProgres() == 0 {
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
