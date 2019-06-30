package snapshot

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	configstorage "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/message"
	"github.com/glower/bakku-app/pkg/storage"
	"github.com/glower/bakku-app/pkg/types"

	"github.com/glower/file-watcher/notification"
	fi "github.com/glower/file-watcher/util"
	"github.com/glower/file-watcher/watcher"
)

// Snapshot ...
type Snapshot struct {
	ctx context.Context

	watcher   *watcher.Watch
	storage   storage.Storager
	messageCh chan message.Message
}

// Setup the snapshot storage
func Setup(ctx context.Context, res *types.GlobalResources) *Snapshot {
	snapShot := &Snapshot{
		watcher:   res.FileWatcher,
		storage:   res.Storage,
		messageCh: res.MessageCh,
	}
	return snapShot
}

// CreateOrUpdate checks if files in a given dir was not backuped
func (s *Snapshot) CreateOrUpdate(path string) error {
	log.Printf("[INFO] snapshot.update(): path=%s\n", path)

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

			for _, backupStorage := range backupStorages {
				if s.fileDifferentToBackup(backupStorage, absoluteFilePath) {
					// fmt.Printf("Send [%s] to [%s]\n", absoluteFilePath, backupStorage)
					relativePath, err := filepath.Rel(path, absoluteFilePath)
					if err != nil {
						return err
					}
					s.watcher.CreateFileAddedNotification(path, relativePath, &notification.MetaInfo{"storage": backupStorage})
				}
			}

		}
		return nil
	})
}

func (s *Snapshot) fileDifferentToBackup(backupStorageName, absoluteFilePath string) bool {
	snapshotEntry, err := s.storage.Get(absoluteFilePath, backupStorageName)
	if err != nil {
		// fmt.Printf("fileDifferentToBackup(): s.storage.Get err: %v\n", err)
		return true
	}
	if snapshotEntry == "" {
		// fmt.Printf("fileDifferentToBackup(): s.storage.Get empty data\n")
		return true
	}

	e := &notification.Event{}
	err = json.Unmarshal([]byte(snapshotEntry), e)
	if err != nil {
		// fmt.Printf("fileDifferentToBackup(): unable to unmarshal file data [%q]: %v", snapshotEntry, err)
		return true
	}
	fileInfo, err := fi.GetFileInformation(absoluteFilePath)
	if err != nil {
		// fmt.Printf("fileDifferentToBackup(): unable to get [%s] info: %v", absoluteFilePath, err)
		return true
	}
	checksum, err := fileInfo.Checksum()
	if err != nil {
		// fmt.Printf("fileDifferentToBackup(): unable to get [%s] checksum: %v", absoluteFilePath, err)
		return true
	}
	if e.Checksum != checksum {
		// fmt.Printf("fileDifferentToBackup(): Old checksum: [%s] checksum: [%s]\n", e.Checksum, checksum)
		return true
	}
	return false
}
