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
	fileStorageProgressCannel     chan *storage.Progress
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
func (s *Storage) Setup(fileStorageProgressCannel chan *storage.Progress) bool {
	log.Println("local.Setup()")
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

// FileChangeNotification returns channel for notifications
func (s *Storage) FileChangeNotification() chan *types.FileChangeNotification {
	return s.fileChangeNotificationChannel
}

// Start local storage
func (s *Storage) Start(ctx context.Context) error {
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

func (s *Storage) handleFileChanges(fileChange *types.FileChangeNotification) {
	log.Printf("local.handleFileChanges(): File [%s] has been changed\n", fileChange.AbsolutePath)
	// This should't happen, but anyway
	if fileChange.Action == types.FileRemoved {
		log.Printf("local.handleFileChanges(): file is removed from the local storage, ignore this for now\n")
		return
	}
	absolutePath := fileChange.AbsolutePath   // /foo/bar/buz/alice.jpg
	relativePath := fileChange.RelativePath   // buz/alice.jpg
	directoryPath := fileChange.DirectoryPath // /foo/bar/

	from := absolutePath
	to := filepath.Join(s.storagePath, filepath.Base(directoryPath), relativePath)

	// don't backup file if it is in progress
	if ok := storage.BackupStarted(from, storageName); ok {
		s.store(from, to, StoreOptions{reportProgress: true})
		storage.BackupFinished(absolutePath, storageName)

		// TODO: first time sync?
		snapshot.UpdateEntry(directoryPath, relativePath, storageName)
		remoteSnapshotPath := filepath.Join(s.storagePath, fileChange.WatchDirectoryName)
		s.SyncSnapshot(directoryPath, remoteSnapshotPath)
	}
}
