package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glower/bakku-app/pkg/config"
	storageconfig "github.com/glower/bakku-app/pkg/config/storage"
	fi "github.com/glower/bakku-app/pkg/file"
	snapshotstorage "github.com/glower/bakku-app/pkg/snapshot/storage"
	"github.com/glower/bakku-app/pkg/snapshot/storage/boltdb"
	"github.com/glower/bakku-app/pkg/types"
)

// Snapshot ...
type Snapshot struct {
	ctx context.Context

	path                          string
	storage                       snapshotstorage.Storage
	FileChangeNotificationChannel chan types.FileChangeNotification
	FileBackupCompleteChannel     chan types.FileBackupComplete
}

// Setup the snapshot storage
func Setup(ctx context.Context, fileChangeNotificationChan chan types.FileChangeNotification, fileBackupCompleteChan chan types.FileBackupComplete) {
	dirs := config.DirectoriesToWatch()
	for _, path := range dirs {

		snap := &Snapshot{
			ctx:                           ctx,
			path:                          path,
			FileChangeNotificationChannel: fileChangeNotificationChan,
			FileBackupCompleteChannel:     fileBackupCompleteChan,
		}
		bolt := boltdb.New(path)
		err := snapshotstorage.Register(bolt)
		if err != nil {
			log.Panicf("[PANIC] snapshot.Setup(): %v\n", err)
		}

		snap.storage = bolt
		go snap.processFileBackupComplete()

		if !bolt.Exist() {
			snap.create()
		} else {
			snap.update()
		}
	}
}

func (s *Snapshot) processFileBackupComplete() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case fileBackup := <-s.FileBackupCompleteChannel:
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
		log.Printf("[ERROR] snapshot.fileBackupComplete(): %v\n", err)
		return
	}

	err = s.updateFileSnapshot(fileBackup.BackupStorageName, fileEntry)
	if err != nil {
		log.Printf("[ERROR] snapshot.fileBackupComplete(): %v\n", err)
		return
	}

	// backup the snapshot file
	backupFileEntry, err := s.generateFileEntry(s.storage.FilePath(), nil)
	if err != nil {
		log.Printf("[ERROR] snapshot.fileBackupComplete(): %v\n", err)
		return
	}

	s.FileChangeNotificationChannel <- *backupFileEntry
}

func (s *Snapshot) create() {
	log.Printf("snapshot.Create(): path=%s\n", s.path)

	filepath.Walk(s.path, func(file string, fileInfo os.FileInfo, err error) error {
		if strings.Contains(file, s.storage.FileName()) {
			return nil
		}
		if !fileInfo.IsDir() {
			xFileInfo := fi.ExtendedFileInformation(file, fileInfo)
			fileEntry, err := s.generateFileEntry(file, xFileInfo)
			if err != nil {
				log.Printf("[ERROR] Create(): %v\n", err)
				return err
			}
			s.FileChangeNotificationChannel <- *fileEntry
		}
		return nil
	})
}

func (s *Snapshot) update() {
	log.Printf("snapshot.update(): path=%s\n", s.path)

	// read all supported backup storages form the config
	backupStorages, err := storageconfig.Active()
	if err != nil {
		log.Panic(err)
	}

	filepath.Walk(s.path, func(filePath string, fileInfo os.FileInfo, err error) error {
		if strings.Contains(filePath, s.storage.FileName()) {
			return nil
		}
		if !fileInfo.IsDir() {
			xFileInfo := fi.ExtendedFileInformation(filePath, fileInfo)
			fileEntry, err := s.generateFileEntry(filePath, xFileInfo)
			if err != nil {
				log.Printf("[ERROR] Update(): %v\n", err)
				return err
			}
			backupToStorages := []string{}
			for _, backupStorage := range backupStorages {
				if s.fileDifferentToBackup(backupStorage, fileEntry) {
					log.Printf("snapshot.update(): local file [%s] is different to the remote copy in [%s] storage", fileEntry.AbsolutePath, backupStorage)
					backupToStorages = append(backupToStorages, backupStorage)
				}
			}
			if len(backupToStorages) > 0 {
				fileEntry.BackupToStorages = backupToStorages
				s.FileChangeNotificationChannel <- *fileEntry
			}

		}
		return nil
	})
}

func (s *Snapshot) updateFileSnapshot(backupStorageName string, entry *types.FileChangeNotification) error {
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

func (s *Snapshot) fileDifferentToBackup(backupStorageName string, entry *types.FileChangeNotification) bool {
	snapshotEntryJSON, err := s.storage.Get(entry.AbsolutePath, backupStorageName)
	if err != nil {
		log.Printf("[ERROR] fileDifferentToBackup(): %v", err)
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

func (s *Snapshot) generateFileEntry(absoluteFilePath string, fileInfo fi.ExtendedFileInfoImplementer) (*types.FileChangeNotification, error) {
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
		log.Printf("snapshot.generateFileEntry(): can't get host name: %v\n", err)
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

	snapshot := types.FileChangeNotification{
		MimeType:           mimeType,
		AbsolutePath:       absoluteFilePath,
		Action:             types.Action(types.FileAdded),
		DirectoryPath:      s.path,
		Machine:            host,
		Name:               fileName,
		RelativePath:       relativePath,
		Size:               fileInfo.Size(),
		Timestamp:          fileInfo.ModTime(),
		WatchDirectoryName: filepath.Base(s.path),
	}

	return &snapshot, nil
}

func unmurshalFileChangeNotification(value string) (types.FileChangeNotification, error) {
	change := types.FileChangeNotification{}
	if err := json.Unmarshal([]byte(value), &change); err != nil {
		return change, fmt.Errorf("cannot unmarshal data [%s]: %v", string(value), err)
	}
	return change, nil
}
