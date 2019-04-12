// +build integration

package watcher

import (
	"os"
	"path/filepath"
	"time"

	"github.com/glower/file-watcher/notification"
)

type FakeFileInfo struct {
	watchDirectoryPath string
	relativeFilePath   string
}

func (i *FakeFileInfo) Checksum() (string, error) {
	return "ABCabc", nil
}
func (i *FakeFileInfo) IsTemporaryFile() bool {
	return false
}
func (f *FakeFileInfo) ContentType() (string, error) {
	return "image/jpeg", nil
}
func (f *FakeFileInfo) Name() string {
	return filepath.Base(f.relativeFilePath)
}
func (f *FakeFileInfo) Size() int64 {
	return 12345
}
func (f *FakeFileInfo) Mode() os.FileMode {
	return 0
}
func (f *FakeFileInfo) ModTime() time.Time {
	return time.Now()
}
func (f *FakeFileInfo) IsDir() bool {
	if f.relativeFilePath == "" {
		return true
	}
	return false
}
func (f *FakeFileInfo) Sys() interface{} {
	return nil
}

func (w *DirectoryWatcher) StartWatching(watchDirectoryPath string) {
	switch watchDirectoryPath {
	case "/test1":
		relativeFilePath := "test/file1.txt"
		fi := &FakeFileInfo{
			watchDirectoryPath: watchDirectoryPath,
			relativeFilePath:   relativeFilePath,
		}
		fileChangeNotifier(watchDirectoryPath, relativeFilePath, fi, notification.ActionType(1)) // FileAdded
	case "/test2":
		relativeFilePath := "test/file2.txt"
		fi := &FakeFileInfo{
			watchDirectoryPath: watchDirectoryPath,
			relativeFilePath:   relativeFilePath,
		}
		fileChangeNotifier(watchDirectoryPath, relativeFilePath, fi, notification.ActionType(2)) // FileRemoved
	}
}
