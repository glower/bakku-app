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

// Exist ...
func Exist(path string) bool {
	return storage.GetByPath(path).Exist()
}

// Snapshot ...
func Snapshot(path string) storage.Storager {
	return storage.GetByPath(path)
}

// FileName ...
func FileName(path string) string {
	return storage.GetByPath(path).FileName()
}

// FilePath ...
func FilePath(path string) string {
	return storage.GetByPath(path).FilePath()
}

// New creates new channals for file notifications
func New(path string, storages []string) *types.Notifications {
	snapshot := boltdb.New(path)
	storage.Register(path, snapshot)
	changesChan := make(chan types.FileChangeNotification)
	changesDoneChan := make(chan bool)

	go CreateOrUpdate(path, changesChan, changesDoneChan)

	return &types.Notifications{
		FileChangeChan: changesChan,
		DoneChan:       changesDoneChan,
	}
}

// CreateOrUpdate a new or update an existing snapshot entry for a given directory path
func CreateOrUpdate(snapshotPath string, fileChangeChan chan<- types.FileChangeNotification, changesDoneChan chan<- bool) {
	log.Printf("snapshot.CreateOrUpdate(): path=%s\n", snapshotPath)

	firstTimeBackup := false
	if !Exist(snapshotPath) {
		firstTimeBackup = true
	}
	filepath.Walk(snapshotPath, func(file string, fileInfo os.FileInfo, err error) error {
		if !fileInfo.IsDir() {
			entry, err := generateFileEntry(snapshotPath, file, fileInfo)
			if firstTimeBackup && err == nil {
				log.Printf(">>> snapshot.CreateOrUpdate(): first backup for: %v\n", entry)
				fileChangeChan <- *entry
			}
		}
		return nil
	})
	if !firstTimeBackup {
		changesDoneChan <- true
	}
}

// func fileChanged(snapshotPath, filePath string, fileInfo os.FileInfo) bool {
// 	err := Snapshot(snapshotPath).Add(filePath, value)
// 	if err != nil {
// 		return nil, err
// 	}
// }

// UpdateEntry ...
func UpdateEntry(fileChange *types.FileChangeNotification, storageName string) {
	// absolutePath := filepath.Join(snapshotPath, filePath)
	absolutePath := fileChange.AbsolutePath  // /foo/bar/buz/alice.jpg
	relativePath := fileChange.RelativePath  // buz/alice.jpg
	snapshotPath := fileChange.DirectoryPath // /foo/bar/

	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		log.Printf("[ERROR] storage.UpdateEntry(): can't stat file [%s]: %v\n", absolutePath, err)
		return
	}
	entry, err := generateFileEntry(snapshotPath, relativePath, fileInfo)
	if err != nil {
		log.Printf("[ERROR] storage.UpdateEntry(): snapshotPath:[%s], filePath:[%s], error=%v\n", snapshotPath, relativePath, err)
		return
	}
	err = updateEntry(snapshotPath, storageName, entry)
	if err != nil {
		log.Printf("[ERROR] storage.UpdateEntry(): can't update file entry file [%s]: %v\n", absolutePath, err)
		return
	}
}

func updateEntry(snapshotPath, storageName string, entry *types.FileChangeNotification) error {
	value, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// func (s *Storage) Add(filePath, bucketName string, value []byte) error {
	err = Snapshot(snapshotPath).Add(entry.AbsolutePath, storageName, value)
	if err != nil {
		return err
	}
	return nil
}

func generateFileEntry(snapshotPath, filePath string, fileInfo os.FileInfo) (*types.FileChangeNotification, error) {
	// log.Printf("snapshot.updateEntry(): snapshotPath=%s, filePath=%s\n", snapshotPath, filePath)
	host, _ := os.Hostname() // TODO: handle this error
	fileName := filepath.Base(filePath)
	relativePath := strings.Replace(filePath, snapshotPath+string(os.PathSeparator), "", -1)
	snapshot := types.FileChangeNotification{
		AbsolutePath:       filePath,
		Action:             types.Action(types.FileAdded),
		DirectoryPath:      snapshotPath,
		Machine:            host,
		Name:               fileName,
		RelativePath:       relativePath,
		Size:               fileInfo.Size(),
		Timestamp:          fileInfo.ModTime(),
		WatchDirectoryName: filepath.Base(snapshotPath),
	}

	return &snapshot, nil
}

// RemoveSnapshotEntry removed entry fron the snapshot file
// TODO: maybe we don't delete file here, but only mark the file as deleted
///      and let the user decide what to do
func RemoveSnapshotEntry(directoryPath, filePath string) {
	log.Printf("RemoveSnapshotEntry(): remove [%s] from [%s]\n", filePath, directoryPath)
	// TODO: fix me!
	// err := Snapshot(directoryPath).Remove(filePath, "default")
	// if err != nil {
	// 	log.Printf("[ERROR] snapshot.RemoveSnapshotEntry(): cannot delete entry [%s] from [%s]: %v\n", directoryPath, filePath, err)
	// }
}

// Diff returns diff between two snapshots as array of FileChangeNotification
func Diff(remoteSnapshotPath, localSnapshotPath string) (*[]types.FileChangeNotification, error) {
	log.Printf("snapshot.Diff(): remote snapshot:[%s] local snapshot: [%s]", remoteSnapshotPath, localSnapshotPath)
	var result []types.FileChangeNotification

	dbRemote, err := boltdb.New(remoteSnapshotPath).GetAll("TODO")
	if err != nil {
		return nil, fmt.Errorf("snapshot.Diff(): can't open snapshot for the path [%s]: %v", remoteSnapshotPath, err)
	}

	dbLocal, err := Snapshot(localSnapshotPath).GetAll("TODO")
	if err != nil {
		return nil, fmt.Errorf("snapshot.Diff(): can't open snapshot for the path [%s]: %v", localSnapshotPath, err)
	}

	for localFile, localInfo := range dbLocal {
		remoteInfo, found := dbRemote[localFile]
		if !found || localInfo != remoteInfo {

			if !found {
				log.Printf("snapshot.Diff(): file [%s] is not found in the remote snapshot\n", localFile)
			} else {
				log.Println("------------------------------------- diff -------------------------------------")
				log.Printf("local:  %s\n", localInfo)
				log.Printf("repote: %s\n", remoteInfo)
				log.Println("------------------------------------- diff -------------------------------------")
			}

			file, err := unmurshalFileChangeNotification(localInfo)
			if err != nil {
				log.Printf("[ERROR] snapshot.Diff(): %s\n", err)
				continue
			}
			result = append(result, file)
		}
		if err != nil {
			log.Printf("[ERROR] snapshot.Diff(): can not get key=[%s]: dbRemote.Get(): %v\n", string(localFile), err)
		}
	}
	return &result, nil
}

func unmurshalFileChangeNotification(value string) (types.FileChangeNotification, error) {
	change := types.FileChangeNotification{}
	if err := json.Unmarshal([]byte(value), &change); err != nil {
		return change, fmt.Errorf("cannot unmarshal data [%s]: %v", string(value), err)
	}
	return change, nil
}
