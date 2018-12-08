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
func SetupWatchers() []types.Notifications {
	list := []types.Notifications{}
	dirs := config.DirectoriesToWatch()
	log.Printf("SetupWatchers(): for %v\n", dirs)

	for _, dir := range dirs {
		//  callbackChan chan types.FileChangeNotification
		watcher, done := watchDirectoryForChanges(filepath.Clean(dir))
		list = append(list, types.Notifications{
			FileChangeChan: watcher,
			DoneChan:       done,
		})
	}
	return list
}

// watchDirectoryForChanges returns a channel with a notification about the changes in the specified directory
func watchDirectoryForChanges(path string) (chan types.FileChangeNotification, chan bool) {
	log.Printf("watchDirectoryForChanges(): [%s]\n", path)
	changesChan := make(chan types.FileChangeNotification)
	changesDoneChan := make(chan bool)
	go snapshot.CreateOrUpdate(path, changesChan, changesDoneChan)
	go watch.NewNotifier(path, changesChan)
	return changesChan, changesDoneChan
}
