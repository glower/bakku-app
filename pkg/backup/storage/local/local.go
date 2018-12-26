package local

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/backup"
	"github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/glower/bakku-app/pkg/config"
	conf "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/otiai10/copy"
)

// Storage local
type Storage struct {
	name                          string // storage name
	fileChangeNotificationChannel chan *types.FileChangeNotification
	fileStorageProgressCannel     chan *backup.Progress
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
func (s *Storage) Setup(fileStorageProgressCannel chan *backup.Progress) bool {
	config := conf.ProviderConf(storageName)
	if config.Active {
		s.name = storageName
		s.fileChangeNotificationChannel = make(chan *types.FileChangeNotification)
		s.fileStorageProgressCannel = fileStorageProgressCannel
		storagePath := filepath.Clean(config.Path)
		s.storagePath = storagePath
		return true
	}
	return false
}

// SyncLocalFilesToBackup ...
func (s *Storage) SyncLocalFilesToBackup() {
	log.Println("local.SyncLocalFilesToBackup(): START")
	dirs := config.DirectoriesToWatch()
	for _, path := range dirs {
		log.Printf("local.SyncLocalFilesToBackup(): %s\n", path)

		remoteSnapshot := filepath.Join(s.storagePath, filepath.Base(path), snapshot.FileName(path))

		localTMPPath := filepath.Join(os.TempDir(), backup.DefultFolderName(), storageName, filepath.Base(path))
		localTMPFile := filepath.Join(localTMPPath, snapshot.FileName(path))

		log.Printf("local.SyncLocalFilesToBackup(): copy snapshot for [%s] from [%s] to [%s]\n",
			path, remoteSnapshot, localTMPFile)

		if err := copy.Copy(remoteSnapshot, localTMPFile); err != nil {
			log.Printf("[ERROR] local.SyncLocalFilesToBackup(): can't copy snapshot for [%s]: %v\n", path, err)
			return
		}

		s.syncFiles(localTMPPath, path)
	}
}

// Store ...
func (s *Storage) Store(fileChange *types.FileChangeNotification) {
	absolutePath := fileChange.AbsolutePath
	relativePath := fileChange.RelativePath
	directoryPath := fileChange.DirectoryPath

	from := absolutePath
	to := filepath.Join(s.storagePath, filepath.Base(directoryPath), relativePath)
	s.store(from, to, StoreOptions{reportProgress: true})
}
