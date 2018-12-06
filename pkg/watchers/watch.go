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
		//  callbackChan chan types.FileChangeNotification
		watcher := watchDirectoryForChanges(filepath.Clean(dir))
		list = append(list, watcher)
	}
	return list
}

// watchDirectoryForChanges returns a channel with a notification about the changes in the specified directory
func watchDirectoryForChanges(path string) chan types.FileChangeNotification {
	log.Printf("watchDirectoryForChanges(): [%s]\n", path)
	changes := make(chan types.FileChangeNotification)
	go snapshot.CreateOrUpdate(path, changes)
	go watch.NewNotifier(path, changes)
	return changes
}
