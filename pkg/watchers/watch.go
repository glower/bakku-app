package watchers

import (
	"log"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/glower/bakku-app/pkg/watchers/watch"
	"github.com/spf13/viper"
)

// SetupWatchers ...
func SetupWatchers() []chan storage.FileChangeNotification {
	list := []chan storage.FileChangeNotification{}

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
func WatchDirectoryForChanges(path string) chan storage.FileChangeNotification {
	snapshotPath := filepath.Join(path, snapshot.Dir())
	changes := make(chan storage.FileChangeNotification)
	if !snapshot.Exist(snapshotPath) {
		log.Printf("WatchDirectoryForChanges: snapshot for [%s] is does not exist\n", path)
		snapshot.Update(path, snapshotPath) // blocking here is ok
		go snapshot.CreateFirstBackup(path, snapshotPath, changes)
	} else {
		// go UpdateSnapshot(path, snapshotPath)
		snapshot.Update(path, snapshotPath)
	}
	go watch.DirectoryChangeNotification(path, changes)
	return changes
}
