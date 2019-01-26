package backup

import (
	"context"
	"log"
	"time"

	backupstorage "github.com/glower/bakku-app/pkg/backup/storage"

	"github.com/glower/bakku-app/pkg/types"
)

type teardown func()

var teardowns = make(map[string]teardown)

// StorageManager ...
type StorageManager struct {
	ctx context.Context

	FileChangeNotificationChannel chan *types.FileChangeNotification
	FileBackupProgressChannel     chan *types.BackupProgress
	FileBackupCompleteChannel     chan *types.FileBackupComplete
}

// Setup runs all implemented storages
func Setup(ctx context.Context, notification chan *types.FileChangeNotification) *StorageManager {
	m := &StorageManager{
		ctx: ctx,

		FileChangeNotificationChannel: notification,
		FileBackupProgressChannel:     make(chan *types.BackupProgress),
		FileBackupCompleteChannel:     make(chan *types.FileBackupComplete),
	}

	for name, storage := range backupstorage.GetAll() {
		ok := storage.Setup(m.FileBackupProgressChannel)
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
		case file := <-m.FileChangeNotificationChannel:
			switch file.Action {
			case types.FileRemoved:
				log.Printf("storage.processeFileChangeNotifications(): file=[%s] was deleted\n", file.AbsolutePath)
			case types.FileAdded, types.FileModified, types.FileRenamedNewName:
				if len(file.BackupToStorages) > 0 {
					storages := backupstorage.GetAll()
					for _, storageName := range file.BackupToStorages {
						if storageProvider, ok := storages[storageName]; ok {
							go sendFileToStorage(file, storageProvider, storageName)
						}
					}
					return
				}
				sendFileToAllStorages(file)

			default:
				log.Printf("[ERROR] ProcessFileChangeNotifications(): unknown file change notification: %#v\n", file)
			}
		}
	}

}

func sendFileToAllStorages(file *types.FileChangeNotification) {
	for storageName, storageProvider := range backupstorage.GetAll() {
		log.Printf("storage.sendFileToAllStorages(): send notification to [%s] storage provider\n", storageName)
		go sendFileToStorage(file, storageProvider, storageName)
	}
}

func sendFileToStorage(fileChange *types.FileChangeNotification, backup backupstorage.BackupStorage, storageName string) {
	log.Printf("sendFileToStorage(): File [%s] has been changed\n", fileChange.AbsolutePath)
	if !InProgress(fileChange, storageName) {
		Start(fileChange, storageName)
		backup.Store(fileChange)
		Finished(fileChange, storageName)
		// snapshotStorage.UpdateEntry(fileChange, storageName)
		backup.SyncSnapshot(fileChange)
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
		teardown()
		backupstorage.Unregister(name)
	}
}

// func processeFilesScanDoneNotifications(ctx context.Context, done <-chan bool) {
// 	log.Println("processeFilesScanDoneNotifications(): setup channels")
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case <-done:
// 			for name := range storages {
// 				// compare local files with remote and copy local files to the backup
// 				log.Printf("storage.processeFilesScanDoneNotifications(): sync local files to backups for [%s]\n", name)
// 				// go storage.SyncLocalFilesToBackup() //
// 			}
// 		}
// 	}
// }

// ProcessProgressCallback ...
// func (m *StorageManager) ProcessProgressCallback(ctx context.Context) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case progress := <-m.ProgressChannel:
// 			// log.Printf("ProcessProgressCallback(): [%s] [%s]\t%.2f%%\n", progress.StorageName, progress.FileName, progress.Percent)
// 			progressJSON, _ := json.Marshal(progress)
// 			// file fotification for the frontend client over the SSE
// 			m.SSEServer.Publish("files", &sse.Event{
// 				Data: []byte(progressJSON),
// 			})
// 		}
// 	}
// }
