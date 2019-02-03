package watchers

import (
	"context"
	"log"

	"github.com/glower/bakku-app/pkg/config"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/bakku-app/pkg/watchers/watch"
)

// SetupFSWatchers adds a watcher for a file changes in all directories from the config
func SetupFSWatchers(ctx context.Context, fileChangeNotificationChan chan types.FileChangeNotification) {
	// read from the configuration file a list of directories to watch
	dirs := config.DirectoriesToWatch()
	log.Printf("watchers.SetupFSWatchers(): for %v\n", dirs)

	for _, dir := range dirs {
		w := watch.SetupDirectoryWatcher(fileChangeNotificationChan)
		go w.StartWatching(dir)
	}
}
