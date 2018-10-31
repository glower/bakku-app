package local

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/backup/storage"
	conf "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/otiai10/copy"
	"github.com/spf13/viper"
)

// Storage local
type Storage struct {
	name                          string // storage name
	fileChangeNotificationChannel chan *types.FileChangeNotification
	fileStorageProgressCannel     chan *storage.Progress
	ctx                           context.Context
	storagePath                   string
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
func (s *Storage) Setup(fileStorageProgressCannel chan *storage.Progress) bool {
	config := conf.ProviderConf(storageName)
	if config.Active {
		log.Println("storage.local.Setup()")
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
	log.Println("storage.local.SyncLocalFilesToBackup()")
	// TODO: move this to some utils or config class, so we don't work with viper direct
	dirs, ok := viper.Get("watch").([]interface{})
	if !ok {
		log.Println("[ERROR] storage.local.SyncLocalFilesToBackup(): nothing to sync")
		return
	}

	for _, path := range dirs {
		path, ok := path.(string)
		if !ok {
			log.Println("[ERROR] storage.local.SyncLocalFilesToBackup(): invalid path")
			continue
		}

		remoteSnapshotPath := filepath.Join(s.storagePath, filepath.Base(path), snapshot.Dir())
		localTMPPath := filepath.Join(os.TempDir(), snapshot.AppName(), storageName, filepath.Base(path))

		log.Printf("storage.local.SyncLocalFilesToBackup(): copy snapshot for [%s] from [%s] to [%s]\n",
			path, remoteSnapshotPath, localTMPPath)

		if err := copy.Copy(remoteSnapshotPath, localTMPPath); err != nil {
			log.Printf("[ERROR] storage.local.SyncLocalFilesToBackup(): cannot copy snapshot for [%s]: %v\n", path, err)
			return
		}

		snapshotPath := filepath.Join(path, snapshot.Dir())
		s.syncFiles(localTMPPath, snapshotPath)
	}
}

// FileChangeNotification returns channel for notifications
func (s *Storage) FileChangeNotification() chan *types.FileChangeNotification {
	return s.fileChangeNotificationChannel
}

// Start local storage
func (s *Storage) Start(ctx context.Context) error {
	log.Println("storage.local.Start()")
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
	// log.Printf("storage.local.handleFileChanges(): File [%#v] has been changed\n", fileChange)
	absolutePath := fileChange.AbsolutePath
	relativePath := fileChange.RelativePath
	directoryPath := fileChange.DirectoryPath

	snapshotPath := snapshot.Path(directoryPath)
	from := absolutePath
	to := filepath.Join(s.storagePath, filepath.Base(directoryPath), relativePath)
	remoteSnapshot := filepath.Join(s.storagePath, filepath.Base(directoryPath), snapshot.Dir())

	// don't backup file if it is in progress
	if ok := storage.BackupStarted(absolutePath, storageName); ok {
		s.store(from, to, StoreOptions{reportProgress: true})
		storage.BackupFinished(absolutePath, storageName)
		storage.UpdateSnapshot(snapshotPath, directoryPath, absolutePath)
		s.SyncSnapshot(snapshotPath, remoteSnapshot)
	}
}
