package fake

import (
	"context"

	"github.com/glower/bakku-app/pkg/backup/storage"
	log "github.com/sirupsen/logrus"
)

// Storage fake
type Storage struct {
	name                          string // storage name
	fileChangeNotificationChannel chan *storage.FileChangeNotification
	ctx                           context.Context
}

func init() {
	log.Println("storage.fake.init()")
	storage.Register("fake", &Storage{})
}

// Setup fake storage
func (s *Storage) Setup(fileChangeNotificationChannel chan *storage.FileChangeNotification) bool {
	log.Println("storage.fake.Setup()")
	s.name = "fake"
	s.fileChangeNotificationChannel = fileChangeNotificationChannel
	return true
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
	log.Printf("storage.fake.handleFileChanges(): File %s has been changed\n", fileChange.File.Name())
}

func store(file string) {

}
