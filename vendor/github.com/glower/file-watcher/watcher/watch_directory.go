package watcher

import (
	"fmt"
	"os"
	"path/filepath"
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

func fileError(lvl string, err error) {
	// TODO: we can print out here if it is configured
	watcher.ErrorCh <- notification.FormatError(lvl, err.Error())
}

func fileDebug(lvl string, msg string) {
	// TODO: we can print out here if it is configured
	watcher.ErrorCh <- notification.FormatError(lvl, msg)
}

func fileChangeNotifier(watchDirectoryPath, relativeFilePath string, fileInfo file.ExtendedFileInfoImplementer, action notification.ActionType) {

	if watcher.Options.IgnoreDirectoies == true && fileInfo.IsDir() {
		fileDebug("DEBUG", fmt.Sprintf("file change for a directory [%s] is filtered", relativeFilePath))
		return
	}

	absoluteFilePath := filepath.Join(watchDirectoryPath, relativeFilePath)

	for _, fileFilter := range watcher.FileFilters {
		if strings.Contains(absoluteFilePath, fileFilter) {
			fileDebug("DEBUG", fmt.Sprintf("file [%s] is filtered", fileFilter))
			return
		}
	}

	for _, actionFilter := range watcher.ActionFilters {
		if action == actionFilter {
			fileDebug("DEBUG", fmt.Sprintf("action [%s] is filtered\n", ActionToString(actionFilter)))
			return
		}
	}

	fileDebug("DEBUG", fmt.Sprintf("watch directory path [%s], relative file path [%s], action [%s]\n", watchDirectoryPath, relativeFilePath, ActionToString(action)))

	wait, exists := watcher.NotificationWaiter.LookupForFileNotification(absoluteFilePath)
	if exists {
		wait <- true
		return
	}

	watcher.NotificationWaiter.RegisterFileNotification(absoluteFilePath)

	host, _ := os.Hostname()
	mimeType, err := fileInfo.ContentType()
	if err != nil {
		fileError("WARNING", fmt.Errorf("can't get ContentType from the file [%s]: %v", absoluteFilePath, err))
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
