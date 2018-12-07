package types

import "time"

// FileChangeNotification ...
type FileChangeNotification struct {
	Action
	Machine            string
	Name               string
	AbsolutePath       string
	RelativePath       string
	DirectoryPath      string
	WatchDirectoryName string
	Size               int64
	Timestamp          time.Time
}

// Action represents what happens with the file
type Action int

const (
	// Invalid action is 0
	Invalid Action = iota
	// FileAdded - the file was added to the directory.
	FileAdded // 1
	// FileRemoved - the file was removed from the directory.
	FileRemoved
	// FileModified - the file was modified. This can be a change in the time stamp or attributes.
	FileModified
	// FileRenamedOldName - the file was renamed and this is the old name.
	FileRenamedOldName
	// FileRenamedNewName - the file was renamed and this is the new name.
	FileRenamedNewName
)
