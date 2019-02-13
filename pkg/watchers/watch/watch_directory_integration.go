// +build integration

package watch

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glower/bakku-app/pkg/types"
)

type FakeFileInfo struct {
	watchDirectoryPath string
	relativeFilePath   string
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
	case "/foo/bar":
		relativeFilePath := "test/file.txt"
		fi := &FakeFileInfo{
			watchDirectoryPath: watchDirectoryPath,
			relativeFilePath:   relativeFilePath,
		}
		fileChangeNotifier(watchDirectoryPath, relativeFilePath, fi, types.Action(1)) // FileAdded
	}
case "/foo":
	relativeFilePath := "test/"
	fi := &FakeFileInfo{
		watchDirectoryPath: watchDirectoryPath,
		relativeFilePath:   relativeFilePath,
	}
	fileChangeNotifier(watchDirectoryPath, relativeFilePath, fi, types.Action(1)) // FileAdded
}
}
