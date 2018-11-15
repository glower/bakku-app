package watchers

import (
	"log"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/config"
	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/bakku-app/pkg/watchers/watch"
)

// SetupWatchers adds a watcher for a file changes in all directories from the config
func SetupWatchers() []chan types.FileChangeNotification {
	list := []chan types.FileChangeNotification{}
	dirs := config.DirectoriesToWatch()
	log.Printf("SetupWatchers(): for %v\n", dirs)

	for _, dir := range dirs {
		watcher := WatchDirectoryForChanges(filepath.Clean(dir))
		list = append(list, watcher)
	}
	return list
}

// WatchDirectoryForChanges returns a channel with a notification about the changes in the specified directory
func WatchDirectoryForChanges(path string) chan types.FileChangeNotification {
	log.Printf("WatchDirectoryForChanges(): %s\n", path)
	changes := make(chan types.FileChangeNotification)
	log.Println("!!!!!!!!!")
	snapshot.CreateOrUpdate(path) // blocking here is ok
	if !snapshot.Exist(path) {
		log.Printf("WatchDirectoryForChanges: snapshot for [%s] does not exist\n", path)
		go snapshot.CreateFirstBackup(path, changes)
	}
	go watch.DirectoryChangeNotification(path, changes)
	return changes
}
