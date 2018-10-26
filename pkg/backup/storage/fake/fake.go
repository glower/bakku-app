package fake

import (
	"context"
	"time"

	"github.com/glower/bakku-app/pkg/backup/storage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Storage fake
type Storage struct {
	name                          string // storage name
	fileChangeNotificationChannel chan *storage.FileChangeNotification
	fileStorageProgressCannel     chan *storage.Progress
	ctx                           context.Context
}

const storageName = "fake"

func init() {
	storage.Register(storageName, &Storage{})
}

// Setup fake storage
func (s *Storage) Setup(fileStorageProgressCannel chan *storage.Progress) bool {
	if isStorageConfigured() {
		s.name = storageName
		s.fileChangeNotificationChannel = make(chan *storage.FileChangeNotification)
		s.fileStorageProgressCannel = fileStorageProgressCannel
		return true
	}
	return false
}

// FileChangeNotification returns channel for notifications
func (s *Storage) FileChangeNotification() chan *storage.FileChangeNotification {
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

func (s *Storage) handleFileChanges(fileChange *storage.FileChangeNotification) {
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

func isStorageConfigured() bool {
	isActive, ok := viper.Get("backup.fake.activeX").(bool)
	if !ok {
		return false
	}
	return isActive
}
