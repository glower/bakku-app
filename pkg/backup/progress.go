package backup

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/glower/bakku-app/pkg/types"
)

var (
	filesInProgressM sync.RWMutex
	filesInProgress  = make(map[string]time.Time)
)

// Start stores the information about files in progress
func Start(fileChange *types.FileChangeNotification, storage string) {
	file := fileChange.AbsolutePath
	filesInProgressM.Lock()
	defer filesInProgressM.Unlock()
	key := buildKey(file, storage)

	// TODO: find good strategy for this case
	if _, dup := filesInProgress[key]; dup {
		log.Printf("storage.BackupStarted(): file [%s] is in progress for the storage provider [%s]\n", file, storage)
		return
	}

	filesInProgress[key] = time.Now()
	return
}

// InProgress ...
func InProgress(fileChange *types.FileChangeNotification, storage string) bool {
	file := fileChange.AbsolutePath
	key := buildKey(file, storage)
	if _, dup := filesInProgress[key]; dup {
		log.Printf("storage.BackupStarted(): file [%s] is in progress for the storage provider [%s]\n", file, storage)
		return true
	}
	return false
}

// Finished ...
func Finished(fileChange *types.FileChangeNotification, storage string) {
	file := fileChange.AbsolutePath
	log.Printf("backup.Finished(): [%s] to [%s]\n", file, storage)
	log.Printf("-------------------------------------------------------\n\n\n")
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
