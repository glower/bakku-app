package types

// FileBackupComplete represents
type FileBackupComplete struct {
	BackupStorageName  string
	AbsolutePath       string
	WatchDirectoryName string
}

// BackupProgress represents a moment of progress.
type BackupProgress struct {
	StorageName string  `json:"storage"`
	FileName    string  `json:"file"`
	Percent     float64 `json:"percent"`
}

// // Notifications ...
// type Notifications struct {
// 	FileChangeChan chan FileChangeNotification
// 	DoneChan       chan bool
// }
