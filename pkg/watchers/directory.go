package watchers

import (
	"github.com/glower/file-change-notification/watch"
)

// WatchDirectoryForChanges returns a channel with a notification about the changes in the specified directory
func WatchDirectoryForChanges(path string) chan watch.FileChangeInfo {
	changes := make(chan watch.FileChangeInfo)
	go watch.DirectoryChangeNotification(path, changes)
	return changes
}
