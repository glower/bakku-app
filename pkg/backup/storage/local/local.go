package local

import (
	"fmt"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/backup/storage"
	conf "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/types"

	"github.com/glower/file-watcher/notification"
)

// Storage local
type Storage struct {
	name                  string // storage name
	eventCh               chan notification.Event
	fileStorageProgressCh chan types.BackupProgress
	storagePath           string
	snapshotPath          string
}

const storageName = "local"
const bufferSize = 1024 * 1024

func init() {
	storage.Register(storageName, &Storage{})
}

// StoreOptions ...
type StoreOptions struct {
	reportProgress bool
}

// Setup local storage
func (s *Storage) Setup(fileStorageProgressCh chan types.BackupProgress) bool {
	config := conf.ProviderConf(storageName)
	if config.Active {
		s.name = storageName
		s.eventCh = make(chan notification.Event)
		s.fileStorageProgressCh = fileStorageProgressCh
		storagePath := filepath.Clean(config.Path)
		s.storagePath = storagePath
		return true
	}
	return false
}

// Store stores a file to a local storage
func (s *Storage) Store(event *notification.Event) {
	absolutePath := event.AbsolutePath
	relativePath := event.RelativePath
	directoryPath := event.DirectoryPath

	fmt.Printf("\nlocal.Store():\n")
	fmt.Printf(">\tabsolutePath:\t%s\n>\trelativePath:\t%s\n>\tdirectoryPath:\t%s\n\n", absolutePath, relativePath, directoryPath)

	from := absolutePath
	to := filepath.Join(s.storagePath, filepath.Base(directoryPath), relativePath)
	s.store(from, to, StoreOptions{reportProgress: false})
}
