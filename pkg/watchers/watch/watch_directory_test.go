// +build integration

package watch

import (
	"log"
	"os"
	"path/filepath"
	"testing"
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
	log.Printf("test.StartWatching(): for [%s]\n", watchDirectoryPath)

	switch watchDirectoryPath {
	case "/foo/bar":
		relativeFilePath := "test/file.txt"
		fi := &FakeFileInfo{
			watchDirectoryPath: watchDirectoryPath,
			relativeFilePath:   relativeFilePath,
		}
		fileChangeNotifier(watchDirectoryPath, relativeFilePath, fi, types.Action(1)) // FileAdded
	}

}

// Run this test with: go test  -tags=integration  -timeout 30s -v github.com\glower\bakku-app\pkg\watchers\watch

func TestSetupDirectoryWatcher(t *testing.T) {
	type args struct {
		callbackChan chan types.FileChangeNotification
	}

	fileChangeNotificationChan := make(chan types.FileChangeNotification)

	tests := []struct {
		name string
		args args
		dir  string
		want *types.FileChangeNotification
	}{
		{
			name: "test 1: file change notification",
			args: args{
				callbackChan: fileChangeNotificationChan,
			},
			dir: "/foo/bar",
			want: &types.FileChangeNotification{
				Action:             1,
				BackupToStorages:   []string(nil),
				MimeType:           "image/jpeg",
				Machine:            "tokyo",
				Name:               "file.txt",
				AbsolutePath:       "\\foo\\bar\\test\\file.txt",
				RelativePath:       "test/file.txt",
				DirectoryPath:      "/foo/bar",
				WatchDirectoryName: "foo",
				Size:               12345,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := SetupDirectoryWatcher(tt.args.callbackChan)
			w.StartWatching(tt.dir)
			action := <-tt.args.callbackChan

			if action.Action != tt.want.Action {
				t.Errorf("action.Action = %v, want %v", action.Action, tt.want.Action)
			}

			if action.MimeType != tt.want.MimeType {
				t.Errorf("action.MimeType = %v, want %v", action.MimeType, tt.want.MimeType)
			}
		})
	}
}
