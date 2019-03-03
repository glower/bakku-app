package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/glower/file-watcher/notification"
	file "github.com/glower/file-watcher/util"
)

// ActionToString maps Action value to string
func ActionToString(action notification.ActionType) string {
	switch action {
	case notification.FileAdded:
		return "added"
	case notification.FileRemoved:
		return "removed"
	case notification.FileModified:
		return "modified"
	case notification.FileRenamedOldName, notification.FileRenamedNewName:
		return "renamed"
	default:
		return "invalid"
	}
}

// DirectoryWatcher ...
type DirectoryWatcher struct {
	ActionFilters []notification.ActionType
	FileFilters   []string
	Options       *Options
	EventCh       chan notification.Event
	ErrorCh       chan notification.Error

	NotificationWaiter notification.Waiter
}

// DirectoryWatcherImplementer ...
type DirectoryWatcherImplementer interface {
	StartWatching(path string)
}

// Options ...
type Options struct {
	IgnoreDirectoies bool
}

var watcher *DirectoryWatcher
var once sync.Once

// TODO: add options for filter dirs

// Create ...
func Create(callbackCh chan notification.Event, errorCh chan notification.Error, filters []notification.ActionType, fileFilters []string, options *Options) *DirectoryWatcher {
	once.Do(func() {
		watcher = &DirectoryWatcher{
			ActionFilters: filters,
			FileFilters:   fileFilters,
			Options:       options,
			EventCh:       callbackCh,
			ErrorCh:       errorCh,
			NotificationWaiter: notification.Waiter{
				EventCh:  callbackCh,
				Timeout:  time.Duration(5 * time.Second),
				MaxCount: 10,
			},
		}
	})
	return watcher
}

func fileError(err error) {
	watcher.ErrorCh <- notification.Error{
		Stack:   string(debug.Stack()),
		Message: err.Error(),
	}
}

func fileChangeNotifier(watchDirectoryPath, relativeFilePath string, fileInfo file.ExtendedFileInfoImplementer, action notification.ActionType) {

	if watcher.Options.IgnoreDirectoies == true && fileInfo.IsDir() {
		log.Printf("fileChangeNotifier(): file change for a directory [%s] is filtered\n", relativeFilePath)
		return
	}

	absoluteFilePath := filepath.Join(watchDirectoryPath, relativeFilePath)

	for _, fileFilter := range watcher.FileFilters {
		if strings.Contains(absoluteFilePath, fileFilter) {
			log.Printf("fileChangeNotifier(): file [%s] is filtered\n", fileFilter)
			return
		}
	}

	for _, actionFilter := range watcher.ActionFilters {
		if action == actionFilter {
			log.Printf("fileChangeNotifier(): action [%s] is filtered\n", ActionToString(actionFilter))
			return
		}
	}

	log.Printf("watch.fileChangeNotifier(): watch directory path [%s], relative file path [%s], action [%s]\n", watchDirectoryPath, relativeFilePath, ActionToString(action))

	wait, exists := watcher.NotificationWaiter.LookupForFileNotification(absoluteFilePath)
	if exists {
		wait <- true
		return
	}

	watcher.NotificationWaiter.RegisterFileNotification(absoluteFilePath)

	host, _ := os.Hostname()
	mimeType, err := fileInfo.ContentType()
	if err != nil {
		fileError(fmt.Errorf("can't get ContentType from the file [%s]: %v", absoluteFilePath, err))
		watcher.NotificationWaiter.UnregisterFileNotification(absoluteFilePath)
		return
	}

	data := &notification.Event{
		MimeType:           mimeType,
		AbsolutePath:       absoluteFilePath,
		Action:             action,
		DirectoryPath:      watchDirectoryPath,
		Machine:            host,
		FileName:           fileInfo.Name(),
		RelativePath:       relativeFilePath,
		Size:               fileInfo.Size(),
		Timestamp:          fileInfo.ModTime(),
		WatchDirectoryName: filepath.Base(watchDirectoryPath),
	}

	go watcher.NotificationWaiter.Wait(data)
}
