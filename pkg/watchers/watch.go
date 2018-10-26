package watchers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glower/bakku-app/pkg/watchers/watch"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

// SetupWatchers ...
func SetupWatchers() []chan watch.FileChangeInfo {
	list := []chan watch.FileChangeInfo{}

	// TODO: move this to some utils
	dirs, ok := viper.Get("watch").([]interface{})
	if !ok {
		log.Println("SetupWatchers(): nothing to watch")
		return list
	}

	for _, dir := range dirs {
		path, ok := dir.(string)
		if !ok {
			log.Println("SetupWatchers(): invalid path")
			continue
		}
		watcher := WatchDirectoryForChanges(filepath.Clean(path))
		list = append(list, watcher)
	}
	return list
}

// WatchDirectoryForChanges returns a channel with a notification about the changes in the specified directory
func WatchDirectoryForChanges(path string) chan watch.FileChangeInfo {
	snapshotPath := fmt.Sprintf("%s%s.snapshot", path, string(os.PathSeparator)) // TODO: do this in some config part
	changes := make(chan watch.FileChangeInfo)
	if !SnapshotExist(snapshotPath) {
		log.Printf("WatchDirectoryForChanges: snapshot for [%s] is does not exist\n", path)
		UpdateSnapshot(path, snapshotPath) // blocking here is ok
		go CreateFirstBackup(path, snapshotPath, changes)
	} else {
		// go UpdateSnapshot(path, snapshotPath)
		UpdateSnapshot(path, snapshotPath)
	}
	go watch.DirectoryChangeNotification(path, changes)
	return changes
}

// SnapshotExist ...
func SnapshotExist(snapshotPath string) bool {
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// UpdateSnapshot ...
func UpdateSnapshot(dir, snapshotPath string) {
	log.Printf("UpdateSnapshot(): %s\n", snapshotPath)
	db, err := leveldb.OpenFile(snapshotPath, nil)
	if err != nil {
		log.Printf("watchers.UpdateSnapshot(): can not open snapshot file [%s]: %v\n", snapshotPath, err)
		return
	}
	defer db.Close()
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			key := path
			value := fmt.Sprintf("%s:%d", f.ModTime(), f.Size())
			db.Put([]byte(key), []byte(value), nil)
		}
		return nil
	})
	log.Println("UpdateSnapshot(): done")
}

// CreateFirstBackup ...
func CreateFirstBackup(dir, snapshotPath string, changes chan watch.FileChangeInfo) {
	db, err := leveldb.OpenFile(snapshotPath, nil)
	if err != nil {
		log.Printf("watchers.UpdateSnapshot(): can not open snapshot file [%s]: %v\n", snapshotPath, err)
	}
	defer db.Close()

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		path := iter.Key()
		filePath := string(path)
		fileName := filepath.Base(filePath)
		relativePath := strings.Replace(filePath, dir, "", -1)
		// fmt.Printf("CreateFirstBackup(): relativePath=[%s]\n", relativePath)
		changes <- watch.FileChangeInfo{
			Action:        watch.Action(watch.FileAdded),
			FilePath:      filePath,
			FileName:      fileName,
			RelativePath:  relativePath,
			DirectoryPath: dir,
		}
	}
	iter.Release()
	err = iter.Error()
}
