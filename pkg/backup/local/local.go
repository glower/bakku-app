package local

import (
	"path/filepath"

	"github.com/glower/bakku-app/pkg/backup"

	conf "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/types"

	"github.com/glower/file-watcher/notification"
)

// Storage local
type Storage struct {
	name string // storage name
	// eventCh               chan notification.Event
	// MessageCh             chan message.Message
	fileStorageProgressCh chan types.BackupProgress
	storagePath           string
	addLatency            bool
}

const storageName = "storage.local"
const bufferSize = 1024 * 1024

func init() {
	backup.Register(storageName, &Storage{})
}

// StoreOptions ...
type StoreOptions struct {
	reportProgress bool
	fileID         string
}

// Setup local storage
func (s *Storage) Setup(m *backup.StorageManager) (bool, error) {
	config := conf.LocalDriveConfig()
	if config.Active {
		s.name = storageName
		// s.eventCh = make(chan notification.Event)
		// s.MessageCh = m.MessageCh
		s.fileStorageProgressCh = m.FileBackupProgressCh
		storagePath := filepath.Clean(config.Path)
		s.storagePath = storagePath
		s.addLatency = config.AddLatency
		return true, nil
	}
	return false, nil
}

// Store stores a file to a local storage
func (s *Storage) Store(event *notification.Event) error {
	absolutePath := event.AbsolutePath
	relativePath := event.RelativePath
	directoryPath := event.DirectoryPath

	// fmt.Printf("\nlocal.Store():\n")
	// fmt.Printf(">\tabsolutePath:\t%s\n>\trelativePath:\t%s\n>\tdirectoryPath:\t%s\n\n", absolutePath, relativePath, directoryPath)

	from := absolutePath
	to := filepath.Join(s.storagePath, filepath.Base(directoryPath), relativePath)
	return s.store(from, to, StoreOptions{
		reportProgress: true,
		fileID:         event.UUID.String(), // Or checksum?
	})
}
