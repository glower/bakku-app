package file

import "strings"

var tmpFiles = []string{".crdownload", ".lock", ".snapshot"}

// IsTemporaryFile ...
func (i *FileInformation) IsTemporaryFile() bool {
	for _, name := range tmpFiles {
		if strings.Contains(i.absoluteFilePath, name) {
			return true
		}
	}
	return false
}
