package fake

import (
	"context"
	"time"

	"github.com/glower/file-watcher/notification"

	conf "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/types"
)

// Storage fake
type Storage struct {
	name                  string // storage name
	eventCh               chan notification.Event
	fileStorageProgressCh chan types.BackupProgress
	ctx                   context.Context
}

const storageName = "fake"

func init() {
	// storage.Register(storageName, &Storage{})
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

// SyncLocalFilesToBackup ...
func (s *Storage) SyncLocalFilesToBackup() {}

// SyncSnapshot ...
func (s *Storage) SyncSnapshot(from, to string) {}

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
				s.fileStorageProgressCh <- types.BackupProgress{
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
