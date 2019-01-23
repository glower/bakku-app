package watchers

import (
	"log"

	"github.com/glower/bakku-app/pkg/config"
	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/bakku-app/pkg/watchers/watch"
)

// SetupFSWatchers adds a watcher for a file changes in all directories from the config
func SetupFSWatchers(fileChangeNotificationChan chan types.FileChangeNotification) {
	// read from the configuration file a list of directories to watch
	dirs := config.DirectoriesToWatch()
	log.Printf("watchers.SetupFSWatchers(): for %v\n", dirs)

	// read all supported backup storages form the config
	// storages, err := storage.Active()
	// if err != nil {
	// 	log.Panic(err)
	// }
	// if len(storages) == 0 {
	// 	log.Panicf("watchDirectoryForChanges(): can't find any active storages for the backup of [%s]", path)
	// }

	for _, dir := range dirs {
		go watch.NewNotifier(path, fileChangeNotificationChan)
	}
}

// watchDirectoryForChanges returns a channel with a notification about the changes in the specified directory
func watchDirectoryForChanges(path string) *types.Notifications {
	log.Printf("watchDirectoryForChanges(): [%s]\n", path)

	snapshotManager := snapshot.New(path, storages)
	go watch.NewNotifier(path, notifications.FileChangeChan)
	return notifications
}
