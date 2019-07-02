package fake

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"time"

	"github.com/glower/file-watcher/notification"

	"github.com/glower/bakku-app/pkg/backup"
	conf "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/message"
	"github.com/glower/bakku-app/pkg/types"
)

// Storage fake
type Storage struct {
	name                  string // storage name
	eventCh               chan notification.Event
	MessageCh             chan message.Message
	fileStorageProgressCh chan types.BackupProgress
}

const storageName = "storage.fake"

func init() {
	backup.Register(storageName, &Storage{})
}

// Setup fake storage
func (s *Storage) Setup(m *backup.StorageManager) (bool, error) {
	config := conf.FakeDriverConfig()
	if config.Active {
		s.name = storageName
		s.eventCh = make(chan notification.Event)
		s.fileStorageProgressCh = m.FileBackupProgressCh
		return true, nil
	}
	return false, nil
}

// Store file on event
func (s *Storage) Store(ev *notification.Event) error {
	file := ev.AbsolutePath
	data := []byte(file)
	p := 0.0
	if rand.Intn(10) < 4 {
		return fmt.Errorf("Random error")
	}
	for {
		<-time.After(1 * time.Second)
		// sleepRandom()
		p = p + 1 + float64(rand.Intn(4))
		s.fileStorageProgressCh <- types.BackupProgress{
			AbsolutePath: file,
			StorageName:  storageName,
			FileName:     ev.FileName,
			ID:           fmt.Sprintf("%x", sha1.Sum(data)),
			Percent:      p,
		}
		if p >= float64(100.0) {
			return nil
		}
	}
}

func sleepRandom() {
	r := 1000000 + rand.Intn(3000000)
	time.Sleep(time.Duration(r) * time.Microsecond)
}
