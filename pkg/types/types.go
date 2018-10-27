package types

import "time"

// FileChangeNotification ...
type FileChangeNotification struct {
	Name          string
	AbsolutePath  string
	RelativePath  string
	DirectoryPath string
	Size          int64
	Timestamp     time.Time
}
