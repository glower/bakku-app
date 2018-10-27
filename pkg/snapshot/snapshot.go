package snapshot

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glower/bakku-app/pkg/types"
	"github.com/syndtr/goleveldb/leveldb"
)

const snapshotDirName = ".snapshot"
const appName = "bakku-app"

// Dir ...
func Dir() string {
	return filepath.Join(snapshotDirName)
}

// AppName ...
func AppName() string {
	return appName
}

// Path returns path for a snapshot for a given directory
func Path(path string) string {
	return filepath.Join(path, snapshotDirName)
}

// Exist ...
func Exist(snapshotPath string) bool {
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Update ...
func Update(dir, snapshotPath string) {
	log.Printf("UpdateSnapshot(): %s\n", snapshotPath)
	db, err := leveldb.OpenFile(snapshotPath, nil)
	if err != nil {
		log.Printf("watchers.UpdateSnapshot(): can not open snapshot file [%s]: %v\n", snapshotPath, err)
		return
	}
	defer db.Close()
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if strings.Contains(path, Dir()) {
			return nil
		}
		if !f.IsDir() {
			UpdateSnapshotEntry(dir, path, f, db)
		}
		return nil
	})
	log.Println("UpdateSnapshot(): done")
}

// UpdateSnapshotEntry ...
func UpdateSnapshotEntry(directoryPath, filePath string, f os.FileInfo, db *leveldb.DB) {
	key := filePath
	fileName := filepath.Base(filePath)
	relativePath := strings.Replace(filePath, directoryPath+string(os.PathSeparator), "", -1)
	snapshot := types.FileChangeNotification{
		AbsolutePath:  filePath,
		RelativePath:  relativePath,
		DirectoryPath: directoryPath,
		Name:          fileName,
		Size:          f.Size(),
		Timestamp:     f.ModTime(),
	}
	value, err := json.Marshal(snapshot)
	if err != nil {
		log.Printf("snapshot.Update(): cannot Marshal to json %#v: %v\n", snapshot, err)
		return
	}
	db.Put([]byte(key), value, nil)
}

// CreateFirstBackup ...
func CreateFirstBackup(dir, snapshotPath string, changes chan types.FileChangeNotification) {
	log.Printf("watchers.CreateFirstBackup(): for=[%s]\n", dir)
	db, err := leveldb.OpenFile(snapshotPath, nil)
	if err != nil {
		log.Printf("[ERROR] watchers.CreateFirstBackup(): can not open snapshot file [%s]: %v\n", snapshotPath, err)
		return
	}
	defer db.Close()

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		path := iter.Key()
		filePath := string(path)
		if strings.Contains(filePath, Dir()) {
			continue
		}
		fileName := filepath.Base(filePath)
		relativePath := strings.Replace(filePath, dir+string(os.PathSeparator), "", -1)
		changes <- types.FileChangeNotification{
			AbsolutePath:  filePath,
			RelativePath:  relativePath,
			DirectoryPath: dir,
			Name:          fileName,
		}
	}
	iter.Release()
}

// Diff returns diff between two snapshots as array of FileChangeNotification
func Diff(remoteSnapshotPath, localSnapshotPath string) (*[]types.FileChangeNotification, error) {
	var result []types.FileChangeNotification

	dbRemote, err := leveldb.OpenFile(remoteSnapshotPath, nil)
	if err != nil {
		return nil, fmt.Errorf("snapshot.Diff(): cannot open snapshot file [%s]: leveldb.OpenFile(): %v", remoteSnapshotPath, err)
	}
	defer dbRemote.Close()

	dbLocal, err := leveldb.OpenFile(localSnapshotPath, nil)
	if err != nil {
		return nil, fmt.Errorf("snapshot.Diff(): can not open snapshot file [%s]: leveldb.OpenFile(): %v", localSnapshotPath, err)
	}
	defer dbLocal.Close()

	iter := dbLocal.NewIterator(nil, nil)
	for iter.Next() {
		localFile := iter.Key()
		localInfo := iter.Value()
		remoteInfo, err := dbRemote.Get(localFile, nil)
		if strings.Contains(string(localFile), ".snapshot") {
			continue
		}
		if err != nil && err.Error() == "leveldb: not found" || string(localInfo) != string(remoteInfo) {
			log.Printf("snapshot.Diff(): key [%s] not found or different in the remote snapshot\n", string(localFile))
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
	iter.Release()
	return &result, nil
}

func unmurshalFileChangeNotification(value []byte) (types.FileChangeNotification, error) {
	change := types.FileChangeNotification{}
	if err := json.Unmarshal(value, &change); err != nil {
		return change, fmt.Errorf("cannot unmarshal data [%s]: %v", string(value), err)
	}
	return change, nil
}
