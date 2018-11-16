package watch

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/glower/bakku-app/pkg/types"
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
	case types.FileRenamedOldName | types.FileRenamedNewName:
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
	log.Printf("watch.fileChangeNotifier(): [%s], action: %s\n", filePath, ActionToString(action))
	var fi os.FileInfo
	var err error
	callbackData := lookup(path)
	if action != types.FileRemoved && action != types.FileRenamedOldName {
		fi, err = os.Stat(filePath)
		if err != nil {
			log.Printf("[ERROR] watch.fileChangeNotifier(): Can not stat file [%s]: %v\n", filePath, err)
			return
		}
	} else {
		log.Printf("watch.fileChangeNotifier(): file [%s] was deleted, update the snapshot\n", file)
		callbackData.CallbackChan <- types.FileChangeNotification{
			Action:        action,
			Name:          filepath.Base(file),
			AbsolutePath:  filePath,
			RelativePath:  file,
			DirectoryPath: callbackData.Path,
		}
		return
	}

	// This is a hack to see if the file is written to the disk on windows
	_, err = os.Open(filePath)
	for err != nil {
		time.Sleep(1 * time.Second)
		_, err = os.Open(filePath)
	}

	if fi != nil {
		callbackData.CallbackChan <- types.FileChangeNotification{
			Action:        action,
			Name:          fi.Name(),
			AbsolutePath:  filePath,
			RelativePath:  file,
			DirectoryPath: callbackData.Path,
			Size:          fi.Size(),
		}
	} else {
		log.Printf("[ERROR] watch.fileChangeNotifier(): FileInfo for [%s] not found!\n", filePath)
	}
}
