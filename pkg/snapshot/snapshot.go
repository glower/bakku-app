package snapshot

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glower/bakku-app/pkg/snapshot/storage"
	// leveldb is default storage implementation for the snapshot
	_ "github.com/glower/bakku-app/pkg/snapshot/storage/leveldb"
	"github.com/glower/bakku-app/pkg/types"
)

// Snapshot ...
func Snapshot(path string) storage.Snapshot {
	return storage.GetDefault().Path(path)
}

// Dir ...
func Dir() string {
	return storage.GetDefault().SnapshotStoragePathName()
}

// StoragePath returns path for a snapshot for a given directory
func StoragePath(path string) string {
	return Snapshot(path).SnapshotStoragePath()
}

// Exist ...
func Exist(path string) bool {
	return Snapshot(path).Exist()
}

// CreateOrUpdate a new or update an existing snapshot entry for a given directory path
func CreateOrUpdate(snapshotPath string, callbackChan chan<- types.FileChangeNotification, changesDoneChan chan<- bool) {
	log.Printf("snapshot.CreateOrUpdate(): path=%s\n", snapshotPath)
	firstTimeBackup := false
	if !Exist(snapshotPath) {
		firstTimeBackup = true
	}
	filepath.Walk(snapshotPath, func(file string, fileInfo os.FileInfo, err error) error {
		// if snapshotPath
		if !fileInfo.IsDir() {
			entry, err := updateEntry(snapshotPath, file, fileInfo)
			if firstTimeBackup && err == nil {
				callbackChan <- *entry
			}
		}
		return nil
	})
	if !firstTimeBackup {
		log.Printf("CreateOrUpdate(): done with new scan for [%s], send signal ...\n", snapshotPath)
		changesDoneChan <- true
	}
	log.Println("CreateOrUpdate(): done")
}

// UpdateEntry ...
func UpdateEntry(snapshotPath, filePath string) {
	absolutePath := filepath.Join(snapshotPath, filePath)
	f, err := os.Stat(absolutePath)
	if err != nil {
		log.Printf("storage.UpdateEntry(): can not stat file [%s]: %v\n", absolutePath, err)
		return
	}
	updateEntry(snapshotPath, filePath, f)
}

func updateEntry(snapshotPath, filePath string, fileInfo os.FileInfo) (*types.FileChangeNotification, error) {
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
	// TODO: maybe move this part to the Add() function
	value, err := json.Marshal(snapshot)
	if err != nil {
		log.Printf("snapshot.Update(): cannot Marshal to json %#v: %v\n", snapshot, err)
		return nil, err
	}

	err = Snapshot(snapshotPath).Add(filePath, value)
	if err != nil {
		log.Printf("snapshot.Update(): cannot Marshal to json %#v: %v\n", snapshot, err)
		return nil, err
	}
	return &snapshot, err
}

// RemoveSnapshotEntry removed entry fron the snapshot file
// TODO: maybe we don't delete file here, but only mark the file as deleted
///      and let the user decide what to do
func RemoveSnapshotEntry(directoryPath, filePath string) {
	log.Printf("RemoveSnapshotEntry(): remove [%s] from [%s]\n", filePath, directoryPath)
	err := Snapshot(directoryPath).Remove(filePath)
	if err != nil {
		log.Printf("[ERROR] snapshot.RemoveSnapshotEntry(): cannot delete entry [%s] from [%s]: %v\n", directoryPath, filePath, err)
	}
}

// Diff returns diff between two snapshots as array of FileChangeNotification
func Diff(remoteSnapshotPath, localSnapshotPath string) (*[]types.FileChangeNotification, error) {
	log.Printf("snapshot.Diff(): remote=[%s] local=[%s]", remoteSnapshotPath, localSnapshotPath)
	var result []types.FileChangeNotification

	dbRemote, err := Snapshot(remoteSnapshotPath).GetAll()
	if err != nil {
		return nil, fmt.Errorf("snapshot.Diff(): cannot open snapshot for the path [%s]: %v", remoteSnapshotPath, err)
	}

	dbLocal, err := Snapshot(localSnapshotPath).GetAll()
	if err != nil {
		return nil, fmt.Errorf("snapshot.Diff(): can not open snapshot for the path [%s]: %v", localSnapshotPath, err)
	}

	for localFile, localInfo := range dbLocal {
		remoteInfo, found := dbRemote[localFile]
		if !found || localInfo != remoteInfo {
			log.Printf("snapshot.Diff(): key [%s] not found or different from the remote snapshot\n", localFile)
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
