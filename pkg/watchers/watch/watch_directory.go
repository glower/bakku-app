package watch

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	NotificationWaiter         notifications.NotificationWaiter
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
			NotificationWaiter: notifications.NotificationWaiter{
				FileChangeNotificationChan: callbackChan,
				Timeout:                    time.Duration(5 * time.Second),
				MaxCount:                   10,
			},
		}
	})
	return watcher
}

func fileChangeNotifier(watchDirectoryPath, relativeFilePath string, fileInfo file.ExtendedFileInfoImplementer, action types.Action) error {

	// ignore changes on the directory
	if fileInfo.IsDir() {
		return fmt.Errorf("file change for a directory is not supported")
	}

	if fileInfo.IsTemporaryFile() {
		return fmt.Errorf("file change for a tmp file is not supported")
	}

	if action == types.FileRemoved || action == types.FileRenamedOldName {
		return fmt.Errorf("file change for a rename/move is not supported")
	}

	absoluteFilePath := filepath.Join(watchDirectoryPath, relativeFilePath)
	log.Printf("watch.fileChangeNotifier(): watch directory path [%s], relative file path [%s], action [%s]\n", watchDirectoryPath, relativeFilePath, ActionToString(action))

	wait, exists := watcher.NotificationWaiter.LookupForFileNotification(absoluteFilePath)
	if exists {
		wait <- true
		return nil
	}

	watcher.NotificationWaiter.RegisterFileNotification(absoluteFilePath)

	host, _ := os.Hostname()
	mimeType, err := fileInfo.ContentType()
	if err != nil {
		watcher.NotificationWaiter.UnregisterFileNotification(absoluteFilePath)
		return fmt.Errorf("can't get ContentType from the file [%s]: %v", absoluteFilePath, err)
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

	go watcher.NotificationWaiter.Wait(data)

	return nil
}
