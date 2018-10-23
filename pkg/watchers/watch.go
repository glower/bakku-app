package watchers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/watchers/watch"
	"github.com/syndtr/goleveldb/leveldb"
)

// WatchDirectoryForChanges returns a channel with a notification about the changes in the specified directory
func WatchDirectoryForChanges(path string) chan watch.FileChangeInfo {
	go UpdateSnapshot(path)
	changes := make(chan watch.FileChangeInfo)
	go watch.DirectoryChangeNotification(path, changes)
	return changes
}

// UpdateSnapshot ...
func UpdateSnapshot(path string) {
	log.Printf("UpdateSnapshot(): %s\n", path)
	// TODO: put leveldb to Scanner
	snapshot := path + ".snapshot"
	db, err := leveldb.OpenFile(snapshot, nil)
	if err != nil {
		log.Printf("watchers.UpdateSnapshot(): can not open snapshot file [%s]: %v\n", snapshot, err)
	}
	defer db.Close()
	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			key := path
			value := fmt.Sprintf("%s:%d", f.ModTime(), f.Size())
			db.Put([]byte(key), []byte(value), nil)
		}
		return nil
	})
	log.Println("UpdateSnapshot(): done")
}
