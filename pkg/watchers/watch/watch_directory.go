package watch

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/glower/bakku-app/pkg/types"
	mime "github.com/glower/bakku-app/pkg/watchers/file"
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

	if isTemporaryFile(filePath) {
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

		wait, exists := lookupForFileNotification(filePath)
		if exists {
			wait <- true
			return
		}

		waitChan := make(chan bool)
		registerFileNotification(waitChan, filePath)

		host, _ := os.Hostname()
		mimeType, err := mime.ContentType(filePath)
		if err != nil {
			log.Printf("[ERROR] watch.fileChangeNotifier(): can't get ContentType from the file [%s]: %v\n", filePath, err)
			unregisterFileNotification(filePath)
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

		go fileNotificationWaiter(waitChan, callbackData, data)

	} else {
		log.Printf("[ERROR] watch.fileChangeNotifier(): FileInfo for [%s] not found!\n", filePath)
	}
}

// -------------------------------------------------------------

var tmpFiles = []string{"crdownload"}

func isTemporaryFile(fileName string) bool {
	for _, name := range tmpFiles {
		if strings.Contains(fileName, name) {
			return true
		}
	}
	return false
}

// -------------------------------------------------------------
var notificationsMutex sync.Mutex
var notificationsChans = make(map[string]chan bool)

func registerFileNotification(waitChan chan bool, path string) {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	notificationsChans[path] = waitChan
}

func unregisterFileNotification(path string) {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	delete(notificationsChans, path)
}

func lookupForFileNotification(path string) (chan bool, bool) {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()
	data, ok := notificationsChans[path]
	return data, ok
}

// FileNotificationWaiter will send fileData to the chan stored in CallbackData after 5 seconds if no signal is
// received on waitChan.
func fileNotificationWaiter(waitChan chan bool, callbackData CallbackData, fileData *types.FileChangeNotification) {
	for {
		select {
		// TODO: add global timeout for like 10 min to avoid go routin zombies
		case <-waitChan:
			log.Printf(">>> waiting for [%s] to be ready", fileData.AbsolutePath)
		case <-time.After(time.Duration(5 * time.Second)):
			log.Printf(">>> done with file [%s]", fileData.AbsolutePath)
			callbackData.CallbackChan <- *fileData
			unregisterFileNotification(fileData.AbsolutePath)
			close(waitChan)
			return
		}
	}
}
