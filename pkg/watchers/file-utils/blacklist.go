package fileutils

import "strings"

var tmpFiles = []string{".crdownload", ".lock"}

func IsTemporaryFile(fileName string) bool {
	for _, name := range tmpFiles {
		if strings.Contains(fileName, name) {
			return true
		}
	}
	return false
}
