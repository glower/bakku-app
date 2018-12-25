package watchers

import (
	"log"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/config"
	"github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/bakku-app/pkg/watchers/watch"
)

// SetupWatchers adds a watcher for a file changes in all directories from the config
func SetupWatchers() []types.Notifications {
	list := []types.Notifications{}
	dirs := config.DirectoriesToWatch()
	log.Printf("SetupWatchers(): for %v\n", dirs)

	for _, dir := range dirs {
		//  callbackChan chan types.FileChangeNotification
		notifications := watchDirectoryForChanges(filepath.Clean(dir))
		list = append(list, *notifications)
	}
	return list
}

// watchDirectoryForChanges returns a channel with a notification about the changes in the specified directory
func watchDirectoryForChanges(path string) *types.Notifications {
	log.Printf("watchDirectoryForChanges(): [%s]\n", path)

	storages, err := storage.Active()
	if err != nil {
		log.Panic(err)
	}
	if len(storages) == 0 {
		log.Panicf("watchDirectoryForChanges(): can't find any active storages for the backup of [%s]", path)
	}

	notifications := snapshot.New(path, storages)
	go watch.NewNotifier(path, notifications.FileChangeChan)
	return notifications
}
