package watch

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/glower/bakku-app/pkg/types"
	fileutils "github.com/glower/bakku-app/pkg/watchers/file-utils"
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

// TODO: we can have different callbacks for different type events
func fileChangeNotifier(path, file string, action types.Action) {
	log.Printf("watch.fileChangeNotifier(): watch directory path [%s], relative file path [%s], action [%s]\n", path, file, ActionToString(action))

	absoluteFilePath := filepath.Join(path, file)
	if fileutils.IsTemporaryFile(absoluteFilePath) {
		return
	}

	if action == types.FileRemoved || action == types.FileRenamedOldName {
		return
	}

	var fileInfo os.FileInfo
	var err error

	fileInfo, err = os.Stat(absoluteFilePath)
	if err != nil {
		log.Printf("watch.fileChangeNotifier(): Can not stat file [%s]: %v\n", absoluteFilePath, err)
		return
	}

	// ignore changes on the directory
	if fileInfo.IsDir() {
		return
	}

	if fileInfo == nil {
		log.Printf("[ERROR] watch.fileChangeNotifier(): FileInfo for [%s] not found!\n", absoluteFilePath)
		return
	}
	wait, exists := notifications.LookupForFileNotification(absoluteFilePath)
	if exists {
		wait <- true
		return
	}

	waitChan := make(chan bool)
	notifications.RegisterFileNotification(waitChan, absoluteFilePath)

	host, _ := os.Hostname()
	mimeType, err := fileutils.ContentType(absoluteFilePath)
	if err != nil {
		log.Printf("[ERROR] watch.fileChangeNotifier(): can't get ContentType from the file [%s]: %v\n", absoluteFilePath, err)
		notifications.UnregisterFileNotification(absoluteFilePath)
		return
	}

	data := &types.FileChangeNotification{
		MimeType:           mimeType,
		AbsolutePath:       absoluteFilePath,
		Action:             action,
		DirectoryPath:      path,
		Machine:            host,
		Name:               fileInfo.Name(),
		RelativePath:       file,
		Size:               fileInfo.Size(),
		Timestamp:          fileInfo.ModTime(),
		WatchDirectoryName: filepath.Base(path),
	}

	go notifications.FileNotificationWaiter(waitChan, watcher.FileChangeNotificationChan, data)
}
