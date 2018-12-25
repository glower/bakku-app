package storage

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/glower/bakku-app/pkg/types"
)

// Progress represents a moment of progress.
type Progress struct {
	StorageName string  `json:"storage"`
	FileName    string  `json:"file"`
	Percent     float64 `json:"percent"`
}

var (
	filesInProgressM sync.RWMutex
	filesInProgress  = make(map[string]time.Time)
)

// BackupStarted ...
func BackupStarted(fileChange *types.FileChangeNotification, storage string) bool {
	file := fileChange.AbsolutePath
	filesInProgressM.Lock()
	defer filesInProgressM.Unlock()
	key := buildKey(file, storage)

	// TODO: find good strategy for this case
	if _, dup := filesInProgress[key]; dup {
		log.Printf("storage.BackupStarted(): file [%s] is in progress for the storage provider [%s]\n", file, storage)
		return false
	}

	filesInProgress[key] = time.Now()
	return true
}

// IsInProgress ...
func IsInProgress(file, storage string) bool {
	return false
}

// BackupFinished ...
func BackupFinished(fileChange *types.FileChangeNotification, storage string) {
	file := fileChange.AbsolutePath
	log.Printf("storage.BackupFinished(): [%s] to [%s]\n", file, storage)
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
