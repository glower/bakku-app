package watch

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/glower/bakku-app/pkg/types"
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
// type FileChangeInfo struct {
// 	Action
// 	FileName      string
// 	FilePath      string
// 	RelativePath  string
// 	DirectoryPath string
// 	IsDir         bool
// }

// CallbackData struct holds information about files in the watched directory
type CallbackData struct {
	CallbackChan chan types.FileChangeNotification
	Path         string
}

var callbackMutex sync.Mutex
var callbackFuncs = make(map[string]CallbackData)

// DirectoryChangeNotification expected path to the directory to watch as string
// and a FileInfo channel for the callback notofications
// Notofication is fired each time file in the directory is changed or some new
// file (or sub-directory) is created
func DirectoryChangeNotification(path string, callbackChan chan types.FileChangeNotification) {
	log.Printf("watch.DirectoryChangeNotification(): path=[%s]\n", path)
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

func unregister(path string) {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	delete(callbackFuncs, path)
}

func fileChangeNotifier(path, file string, action Action) {
	if strings.Contains(file, snapshot.Dir()) {
		return
	}

	filePath := filepath.Join(path, file)
	log.Printf("watch.fileChangeNotifier(): [%s], action: %s\n", filePath, ActionToString(action))
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

	// This is a hack to see if the file is written to the disk on windows
	_, err = os.Open(filePath)
	for err != nil {
		time.Sleep(1 * time.Second)
		_, err = os.Open(filePath)
	}

	if fi != nil {
		callbackData.CallbackChan <- types.FileChangeNotification{
			Name:          fi.Name(),
			AbsolutePath:  filePath,
			RelativePath:  file,
			DirectoryPath: callbackData.Path,
		}
	} else {
		log.Printf("[ERROR] watch.fileChangeNotifier(): FileInfo for [%s] not found!\n", filePath)
	}
}
