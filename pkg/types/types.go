package types

// FileBackupComplete represents
type FileBackupComplete struct {
	BackupStorageName  string
	AbsolutePath       string
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
	Status          string `json:"status"`
}
