package fake

import (
	"context"
	"time"

	"github.com/glower/bakku-app/pkg/backup/storage"
	log "github.com/sirupsen/logrus"
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
	log.Println("storage.fake.init()")
	storage.Register(storageName, &Storage{})
}

// Setup fake storage
func (s *Storage) Setup(fileStorageProgressCannel chan *storage.Progress) bool {
	log.Println("storage.fake.Setup()")
	s.name = storageName
	s.fileChangeNotificationChannel = make(chan *storage.FileChangeNotification)
	s.fileStorageProgressCannel = fileStorageProgressCannel
	return true
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
				go s.handleFileChanges(fileChange)
			}
		}
	}()
	return nil
}

func (s *Storage) handleFileChanges(fileChange *storage.FileChangeNotification) {
	log.Printf("storage.fake.handleFileChanges(): File %s has been changed\n", fileChange.File.Name())
	file := fileChange.File.Name()
	storage.BackupStarted(file, storageName)
	s.store(file)
	storage.BackupFinished(file, storageName)
}

func (s *Storage) store(file string) {
	p := 0.0
	for {
		select {
		case <-s.ctx.Done():
			// context has finished - exit
			return
		case <-time.After(1 * time.Second):
			p = p + 10
			progress := &storage.Progress{
				StorageName: storageName,
				FileName:    file,
				Percent:     p,
			}
			s.fileStorageProgressCannel <- progress
			if p >= float64(100.0) {
				// log.Printf("storage.fake.store(): Done uploading file [%s]\n", file)
				return
			}
		}
	}
}
