package fake

import (
	"math/rand"
	"time"

	"github.com/glower/file-watcher/notification"

	"github.com/glower/bakku-app/pkg/backup/storage"
	conf "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/types"
)

// Storage fake
type Storage struct {
	name                  string // storage name
	eventCh               chan notification.Event
	fileStorageProgressCh chan types.BackupProgress
}

const storageName = "fake"

func init() {
	storage.Register(storageName, &Storage{})
}

// Setup fake storage
func (s *Storage) Setup(fileStorageProgressCh chan types.BackupProgress) bool {
	config := conf.ProviderConf(storageName)
	if config.Active {
		s.name = storageName
		s.eventCh = make(chan notification.Event)
		s.fileStorageProgressCh = fileStorageProgressCh
		return true
	}
	return false
}

func (s *Storage) Store(ev *notification.Event) {
	file := ev.AbsolutePath
	p := 0.0
	for {
		select {
		case <-time.After(1 * time.Second):
			sleepRandom()
			p = p + 5
			s.fileStorageProgressCh <- types.BackupProgress{
				StorageName: storageName,
				FileName:    file,
				Percent:     p,
			}
			if p >= float64(100.0) {
				return
			}
		}
	}
}

func sleepRandom() {
	r := 100000 + rand.Intn(500000)
	time.Sleep(time.Duration(r) * time.Microsecond)
}
