/*
Package notifications fixes the problem of multiple file change notifications for the same file from the OS API.
With FileNotificationWaiter(chanA, chanB, data) you can send data to the chanB if nothing was send to the chanA for 5 seconds

The flow:
- For each file create a channel and store it with RegisterFileNotification()
- Call FileNotificationWaiter() as a go routin with the created channel and other needed data
- On the next file change notification check if the channel for this file exists, if so send true to the channel
- If nothing was send on the channel, FileNotificationWaiter() will send the data to the provided channel after 5 seconds*/
package notification

import (
	"testing"
	"time"
)

func TestNotificationWaiter_RegisterFileNotification(t *testing.T) {
	type fields struct {
		EventCh  chan Event
		Timeout  time.Duration
		MaxCount int
	}
	type args struct {
		path string
	}

	tests := []struct {
		name                 string
		fields               fields
		args                 args
		notificationExpected bool
		fileData             *Event
	}{
		{
			name: "test 1: notification is fired after Timeout",
			fields: fields{
				EventCh:  make(chan Event),
				Timeout:  time.Duration(1 * time.Millisecond),
				MaxCount: 10,
			},
			args: args{
				path: "/foo/bar/test.txt",
			},
			fileData: &Event{
				Action:             ActionType(1),
				AbsolutePath:       "/foo/bar/test.txt",
				RelativePath:       "bar/test.txt",
				WatchDirectoryName: "foo",
				DirectoryPath:      "/foo",
			},
			notificationExpected: true,
		},
		{
			name: "test 2: notification is not fired",
			fields: fields{
				EventCh:  make(chan Event),
				Timeout:  time.Duration(5 * time.Second),
				MaxCount: 1,
			},
			args: args{
				path: "/foo/bar/test.txt",
			},
			fileData: &Event{
				Action:             ActionType(1),
				AbsolutePath:       "/foo/bar/test.txt",
				RelativePath:       "bar/test.txt",
				WatchDirectoryName: "foo",
				DirectoryPath:      "/foo",
			},
			notificationExpected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Waiter{
				EventCh:  tt.fields.EventCh,
				Timeout:  tt.fields.Timeout,
				MaxCount: tt.fields.MaxCount,
			}
			w.RegisterFileNotification(tt.args.path)
			waitChan, exists := w.LookupForFileNotification(tt.args.path)
			if !exists {
				t.Errorf("LookupForFileNotification(): got %v, want %v", exists, true)
			}

			_, exists = w.LookupForFileNotification("/some/other/path")
			if exists {
				t.Errorf("LookupForFileNotification(): got %v, want %v", exists, false)
			}

			go w.Wait(tt.fileData)
			waitChan <- true
			if tt.notificationExpected {
				file := <-tt.fields.EventCh
				if file.AbsolutePath != tt.fileData.AbsolutePath {
					t.Errorf("FileChangeNotification: got AbsolutePath=%s, want %s", file.AbsolutePath, tt.fileData.AbsolutePath)
				}
			}

			w.UnregisterFileNotification(tt.args.path)
			_, exists = w.LookupForFileNotification(tt.args.path)
			if exists {
				t.Errorf("LookupForFileNotification(): got %v, want %v", exists, false)
			}
		})
	}
}
