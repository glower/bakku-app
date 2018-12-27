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

// DirectoryChangeWacher ...
type DirectoryChangeWacher interface {
	SetupDirectoryChangeNotification(string)
}

// CallbackData struct holds information about files in the watched directory
type CallbackData struct {
	CallbackChan chan types.FileChangeNotification
	Path         string
}

var callbackMutex sync.Mutex
var callbackFuncs = make(map[string]CallbackData)
var callbackChannels = make(map[string]chan types.FileChangeNotification)

// NewNotifier expected path to the directory to watch as string
// and a FileInfo channel for the callback notofications
// Notofication is fired each time file in the directory is changed or some new
// file (or sub-directory) is created
func NewNotifier(path string, callbackChan chan types.FileChangeNotification) {
	w := &DirectoryChangeWacherImplementer{}
	directoryChangeNotification(path, callbackChan, w)
}

func directoryChangeNotification(path string, callbackChan chan types.FileChangeNotification, w DirectoryChangeWacher) {
	log.Printf("watch.DirectoryChangeNotification(): path=[%s]\n", path)
	data := CallbackData{
		CallbackChan: callbackChan,
		Path:         path,
	}
	register(data, path)
	w.SetupDirectoryChangeNotification(path)
}

func register(data CallbackData, path string) {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	callbackFuncs[path] = data
}

func unregister(path string) {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	delete(callbackFuncs, path)
}

// TODO: we can have different callbacks for different type events
func fileChangeNotifier(path, file string, action types.Action) {
	filePath := filepath.Join(path, file)

	if fileutils.IsTemporaryFile(filePath) {
		// log.Printf("watch.fileChangeNotifier(): file [%s] is a temporary file\n", filePath)
		return
	}

	//
	var fileInfo os.FileInfo
	var err error

	if action != types.FileRemoved && action != types.FileRenamedOldName {
		fileInfo, err = os.Stat(filePath)
		if err != nil {
			// log.Printf("watch.fileChangeNotifier(): Can not stat file [%s]: %v\n", filePath, err)
			return
		}

		// ignore changes on the directory
		if fileInfo.IsDir() {
			// log.Printf("watch.fileChangeNotifier(): [%s] is not a file, ignore\n", filePath)
			return
		}
	} else {
		// TODO: implement file delete strategy
		// log.Printf("watch.fileChangeNotifier(): not supported action [%d] for file [%s], ignore\n", action, filePath)
		return
	}

	log.Printf("watch.fileChangeNotifier(): file [%s] action [%s]\n", filePath, ActionToString(action))

	if fileInfo != nil {
		wait, exists := notifications.LookupForFileNotification(filePath)
		if exists {
			wait <- true
			return
		}

		waitChan := make(chan bool)
		notifications.RegisterFileNotification(waitChan, filePath)

		host, _ := os.Hostname()
		mimeType, err := fileutils.ContentType(filePath)
		if err != nil {
			log.Printf("[ERROR] watch.fileChangeNotifier(): can't get ContentType from the file [%s]: %v\n", filePath, err)
			notifications.UnregisterFileNotification(filePath)
			return
		}

		callbackData := lookup(path)
		data := &types.FileChangeNotification{
			MimeType:           mimeType,
			AbsolutePath:       filePath,
			Action:             action,
			DirectoryPath:      callbackData.Path,
			Machine:            host,
			Name:               fileInfo.Name(),
			RelativePath:       file,
			Size:               fileInfo.Size(),
			Timestamp:          fileInfo.ModTime(),
			WatchDirectoryName: filepath.Base(callbackData.Path),
		}

		go notifications.FileNotificationWaiter(waitChan, callbackData.CallbackChan, data)

	} else {
		log.Printf("[ERROR] watch.fileChangeNotifier(): FileInfo for [%s] not found!\n", filePath)
	}
}
