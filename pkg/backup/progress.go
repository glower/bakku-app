package backup

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/glower/file-watcher/notification"
)

var (
	filesInProgressM sync.RWMutex
	filesInProgress  = make(map[string]time.Time)
)

// Start stores the information about files in progress
func Start(fileChange *notification.Event, storage string) {
	file := fileChange.AbsolutePath
	filesInProgressM.Lock()
	defer filesInProgressM.Unlock()
	key := buildKey(file, storage)

	// TODO: find good strategy for this case
	if _, dup := filesInProgress[key]; dup {
		log.Printf("storage.Start(): file [%s] is in progress for the storage provider [%s]\n", file, storage)
		return
	}

	filesInProgress[key] = time.Now()
	return
}

// InProgress ...
func InProgress(fileChange *notification.Event, storage string) bool {
	file := fileChange.AbsolutePath
	key := buildKey(file, storage)
	if _, dup := filesInProgress[key]; dup {
		log.Printf("storage.InProgress(): file [%s] is in progress for the storage provider [%s]\n", file, storage)
		return true
	}
	return false
}

// Finish ...
func Finish(fileChange *notification.Event, storage string) {
	file := fileChange.AbsolutePath
	filesInProgressM.Lock()
	defer filesInProgressM.Unlock()
	key := buildKey(file, storage)
	delete(filesInProgress, key)
}

// TotalFilesInProgres returns total number of files in progress
func TotalFilesInProgres() int {
	return len(filesInProgress)
}

func buildKey(file, storage string) string {
	return fmt.Sprintf("%s:%s", file, storage)
}
