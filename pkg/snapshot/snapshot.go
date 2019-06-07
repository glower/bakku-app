package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	storageconfig "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/message"
	snapshotstorage "github.com/glower/bakku-app/pkg/snapshot/storage"
	"github.com/glower/bakku-app/pkg/snapshot/storage/boltdb"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/google/uuid"

	"github.com/glower/file-watcher/notification"
	fi "github.com/glower/file-watcher/util"
)

// Snapshot ...
type Snapshot struct {
	ctx context.Context

	path      string
	storage   snapshotstorage.Storage
	EventCh   chan notification.Event
	MessageCh chan message.Message
	*SnapshotManger
}

type SnapshotManger struct {
	FileBackupCompleteCh chan types.FileBackupComplete
}

// Setup the snapshot storage
func Setup(ctx context.Context, dirsToWatch []string, eventCh chan notification.Event, messageCh chan message.Message) *SnapshotManger {
	sm := &SnapshotManger{
		FileBackupCompleteCh: make(chan types.FileBackupComplete),
	}
	for _, path := range dirsToWatch {
		var err error
		snap := &Snapshot{
			ctx:            ctx,
			path:           path,
			MessageCh:      messageCh,
			EventCh:        eventCh,
			SnapshotManger: sm,
		}
		bolt := boltdb.New(path)
		err = snapshotstorage.Register(bolt)
		if err != nil {
			fmt.Printf("snapshot.Setup(): PANIC %v\n", err)
			snap.MessageCh <- message.FormatMessage("PANIC", err.Error(), "snapshot")
		}

		snap.storage = bolt
		go snap.processFileBackupComplete()

		if !bolt.Exist() {
			err = snap.create()
		} else {
			err = snap.update()
		}

		if err != nil {
			snap.MessageCh <- message.FormatMessage("PANIC", err.Error(), "snapshot")
		}
	}
	return sm
}

func (s *Snapshot) processFileBackupComplete() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case fileBackup := <-s.FileBackupCompleteCh:
			go s.fileBackupComplete(fileBackup)
		}
	}
}

func (s *Snapshot) fileBackupComplete(fileBackup types.FileBackupComplete) {
	log.Printf("snapshot.fileBackupComplete(): file [%s] is backuped to [%s]\n", fileBackup.AbsolutePath, fileBackup.BackupStorageName)
	if strings.Contains(fileBackup.AbsolutePath, s.storage.FileName()) {
		return
	}

	fileEntry, err := s.generateFileEntry(fileBackup.AbsolutePath, nil)
	if err != nil {
		s.MessageCh <- message.FormatMessage("ERROR", err.Error(), "snapshot.fileBackupComplete")
		return
	}

	err = s.updateFileSnapshot(fileBackup.BackupStorageName, fileEntry)
	if err != nil {
		s.MessageCh <- message.FormatMessage("ERROR", err.Error(), "snapshot.fileBackupComplete")
		return
	}

	// backup the snapshot file
	backupFileEntry, err := s.generateFileEntry(s.storage.FilePath(), nil)
	if err != nil {
		s.MessageCh <- message.FormatMessage("ERROR", err.Error(), "snapshot.fileBackupComplete")
		return
	}
	s.EventCh <- *backupFileEntry
}

func (s *Snapshot) create() error {
	log.Printf("snapshot.Create(): path=%s\n", s.path)

	return filepath.Walk(s.path, func(file string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(file, s.storage.FileName()) {
			return nil
		}
		if !fileInfo.IsDir() {
			xFileInfo := fi.ExtendedFileInformation(file, fileInfo)
			fileEntry, err := s.generateFileEntry(file, xFileInfo)
			if err != nil {
				return err
			}
			s.EventCh <- *fileEntry
		}
		return nil
	})
}

func (s *Snapshot) update() error {
	log.Printf("snapshot.update(): path=%s\n", s.path)

	// read all supported backup storages form the config
	backupStorages, err := storageconfig.Active()
	if err != nil {
		return err
	}

	return filepath.Walk(s.path, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(filePath, s.storage.FileName()) {
			return nil
		}
		if !fileInfo.IsDir() {
			xFileInfo := fi.ExtendedFileInformation(filePath, fileInfo)
			fileEntry, err := s.generateFileEntry(filePath, xFileInfo)
			if err != nil {
				return err
			}
			backupToStorages := []string{}
			for _, backupStorage := range backupStorages {
				if s.fileDifferentToBackup(backupStorage, fileEntry) {
					// log.Printf("snapshot.update(): local file [%s] is different to the remote copy in [%s] storage", fileEntry.AbsolutePath, backupStorage)
					backupToStorages = append(backupToStorages, backupStorage)
				}
			}
			if len(backupToStorages) > 0 {
				// fileEntry.BackupToStorages = backupToStorages // TODO !!!
				s.EventCh <- *fileEntry
			}

		}
		return nil
	})
}

func (s *Snapshot) updateFileSnapshot(backupStorageName string, entry *notification.Event) error {
	log.Printf("storage.updateFileSnapshot(): file=%s, storage=%s", entry.AbsolutePath, backupStorageName)
	value, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	err = s.storage.Add(entry.AbsolutePath, backupStorageName, value)
	if err != nil {
		return err
	}
	return nil
}

func (s *Snapshot) fileDifferentToBackup(backupStorageName string, entry *notification.Event) bool {
	snapshotEntryJSON, err := s.storage.Get(entry.AbsolutePath, backupStorageName)
	if err != nil {
		// log.Printf("[INFO] fileDifferentToBackup(): %v", err)
		// TODO: add ErrorBucketNotFound for "bolt.Get(): bucket [fake] not found" and check for it
		return true
	}
	if snapshotEntryJSON == "" {
		return true
	}

	entryJSON, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("[ERROR] fileDifferentToBackup(): Marshal error: %v\n", err)
		return true
	}

	// Maybe it is better to compair the object and not the JSON string
	if string(entryJSON) != snapshotEntryJSON {
		fmt.Printf("\nfile>   %s\nbackup> %s\n\n", string(entryJSON), snapshotEntryJSON)
		return true
	}

	return false
}

func (s *Snapshot) generateFileEntry(absoluteFilePath string, fileInfo fi.ExtendedFileInfoImplementer) (*notification.Event, error) {
	// log.Printf("snapshot.generateFileEntry(): snapshotPath=%s, filePath=%s\n", s.path, absoluteFilePath)

	if !filepath.IsAbs(absoluteFilePath) {
		return nil, fmt.Errorf("filepath %s is not absolute", absoluteFilePath)
	}

	var err error
	if fileInfo == nil {
		fileInfo, err = fi.GetFileInformation(absoluteFilePath)
		if err != nil {
			return nil, err
		}
	}

	host, err := os.Hostname()
	if err != nil {
		log.Printf("[ERROR] snapshot.generateFileEntry(): can't get host name: %v\n", err)
		host = "unknown"
	}

	mimeType, err := fileInfo.ContentType()
	if err != nil {
		log.Printf("[ERROR] snapshot.generateFileEntry(): can't get ContentType from the file [%s]: %v\n", absoluteFilePath, err)
		return nil, err
	}

	fileName := filepath.Base(absoluteFilePath)

	relativePath, err := filepath.Rel(s.path, absoluteFilePath)
	if err != nil {
		return nil, err
	}

	checksum, err := fileInfo.Checksum()
	if err != nil {
		return nil, err
	}

	snapshot := notification.Event{
		MimeType:           mimeType,
		AbsolutePath:       absoluteFilePath,
		Action:             notification.ActionType(notification.FileAdded),
		DirectoryPath:      s.path,
		Machine:            host,
		FileName:           fileName,
		RelativePath:       relativePath,
		Size:               fileInfo.Size(),
		Timestamp:          fileInfo.ModTime(),
		WatchDirectoryName: filepath.Base(s.path),
		UUID:               uuid.New(),
		Checksum:           checksum,
	}

	return &snapshot, nil
}
