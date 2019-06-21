package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/storage"
	"github.com/glower/bakku-app/pkg/message"
	"github.com/glower/bakku-app/pkg/types"
	configstorage "github.com/glower/bakku-app/pkg/config/storage"

	"github.com/glower/file-watcher/watcher"
	"github.com/glower/file-watcher/notification"
	fi "github.com/glower/file-watcher/util"
)

// Snapshot ...
type Snapshot struct {
	ctx context.Context

	watcher   *watcher.Watch
	storage   storage.Storage
	MessageCh chan message.Message
}

// Setup the snapshot storage
func Setup(ctx context.Context, res types.GlobalResources) *Snapshot {

	snapShot := &Snapshot{
		watcher: res.FileWatcher,
		storage: res.Storage,
		MessageCh: res.MessageCh,
	}
	return snapShot
}

// func (s *Snapshot) processFileBackupComplete() {
// 	for {
// 		select {
// 		case <-s.ctx.Done():
// 			return
// 		case fileBackup := <-s.FileBackupCompleteCh:
// 			go s.fileBackupComplete(fileBackup)
// 		}
// 	}
// }

// func (s *Snapshot) fileBackupComplete(fileBackup types.FileBackupComplete) {
// 	log.Printf("snapshot.fileBackupComplete(): file [%s] is backuped to [%s]\n", fileBackup.AbsolutePath, fileBackup.BackupStorageName)
// 	if strings.Contains(fileBackup.AbsolutePath, s.storage.FileName()) {
// 		return
// 	}

// 	fileEntry, err := s.generateFileEntry(fileBackup.AbsolutePath, nil)
// 	if err != nil {
// 		s.MessageCh <- message.FormatMessage("ERROR", err.Error(), "snapshot.fileBackupComplete")
// 		return
// 	}

// 	err = s.updateFileSnapshot(fileBackup.BackupStorageName, fileEntry)
// 	if err != nil {
// 		s.MessageCh <- message.FormatMessage("ERROR", err.Error(), "snapshot.fileBackupComplete")
// 		return
// 	}

// 	// backup the snapshot file
// 	backupFileEntry, err := s.generateFileEntry(s.storage.FilePath(), nil)
// 	if err != nil {
// 		s.MessageCh <- message.FormatMessage("ERROR", err.Error(), "snapshot.fileBackupComplete")
// 		return
// 	}
// 	s.EventCh <- *backupFileEntry
// }

// func (s *Snapshot) createOrUpdate(path string) error {
// 	log.Printf("snapshot.Create(): path=%s\n", path)

// 	return filepath.Walk(path, func(file string, fileInfo os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if strings.Contains(file, s.storage.FileName()) {
// 			return nil
// 		}
// 		if !fileInfo.IsDir() {
// 			xFileInfo := fi.ExtendedFileInformation(file, fileInfo)
// 			fileEntry, err := s.generateFileEntry(file, xFileInfo)
// 			if err != nil {
// 				return err
// 			}
// 			s.EventCh <- *fileEntry
// 		}
// 		return nil
// 	})
// }

func (s *Snapshot) createOrUpdate(path string) error {
	log.Printf("snapshot.update(): path=%s\n", path)

	// read all supported backup storages form the config
	backupStorages, err := configstorage.Active()
	if err != nil {
		return err
	}

	return filepath.Walk(path, func(absoluteFilePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			backupToStorages := []string{}
			for _, backupStorage := range backupStorages {
				if s.fileDifferentToBackup(backupStorage, absoluteFilePath) {
					backupToStorages = append(backupToStorages, backupStorage)
				}
			}
			if len(backupToStorages) > 0 {
				relativePath, err := filepath.Rel(path, absoluteFilePath)
				if err != nil {
					return err
				}
				s.watcher.CreateFileAddedNotification(path, relativePath)
			}

		}
		return nil
	})
}

func (s *Snapshot) fileDifferentToBackup(backupStorageName, absoluteFilePath string) bool {
	snapshotEntryJSON, err := s.storage.Get(absoluteFilePath, backupStorageName)
	if err != nil {
		return true
	}
	if snapshotEntryJSON == "" {
		return true
	}

	e := &notification.Event{}
	decoder := json.NewDecoder(snapshotEntryJSON)
	err = decoder.Decode(&e)
	if err != nil {
		fmt.Printf("unable to unmarshal config: %v", err)
		return true
	}
	checksum := fi.FileInformation.Checksum(absoluteFilePath)
	if e.Checksum != checksum {
		return true
	}
	return false
}
