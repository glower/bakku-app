package gdrive

import (
	"context"
	"log"

	"github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/glower/bakku-app/pkg/types"
)

// Storage ...
type Storage struct {
	name                          string // storage name
	fileChangeNotificationChannel chan *types.FileChangeNotification
	fileStorageProgressCannel     chan *storage.Progress
	ctx                           context.Context
	storagePath                   string
	clientID                      string
	clientSecret                  string
}

const storageName = "gdrive"

func init() {
	storage.Register(storageName, &Storage{})
}

// SyncSnapshot syncs the snapshot dir to the storage
func (s *Storage) SyncSnapshot(from, to string) {}

// Setup gdrive storage
func (s *Storage) Setup(fileStorageProgressCannel chan *storage.Progress) bool {
	return true
}

// SyncLocalFilesToBackup ...
func (s *Storage) SyncLocalFilesToBackup() {}

// FileChangeNotification returns channel for notifications
func (s *Storage) FileChangeNotification() chan *types.FileChangeNotification {
	return s.fileChangeNotificationChannel
}

// Start local storage
func (s *Storage) Start(ctx context.Context) error {
	log.Println("storage.local.Start()")
	s.ctx = ctx
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case fileChange := <-s.fileChangeNotificationChannel:
				go s.handleFileChanges(fileChange)
			}
		}
	}()
	return nil
}

func (s *Storage) handleFileChanges(fileChange *types.FileChangeNotification) {}
