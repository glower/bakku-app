package fake

import (
	"context"
	"log"

	"github.com/glower/bakku-app/pkg/backup/storage"
)

// Storage fake
type Storage struct {
	name                          string // storage name
	fileChangeNotificationChannel chan *storage.FileChangeNotification
	ctx                           context.Context
}

func init() {
	storage.Register("hipchat", &Storage{})
}

// Setup fake storage
func (s *Storage) Setup(fileChangeNotificationChannel chan *storage.FileChangeNotification) bool {
	s.name = "fake"
	s.fileChangeNotificationChannel = fileChangeNotificationChannel
	return true
}

// Start fake storage
func (s *Storage) Start(ctx context.Context) error {
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
	log.Printf("File %s has been changed\n", fileChange.File.Name())
}

func store(file string) {

}
