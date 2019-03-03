// +build integration

package watcher

import (
	"testing"

	"github.com/glower/file-watcher/notification"
)

func TestSetupDirectoryWatcher(t *testing.T) {
	type args struct {
		callbackChan  chan notification.Event
		errorCh       chan notification.Error
		options       *Options
		actionFilters []notification.ActionType
		fileFilters   []string
	}

	eventCh := make(chan notification.Event)
	errorCh := make(chan notification.Error)

	tests := []struct {
		name string
		args args
		dir  string
		want *notification.Event
	}{
		{
			name: "test 1: file change notification",
			args: args{
				callbackChan:  eventCh,
				errorCh:       errorCh,
				options:       &Options{},
				actionFilters: []notification.ActionType{},
				fileFilters:   []string{},
			},
			dir: "/test1",
			want: &notification.Event{
				Action:             1,
				MimeType:           "image/jpeg",
				Machine:            "tokyo",
				FileName:           "file1.txt",
				AbsolutePath:       "\\foo\\bar\\test\\file1.txt",
				RelativePath:       "test/file1.txt",
				DirectoryPath:      "/test1",
				WatchDirectoryName: "foo",
				Size:               12345,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := Create(tt.args.callbackChan, tt.args.errorCh, tt.args.actionFilters, tt.args.fileFilters, tt.args.options)
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
