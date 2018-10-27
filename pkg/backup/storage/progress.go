package storage

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	// log "github.com/sirupsen/logrus"
	"log"

	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/syndtr/goleveldb/leveldb"
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
func BackupStarted(file, storage string) bool {
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
func BackupFinished(file, storage string) {
	log.Printf("storage.BackupFinished(): [%s]", file)
	filesInProgressM.Lock()
	defer filesInProgressM.Unlock()
	key := buildKey(file, storage)
	delete(filesInProgress, key)
}

// UpdateSnapshot ...
func UpdateSnapshot(snapshotPath, directoryPath, absolutePath string) {
	if strings.Contains(absolutePath, snapshot.Dir()) {
		return
	}
	log.Printf("storage.UpdateSnapshot(): %s\n", snapshotPath)
	db, err := leveldb.OpenFile(snapshotPath, nil)
	if err != nil {
		log.Printf("storage.UpdateSnapshot(): can not open snapshot file [%s]: %v\n", snapshotPath, err)
		return
	}
	defer db.Close()
	f, err := os.Stat(absolutePath)
	if err != nil {
		log.Printf("storage.UpdateSnapshot(): can not stat file [%s]: %v\n", absolutePath, err)
		return
	}
	snapshot.UpdateSnapshotEntry(directoryPath, absolutePath, f, db)

	log.Println("storage.UpdateSnapshot(): done")
}

// TotalFilesInProgres returns total number of files in progress
func TotalFilesInProgres() int {
	return len(filesInProgress)
}

func buildKey(file, storage string) string {
	return fmt.Sprintf("%s:%s", file, storage)
}
