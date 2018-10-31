package fake

import (
	"context"
	"log"
	"time"

	"github.com/glower/bakku-app/pkg/backup/storage"
	conf "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/types"
)

// Storage fake
type Storage struct {
	name                          string // storage name
	fileChangeNotificationChannel chan *types.FileChangeNotification
	fileStorageProgressCannel     chan *storage.Progress
	ctx                           context.Context
}

const storageName = "fake"

func init() {
	storage.Register(storageName, &Storage{})
}

// Setup fake storage
func (s *Storage) Setup(fileStorageProgressCannel chan *storage.Progress) bool {
	config := conf.ProviderConf(storageName)
	if config.Active {
		s.name = storageName
		s.fileChangeNotificationChannel = make(chan *types.FileChangeNotification)
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
func (s *Storage) FileChangeNotification() chan *types.FileChangeNotification {
	return s.fileChangeNotificationChannel
}

// Start fake storage
func (s *Storage) Start(ctx context.Context) error {
	log.Println("storage.fake.Start()")
	s.ctx = ctx
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case fileChange := <-s.fileChangeNotificationChannel:
				s.handleFileChanges(fileChange)
			}
		}
	}()
	return nil
}

func (s *Storage) handleFileChanges(fileChange *types.FileChangeNotification) {
	log.Printf("storage.fake.handleFileChanges(): File %s has been changed\n", fileChange.Name)
	file := fileChange.Name
	storage.BackupStarted(file, storageName)
	s.store(file)
}

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
				progress := &storage.Progress{
					StorageName: storageName,
					FileName:    file,
					Percent:     p,
				}
				s.fileStorageProgressCannel <- progress
				if p >= float64(100.0) {
					storage.BackupFinished(file, storageName)
					return
				}
			}
		}
	}()
}
