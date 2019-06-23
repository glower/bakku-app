package backup

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/glower/file-watcher/notification"

	"github.com/glower/bakku-app/pkg/event"
	"github.com/glower/bakku-app/pkg/message"
	"github.com/glower/bakku-app/pkg/storage"
	"github.com/glower/bakku-app/pkg/types"
)

type teardown func()

// Storage represents an interface for a backup storage provider
type Storage interface {
	Setup(*StorageManager) (bool, error)
	Store(*notification.Event) error
}

var teardowns = make(map[string]teardown)

// StorageManager ...
type StorageManager struct {
	Ctx context.Context

	LocalSnapshotStorage storage.Storage

	MessageCh            chan message.Message
	EventCh              chan notification.Event
	FileBackupProgressCh chan types.BackupProgress
	// fileBackupCompleteCh []chan types.FileBackupComplete
}

// Setup runs all implemented storages
func Setup(ctx context.Context, res types.GlobalResources, eventBuffer *event.Buffer) *StorageManager {
	// eventCh chan<- notification.Event, fileBackupCompleteCh chan types.FileBackupComplete
	m := &StorageManager{
		Ctx: ctx,

		EventCh:              eventBuffer.EvenOutCh, //eventCh,
		MessageCh:            res.MessageCh,
		LocalSnapshotStorage: res.Storage,
		FileBackupProgressCh: make(chan types.BackupProgress),
	}

	for name, storage := range GetAll() {
		ok, err := storage.Setup(m)
		if ok && err == nil {
			log.Printf("Setup(): backup storage [%s] is ready\n", name)
			_, cancel := context.WithCancel(context.Background())
			teardowns[name] = func() { cancel() }
		} else if !ok && err == nil {
			log.Printf("storage.SetupManager(): storage [%s] is not configured\n", name)
			Unregister(name)
		}
		if err != nil {
			m.MessageCh <- message.FormatMessage("ERROR", err.Error(), name)
		}
	}
	go m.ProcessNotifications(ctx)
	return m
}

// ProcessNotifications sends file change notofocations to all registerd storages
func (m *StorageManager) ProcessNotifications(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case file := <-m.EventCh:
			switch file.Action {
			case notification.FileRemoved:
				log.Printf("storage.processeFileChangeNotifications(): file=[%s] was deleted\n", file.AbsolutePath)
			case notification.FileAdded, notification.FileModified, notification.FileRenamedNewName:
				// log.Printf("backup.ProcessNotifications(): [%s] was added or modified\n", file.AbsolutePath)
				m.sendFileToAllStorages(&file)
			default:
				log.Printf("[ERROR] ProcessFileChangeNotifications(): unknown file change notification: %#v\n", file)
			}
		}
	}
}

func (m *StorageManager) sendFileToAllStorages(event *notification.Event) {
	for storageName, storageProvider := range GetAll() {
		go m.sendFileToStorage(event, storageProvider, storageName)
	}
}

func (m *StorageManager) sendFileToStorage(event *notification.Event, backup Storage, storageName string) {
	if InProgress(event, storageName) {
		return
	}

	Start(event, storageName)
	err := backup.Store(event)
	if err != nil {
		m.MessageCh <- message.FormatMessage("ERROR", err.Error(), storageName)
		Finish(event, storageName)
		return
	}

	m.updateLocalStorage(event, storageName)
}

func (m *StorageManager) updateLocalStorage(event *notification.Event, storageName string) {
	value, err := json.Marshal(&event)
	if err != nil {
		m.MessageCh <- message.FormatMessage("ERROR", err.Error(), "backup")
	}
	m.LocalSnapshotStorage.Add(event.AbsolutePath, storageName, value)
}

// Stop eveything
func Stop() {
	// block here untill all files are transferd
	for {
		<-time.After(1 * time.Second)
		inProgress := TotalFilesInProgres()
		log.Printf("TotalFilesInProgres: %d\n", inProgress)
		if true {
			teardownAll()
			return
		}
	}
}

func teardownAll() {
	for name, teardown := range teardowns {
		teardown()
		Unregister(name)
	}
}
