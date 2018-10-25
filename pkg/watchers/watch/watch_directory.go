package watch

import (
	"log"
	"os"
	"strings"
	"sync"
)

// Action stores corresponding action from Windows api, see: https://docs.microsoft.com/en-us/windows/desktop/api/winnt/ns-winnt-_file_notify_information
type Action int

const (
	// Invalid action is 0
	Invalid Action = iota
	// FileAdded - the file was added to the directory.
	FileAdded
	// FileRemoved - the file was removed from the directory.
	FileRemoved
	// FileModified - the file was modified. This can be a change in the time stamp or attributes.
	FileModified
	// FileRenamedOldName - the file was renamed and this is the old name.
	FileRenamedOldName
	// FileRenamedNewName - the file was renamed and this is the new name.
	FileRenamedNewName
)

// ActionToString maps Action value to string
func ActionToString(action Action) string {
	switch action {
	case FileAdded:
		return "added"
	case FileRemoved:
		return "removed"
	case FileModified:
		return "modified"
	case FileRenamedOldName | FileRenamedNewName:
		return "renamed"
	default:
		return "invalid"
	}
}

// FileChangeInfo ...
type FileChangeInfo struct {
	Action
	FileInfo os.FileInfo
	FilePath string
}

// CallbackData struct holds information about files in the watched directory
type CallbackData struct {
	CallbackChan chan FileChangeInfo
	Path         string
}

var callbackMutex sync.Mutex
var callbackFuncs = make(map[string]CallbackData)

// DirectoryChangeNotification expected path to the directory to watch as string
// and a FileInfo channel for the callback notofications
// Notofication is fired each time file in the directory is changed or some new
// file (or sub-directory) is created
func DirectoryChangeNotification(path string, callbackChan chan FileChangeInfo) {

	data := CallbackData{
		CallbackChan: callbackChan,
		Path:         path,
	}
	register(data, path)
	setupDirectoryChangeNotification(path)
}

func register(data CallbackData, path string) {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	callbackFuncs[path] = data
}

func lookup(path string) CallbackData {
	log.Printf("watch.lookup(): %s\n", path)
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	data, ok := callbackFuncs[path]
	if !ok {
		log.Printf("watch.lookup(): callback data for path=%s are not found!!!\n", path)
	}
	return data
}

func unregister(path string) {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	delete(callbackFuncs, path)
}

func fileChangeNotifier(path, file string, action Action) {
	log.Printf("watch.fileChangeNotifier(): [%s%s], action: %d\n", path, file, action)
	filePath := strings.TrimSpace(path + file)
	var fi os.FileInfo
	var err error
	if action != FileRemoved && action != FileRenamedOldName {
		fi, err = os.Stat(filePath)
		if err != nil {
			log.Printf("[ERROR] Can not stat file [%s]: %v\n", filePath, err)
			return
		}
	}
	callbackData := lookup(path)

	if fi != nil {
		callbackData.CallbackChan <- FileChangeInfo{
			Action:   action,
			FileInfo: fi,
			FilePath: filePath,
		}
	} else {
		log.Printf("[ERROR] watch.fileChangeNotifier(): FileInfo for [%s] not found!\n", filePath)
	}
}
