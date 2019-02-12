package watch

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/glower/bakku-app/pkg/file"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/bakku-app/pkg/watchers/notifications"
)

// ActionToString maps Action value to string
func ActionToString(action types.Action) string {
	switch action {
	case types.FileAdded:
		return "added"
	case types.FileRemoved:
		return "removed"
	case types.FileModified:
		return "modified"
	case types.FileRenamedOldName, types.FileRenamedNewName:
		return "renamed"
	default:
		return "invalid"
	}
}

// DirectoryWatcher ...
type DirectoryWatcher struct {
	FileChangeNotificationChan chan types.FileChangeNotification
}

// DirectoryWatcherImplementer ...
type DirectoryWatcherImplementer interface {
	StartWatching(path string)
}

var watcher *DirectoryWatcher
var once sync.Once

// SetupDirectoryWatcher ...
func SetupDirectoryWatcher(callbackChan chan types.FileChangeNotification) *DirectoryWatcher {
	once.Do(func() {
		watcher = &DirectoryWatcher{
			FileChangeNotificationChan: callbackChan,
		}
	})
	return watcher
}

func fileChangeNotifier(watchDirectoryPath, relativeFilePath string, fileInfo file.ExtendedFileInfoImplementer, action types.Action) {

	// ignore changes on the directory
	if fileInfo.IsDir() {
		return
	}

	if fileInfo.IsTemporaryFile() {
		return
	}

	if action == types.FileRemoved || action == types.FileRenamedOldName {
		return
	}

	absoluteFilePath := filepath.Join(watchDirectoryPath, relativeFilePath)
	log.Printf("watch.fileChangeNotifier(): watch directory path [%s], relative file path [%s], action [%s]\n", watchDirectoryPath, relativeFilePath, ActionToString(action))

	wait, exists := notifications.LookupForFileNotification(absoluteFilePath)
	if exists {
		wait <- true
		return
	}

	waitChan := make(chan bool)
	notifications.RegisterFileNotification(waitChan, absoluteFilePath)

	host, _ := os.Hostname()
	mimeType, err := fileInfo.ContentType()
	if err != nil {
		log.Printf("[ERROR] watch.fileChangeNotifier(): can't get ContentType from the file [%s]: %v\n", absoluteFilePath, err)
		notifications.UnregisterFileNotification(absoluteFilePath)
		return
	}

	data := &types.FileChangeNotification{
		MimeType:           mimeType,
		AbsolutePath:       absoluteFilePath,
		Action:             action,
		DirectoryPath:      watchDirectoryPath,
		Machine:            host,
		Name:               fileInfo.Name(),
		RelativePath:       relativeFilePath,
		Size:               fileInfo.Size(),
		Timestamp:          fileInfo.ModTime(),
		WatchDirectoryName: filepath.Base(watchDirectoryPath),
	}

	go notifications.FileNotificationWaiter(waitChan, watcher.FileChangeNotificationChan, data)
}
