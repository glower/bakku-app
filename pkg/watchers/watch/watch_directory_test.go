// +build integration

package watch

import (
	"testing"

	"github.com/glower/bakku-app/pkg/types"
)

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
