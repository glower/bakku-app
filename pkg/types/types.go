package types

import "time"

// FileChangeNotification ...
type FileChangeNotification struct {
	Action
	// If we want to store this file only at specific backup storage provider
	BackupToStorages   []string
	MimeType           string
	Machine            string
	Name               string
	AbsolutePath       string
	RelativePath       string
	DirectoryPath      string
	WatchDirectoryName string
	Size               int64
	Timestamp          time.Time
}

// Notifications ...
type Notifications struct {
	FileChangeChan chan FileChangeNotification
	DoneChan       chan bool
}

// Action represents what happens with the file
type Action int

const (
	// Invalid action is 0
	Invalid Action = iota
	// FileAdded - the file was added to the directory.
	FileAdded // 1
	// FileRemoved - the file was removed from the directory.
	FileRemoved // 2
	// FileModified - the file was modified. This can be a change in the time stamp or attributes.
	FileModified // 3
	// FileRenamedOldName - the file was renamed and this is the old name.
	FileRenamedOldName // 4
	// FileRenamedNewName - the file was renamed and this is the new name.
	FileRenamedNewName // 5
)
