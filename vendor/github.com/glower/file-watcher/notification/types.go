package notification

import (
	"time"

	"github.com/google/uuid"
)

// ActionType represents what happens with the file
type ActionType int

// Event ...
type Event struct {
	Action             ActionType
	MimeType           string
	Machine            string
	FileName           string
	AbsolutePath       string
	RelativePath       string
	DirectoryPath      string
	WatchDirectoryName string
	Size               int64
	Timestamp          time.Time
	UUID               uuid.UUID
	Checksum           string
}

const (
	// Invalid action is 0
	Invalid ActionType = iota
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
