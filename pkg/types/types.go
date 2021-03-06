package types

import (
	"github.com/glower/bakku-app/pkg/message"
	"github.com/glower/bakku-app/pkg/storage"
	"github.com/glower/file-watcher/watcher"
)

// BackupComplete represents
type BackupComplete struct {
	Success            bool
	StorageName        string
	FilePath           string
	WatchDirectoryName string
}

// BackupProgress represents a moment of progress.
type BackupProgress struct {
	StorageName  string  `json:"storage"`
	FileName     string  `json:"file"`
	AbsolutePath string  `json:"path"`
	ID           string  `json:"id"`
	Percent      float64 `json:"percent"`
}

type BackupStatus struct {
	TotalFiles      int    `json:"total"`
	FilesInProgress int    `json:"in_progress"`
	FilesDone       int    `json:"done"`
	Status          string `json:"status"`
}

type GlobalResources struct {
	BackupCompleteCh chan BackupComplete
	MessageCh        chan message.Message
	FileWatcher      *watcher.Watch
	Storage          storage.Storager
}
