package backup

import (
	"context"
	"log"
	"time"

	"github.com/glower/file-watcher/notification"

	"github.com/glower/bakku-app/pkg/event"
	"github.com/glower/bakku-app/pkg/message"
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

	MessageCh            chan message.Message
	EventCh              chan notification.Event
	FileBackupProgressCh chan types.BackupProgress
	fileBackupCompleteCh []chan types.FileBackupComplete

	// BackupEventsSubscriber chan chan interface{}
}

// Setup runs all implemented storages
func Setup(ctx context.Context, messageCh chan message.Message, eventBuffer *event.Buffer) *StorageManager {
	// eventCh chan<- notification.Event, fileBackupCompleteCh chan types.FileBackupComplete
	m := &StorageManager{
		Ctx: ctx,

		EventCh:              eventBuffer.EvenOutCh, //eventCh,
		MessageCh:            messageCh,
		FileBackupProgressCh: make(chan types.BackupProgress),
		fileBackupCompleteCh: []chan types.FileBackupComplete{},

		// BackupEventsSubscriber: make(chan chan interface{}),
	}

	m.fileBackupCompleteCh = append(m.fileBackupCompleteCh, eventBuffer.BackupDoneCh)

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
			log.Printf("[ERROR] Setup(): backup storage [%s] error=%v\n", name, err)
			// TODO: write error to the error channel
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
		// log.Printf("storage.sendFileToAllStorages(): send notification to [%s] storage provider\n", storageName)
		go m.sendFileToStorage(event, storageProvider, storageName)
	}
}

func (m *StorageManager) sendFileToStorage(event *notification.Event, backup Storage, storageName string) {
	if InProgress(event, storageName) {
		return
	}

	log.Printf("!!!!!!!!!!!!!! sendFileToStorage(): send file [%s] to storage [%s]", event.AbsolutePath, storageName)
	Start(event, storageName)
	err := backup.Store(event)
	if err != nil {
		m.MessageCh <- message.FormatMessage("ERROR", err.Error(), storageName)
		Finish(event, storageName)
		return
	}

	log.Printf("sendFileToStorage(): backup of [%s] to storage [%s] is complete", event.AbsolutePath, storageName)
	m.fileBackupComple(types.FileBackupComplete{
		BackupStorageName:  storageName,
		AbsolutePath:       event.AbsolutePath,
		WatchDirectoryName: event.WatchDirectoryName,
	})
}

func (m *StorageManager) SubscribeForFileBackupCompleteEvent(ch chan types.FileBackupComplete) {
	// TODO: add mutex
	m.fileBackupCompleteCh = append(m.fileBackupCompleteCh, ch)
}

func (m *StorageManager) fileBackupComple(e types.FileBackupComplete) {
	for _, ch := range m.fileBackupCompleteCh {
		ch <- e
	}
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
