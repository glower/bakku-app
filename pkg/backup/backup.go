package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/glower/file-watcher/notification"

	"github.com/glower/bakku-app/pkg/event"
	"github.com/glower/bakku-app/pkg/message"
	"github.com/glower/bakku-app/pkg/storage"
	"github.com/glower/bakku-app/pkg/types"
)

// token represents a request that is being processed.
type token struct{}

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

	tokens               chan token
	MessageCh            chan message.Message
	EventCh              chan notification.Event
	FileBackupProgressCh chan types.BackupProgress
	LocalSnapshotStorage storage.Storager
	r                    *types.GlobalResources
}

// Setup runs all implemented storages
func Setup(ctx context.Context, res *types.GlobalResources, eventBuffer *event.Buffer) *StorageManager {
	m := &StorageManager{
		Ctx:     ctx,
		EventCh: eventBuffer.EvenOutCh,
		// MessageCh:            res.MessageCh,
		LocalSnapshotStorage: res.Storage,
		FileBackupProgressCh: make(chan types.BackupProgress),
		tokens:               make(chan token, 5),
		r:                    res,
		// BackupCompleteCh:     res.BackupCompleteCh,
	}

	for i := 0; i < 5; i++ {
		m.tokens <- token{}
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
			m.r.MessageCh <- message.FormatMessage("ERROR", err.Error(), name)
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
				fmt.Printf("backup: file=[%s] was deleted\n", file.AbsolutePath)
			case notification.FileAdded, notification.FileModified, notification.FileRenamedNewName:
				fmt.Printf(">>>>> backup: file [%s] was added, in progres=%d\n", file.AbsolutePath, TotalFilesInProgres())
				for storageName, storageProvider := range GetAll() {
					tok := <-m.tokens
					go m.sendFileToStorage(&file, storageProvider, storageName, tok)
				}

			default:
				log.Printf("[ERROR] ProcessFileChangeNotifications(): unknown file change notification: %#v\n", file)
			}
		}
	}
}

func (m *StorageManager) sendFileToStorage(event *notification.Event, backup Storage, storageName string, t token) {
	// fmt.Printf(">>>>> sendFileToStorage(): backup %s => %s BEGIN\n", event.AbsolutePath, storageName)

	if InProgress(event, storageName) {
		return
	}

	Start(event, storageName)
	err := backup.Store(event)
	Finish(event, storageName)

	if err != nil {
		m.r.MessageCh <- message.FormatMessage("ERROR", err.Error(), storageName)
		return
	}

	err = m.updateLocalStorage(event, storageName)
	if err != nil {
		m.r.MessageCh <- message.FormatMessage("ERROR", err.Error(), storageName)
		return
	}
	// fmt.Printf(">>>>> sendFileToStorage():backup  %s => %s DONE\n", event.AbsolutePath, storageName)
	m.r.BackupCompleteCh <- types.BackupComplete{
		StorageName: storageName,
		FilePath:    event.AbsolutePath,
	}
	// fmt.Printf(">>>>> sendFileToStorage(): backup  %s => %s BackupCompleteCh received\n", event.AbsolutePath, storageName)
	m.tokens <- t
}

func (m *StorageManager) updateLocalStorage(event *notification.Event, storageName string) error {
	value, err := json.Marshal(&event)
	if err != nil {
		return err
	}
	err = m.LocalSnapshotStorage.Add(event.AbsolutePath, storageName, value)
	if err != nil {
		return err
	}
	return nil
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
