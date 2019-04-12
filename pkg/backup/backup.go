package backup

import (
	"context"
	"log"
	"time"

	backupstorage "github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/glower/file-watcher/notification"

	"github.com/glower/bakku-app/pkg/types"
)

type teardown func()

var teardowns = make(map[string]teardown)

// StorageManager ...
type StorageManager struct {
	ctx context.Context

	EventCh              chan notification.Event
	FileBackupProgressCh chan types.BackupProgress
	FileBackupCompleteCh chan types.FileBackupComplete
}

// Setup runs all implemented storages
func Setup(ctx context.Context, eventCh chan notification.Event) *StorageManager {
	m := &StorageManager{
		ctx: ctx,

		EventCh:              eventCh,
		FileBackupProgressCh: make(chan types.BackupProgress),
		FileBackupCompleteCh: make(chan types.FileBackupComplete),
	}

	for name, storage := range backupstorage.GetAll() {
		ok := storage.Setup(m.FileBackupProgressCh)
		if ok {
			log.Printf("Setup(): backup storage [%s] is ready\n", name)
			_, cancel := context.WithCancel(context.Background())
			teardowns[name] = func() { cancel() }
		} else {
			log.Printf("storage.SetupManager(): storage [%s] is not configured\n", name)
			backupstorage.Unregister(name)
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
				//// TODO: XXX
				// if len(file.BackupToStorages) > 0 {
				// 	storages := backupstorage.GetAll()
				// 	for _, storageName := range file.BackupToStorages {
				// 		if storageProvider, ok := storages[storageName]; ok {
				// 			go m.sendFileToStorage(&file, storageProvider, storageName)
				// 		}
				// 	}
				// } else {
				// 	m.sendFileToAllStorages(&file)
				// }

			default:
				log.Printf("[ERROR] ProcessFileChangeNotifications(): unknown file change notification: %#v\n", file)
			}
		}
	}
}

func (m *StorageManager) sendFileToAllStorages(event *notification.Event) {
	for storageName, storageProvider := range backupstorage.GetAll() {
		// log.Printf("storage.sendFileToAllStorages(): send notification to [%s] storage provider\n", storageName)
		go m.sendFileToStorage(event, storageProvider, storageName)
	}
}

func (m *StorageManager) sendFileToStorage(event *notification.Event, backup backupstorage.BackupStorage, storageName string) {
	if InProgress(event, storageName) {
		return
	}

	// log.Printf("sendFileToStorage(): send file [%s] to storage [%s]", event.AbsolutePath, storageName)
	Start(event, storageName)
	backup.Store(event)
	Finish(event, storageName)
	// log.Printf("sendFileToStorage(): backup of [%s] to storage [%s] is complete", event.AbsolutePath, storageName)

	m.FileBackupCompleteCh <- types.FileBackupComplete{
		BackupStorageName:  storageName,
		AbsolutePath:       event.AbsolutePath,
		WatchDirectoryName: event.WatchDirectoryName,
	}
}

// Stop eveything
func Stop() {
	// block here untill all files are transferd
	for {
		select {
		case <-time.After(1 * time.Second):
			inProgress := TotalFilesInProgres()
			log.Printf("TotalFilesInProgres: %d\n", inProgress)
			if true {
				teardownAll()
				return
			}
		}
	}
}

func teardownAll() {
	for name, teardown := range teardowns {
		teardown()
		backupstorage.Unregister(name)
	}
}
