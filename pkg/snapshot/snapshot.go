package snapshot

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glower/bakku-app/pkg/snapshot/storage"
	"github.com/glower/bakku-app/pkg/snapshot/storage/boltdb"
	"github.com/glower/bakku-app/pkg/types"
)

// Snapshot ...
type Snapshot struct {
	backupStorages  []string
	snapshotStorage storage.Storage
	path            string
	changesChan     chan types.FileChangeNotification
	changesDoneChan chan bool
}

// New creates new channals for file notifications
func New(path string, storages []string) *Snapshot {
	bolt := boltdb.New(path)
	storage.Register(path, bolt)

	snap := &Snapshot{}
	snap.path = path
	snap.backupStorages = storages
	snap.snapshotStorage = bolt
	snap.changesChan = make(chan types.FileChangeNotification)
	snap.changesDoneChan = make(chan bool)

	return snap
}

// CreateOrUpdate a new or update an existing snapshot entry for a given directory path
func (s *Snapshot) CreateOrUpdate() {
	log.Printf("snapshot.CreateOrUpdate(): path=%s storages=%v\n", s.path, s.backupStorages)
	firstTimeBackup := false
	// TODO: check snapshot for the path and storage name
	if s.snapshotStorage.Exist() {
		firstTimeBackup = true
	}

	filepath.Walk(s.path, func(file string, fileInfo os.FileInfo, err error) error {
		if strings.Contains(file, s.snapshotStorage.FileName()) {
			return nil
		}
		if !fileInfo.IsDir() {
			entry, err := s.generateFileEntry(file, fileInfo)
			if firstTimeBackup && err == nil {
				s.changesChan <- *entry
				return nil
			}
			if err == nil {
				for _, storageName := range s.backupStorages {
					new, err := s.isFileDifferentToBackup(storageName, entry)
					if err == nil && new {
						log.Printf(" File [%s] is new or different to the copy in [%s] storage\n", file, storageName)
						s.changesChan <- *entry
						return nil
					}
					if err != nil {
						log.Printf("[ERROR] CreateOrUpdate(): %v\n", err)
						return err
					}
				}
			}

		}
		return nil
	})

	if !firstTimeBackup {
		s.changesDoneChan <- true
	}
}

// UpdateEntry ...
func UpdateEntry(fileChange *types.FileChangeNotification, backupStorageName string) {
	absolutePath := fileChange.AbsolutePath
	relativePath := fileChange.RelativePath
	snapshotPath := fileChange.DirectoryPath

	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		log.Printf("[ERROR] storage.UpdateEntry(): can't stat file [%s]: %v\n", absolutePath, err)
		return
	}

	entry, err := s.generateFileEntry(absolutePath, fileInfo)
	if err != nil {
		log.Printf("[ERROR] storage.UpdateEntry(): snapshotPath:[%s], filePath:[%s], error=%v\n", snapshotPath, relativePath, err)
		return
	}

	err = s.updateEntry(backupStorageName, entry)
	if err != nil {
		log.Printf("[ERROR] storage.UpdateEntry(): can't update file entry file [%s]: %v\n", absolutePath, err)
		return
	}
}

func (s *Snapshot) isFileDifferentToBackup(backupStorageName string, entry *types.FileChangeNotification) (bool, error) {
	log.Printf("isFileDifferentToBackup(): backupStorageName=[%s]\n", backupStorageName)
	snapshotEntryJSON, err := s.snapshotStorage.Get(entry.AbsolutePath, backupStorageName)
	if err != nil {
		return true, err
	}
	if snapshotEntryJSON == "" {
		return true, nil
	}

	entryJSON, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("[ERROR] isFileDifferentToBackup(): Marshal error: %v\n", err)
		return true, err
	}

	// Maybe it is better to compair the object and not the JSON string
	if string(entryJSON) != snapshotEntryJSON {
		return false, nil
	}

	return false, nil
}

func (s *Snapshot) updateEntry(backupStorageName string, entry *types.FileChangeNotification) error {
	value, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	err = s.snapshotStorage.Add(entry.AbsolutePath, backupStorageName, value)
	if err != nil {
		return err
	}
	return nil
}

// filePath must be absulute path
func (s *Snapshot) generateFileEntry(filePath string, fileInfo os.FileInfo) (*types.FileChangeNotification, error) {
	log.Printf("snapshot.generateFileEntry(): snapshotPath=%s, filePath=%s\n", s.path, filePath)
	// TODO: check if filePath and snapshotPath are absolute!
	host, _ := os.Hostname() // TODO: handle this error
	fileName := filepath.Base(filePath)
	relativePath := strings.Replace(filePath, s.path+string(os.PathSeparator), "", -1)
	snapshot := types.FileChangeNotification{
		// TODO: add mime type here!
		AbsolutePath:       filePath,
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
