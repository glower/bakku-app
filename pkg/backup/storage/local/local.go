package local

import (
	"context"
	"log"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/backup/storage"
	conf "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/types"
)

// Storage local
type Storage struct {
	name                          string // storage name
	fileChangeNotificationChannel chan types.FileChangeNotification
	fileStorageProgressCannel     chan types.BackupProgress
	ctx                           context.Context
	storagePath                   string
	snapshotPath                  string
}

const storageName = "local"
const bufferSize = 1024 * 1024

func init() {
	log.Println("local.init()")
	storage.Register(storageName, &Storage{})
}

// StoreOptions ...
type StoreOptions struct {
	reportProgress bool
}

// Setup local storage
func (s *Storage) Setup(fileStorageProgressCannel chan types.BackupProgress) bool {
	config := conf.ProviderConf(storageName)
	if config.Active {
		s.name = storageName
		s.fileChangeNotificationChannel = make(chan types.FileChangeNotification)
		s.fileStorageProgressCannel = fileStorageProgressCannel
		storagePath := filepath.Clean(config.Path)
		s.storagePath = storagePath
		return true
	}
	return false
}

// Store ...
func (s *Storage) Store(fileChange *types.FileChangeNotification) {
	log.Printf("Store(): %#v\n", fileChange)
	absolutePath := fileChange.AbsolutePath
	relativePath := fileChange.RelativePath
	directoryPath := fileChange.DirectoryPath

	from := absolutePath
	to := filepath.Join(s.storagePath, filepath.Base(directoryPath), relativePath)
	s.store(from, to, StoreOptions{reportProgress: false})
}
