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

// const snapshotDirName = ".snapshot"
const appName = "bakku-app"

// Dir ...
func Dir() string {
	return ""
}

// AppName ...
func AppName() string {
	return appName
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
func CreateOrUpdate(snapshotPath string) {
	log.Printf("snapshot.CreateOrUpdate(): path=%s\n", snapshotPath)
	filepath.Walk(snapshotPath, func(file string, fileInfo os.FileInfo, err error) error {
		// if strings.Contains(path, Dir()) {
		// 	return nil
		// }
		if !fileInfo.IsDir() {
			updateEntry(snapshotPath, file, fileInfo)
		}
		return nil
	})
	log.Println("UpdateSnapshot(): done")
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

func updateEntry(snapshotPath, filePath string, fileInfo os.FileInfo) {
	log.Printf("snapshot.updateEntry(): snapshotPath=%s, filePath=%s\n", snapshotPath, filePath)
	host, _ := os.Hostname() // TODO: handle this error
	fileName := filepath.Base(filePath)
	relativePath := strings.Replace(filePath, snapshotPath+string(os.PathSeparator), "", -1)
	snapshot := types.FileChangeNotification{
		AbsolutePath:  filePath,
		RelativePath:  relativePath,
		DirectoryPath: snapshotPath,
		Name:          fileName,
		Size:          fileInfo.Size(),
		Timestamp:     fileInfo.ModTime(),
		Machine:       host,
	}
	// TODO: maybe move this part to the Add() function
	value, err := json.Marshal(snapshot)
	if err != nil {
		log.Printf("snapshot.Update(): cannot Marshal to json %#v: %v\n", snapshot, err)
		return
	}

	err = Snapshot(snapshotPath).Add(filePath, value)
	if err != nil {
		log.Printf("snapshot.Update(): cannot Marshal to json %#v: %v\n", snapshot, err)
		return
	}
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

// CreateFirstBackup ...
func CreateFirstBackup(snapshotPath string, changes chan types.FileChangeNotification) {
	log.Printf("watchers.CreateFirstBackup(): for=[%s]\n", snapshotPath)
	data, err := Snapshot(snapshotPath).GetAll()
	if err != nil {
		log.Printf("[ERROR] watchers.CreateFirstBackup(): can not open snapshot file [%s]: %v\n", snapshotPath, err)
		return
	}

	for _, fileChangeJSON := range data {
		fileSnapshot, err := unmurshalFileChangeNotification(fileChangeJSON)
		if err != nil {
			log.Printf("[ERROR] snapshot.CreateFirstBackup(): %s\n", err)
			continue
		}
		changes <- fileSnapshot
	}
}

// Diff returns diff between two snapshots as array of FileChangeNotification
func Diff(remoteSnapshotPath, localSnapshotPath string) (*[]types.FileChangeNotification, error) {
	var result []types.FileChangeNotification

	dbRemote, err := Snapshot(remoteSnapshotPath).GetAll()
	if err != nil {
		return nil, fmt.Errorf("snapshot.Diff(): cannot open snapshot file [%s]: leveldb.OpenFile(): %v", remoteSnapshotPath, err)
	}

	dbLocal, err := Snapshot(localSnapshotPath).GetAll()
	if err != nil {
		return nil, fmt.Errorf("snapshot.Diff(): can not open snapshot file [%s]: leveldb.OpenFile(): %v", localSnapshotPath, err)
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
