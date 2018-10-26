package watchers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/watchers/watch"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

// SetupWatchers ...
func SetupWatchers() []chan watch.FileChangeInfo {
	list := []chan watch.FileChangeInfo{}

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
	if !SnapshotExist(snapshotPath) {
		log.Printf("WatchDirectoryForChanges: snapshot for [%s] is does not exist\n", path)
	}
	go UpdateSnapshot(snapshotPath)
	changes := make(chan watch.FileChangeInfo)
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
func UpdateSnapshot(snapshotPath string) {
	log.Printf("UpdateSnapshot(): %s\n", snapshotPath)
	db, err := leveldb.OpenFile(snapshotPath, nil)
	if err != nil {
		log.Printf("watchers.UpdateSnapshot(): can not open snapshot file [%s]: %v\n", snapshotPath, err)
	}
	defer db.Close()
	filepath.Walk(snapshotPath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			key := path
			value := fmt.Sprintf("%s:%d", f.ModTime(), f.Size())
			db.Put([]byte(key), []byte(value), nil)
		}
		return nil
	})
	log.Println("UpdateSnapshot(): done")
}
