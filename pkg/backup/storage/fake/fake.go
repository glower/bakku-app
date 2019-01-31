package fake

import (
	"context"
	"time"

	conf "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/types"
)

// Storage fake
type Storage struct {
	name                          string // storage name
	fileChangeNotificationChannel chan types.FileChangeNotification
	fileStorageProgressCannel     chan types.BackupProgress
	ctx                           context.Context
}

const storageName = "fake"

func init() {
	// storage.Register(storageName, &Storage{})
}

// Setup fake storage
func (s *Storage) Setup(fileStorageProgressCannel chan types.BackupProgress) bool {
	config := conf.ProviderConf(storageName)
	if config.Active {
		s.name = storageName
		s.fileChangeNotificationChannel = make(chan types.FileChangeNotification)
		s.fileStorageProgressCannel = fileStorageProgressCannel
		return true
	}
	return false
}

// SyncLocalFilesToBackup ...
func (s *Storage) SyncLocalFilesToBackup() {}

// SyncSnapshot ...
func (s *Storage) SyncSnapshot(from, to string) {}

// FileChangeNotification returns channel for notifications
func (s *Storage) FileChangeNotification() chan types.FileChangeNotification {
	return s.fileChangeNotificationChannel
}

// func (s *Storage) HandleFileChanges(fileChange *types.FileChangeNotification) {
// 	log.Printf("storage.fake.handleFileChanges(): File %s has been changed\n", fileChange.Name)
// 	file := fileChange.Name
// 	storage.BackupStarted(file, storageName)
// 	s.store(file)
// }

func (s *Storage) store(file string) {
	p := 0.0
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				// context has finished - exit
				return
			case <-time.After(1 * time.Second):
				p = p + 50
				s.fileStorageProgressCannel <- types.BackupProgress{
					StorageName: storageName,
					FileName:    file,
					Percent:     p,
				}
				if p >= float64(100.0) {
					// backup.Finished(file, storageName)
					return
				}
			}
		}
	}()
}
