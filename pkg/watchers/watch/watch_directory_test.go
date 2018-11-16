package watch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glower/bakku-app/pkg/types"
)

type ProtoNotification struct {
	dir    func() string
	action func() types.Action
}

func (p *ProtoNotification) SetupDirectoryChangeNotification(path string) {
	fileChangeNotifier(path, p.dir(), types.Action(p.action()))
}

func TestFileChangeNotifierActions(t *testing.T) {
	testcases := []struct {
		Name            string
		DirectoryToWach string
		Notification    func()
		FileName        string
		Action          types.Action
	}{
		{
			Name:            "scenario 1: file was added",
			DirectoryToWach: filepath.Join(os.TempDir(), "bakku-app", "tests", "test-1"),
			FileName:        "foo.jpg",
			Action:          types.FileAdded,
		},
		{
			Name:            "scenario 2: file was modified",
			DirectoryToWach: filepath.Join(os.TempDir(), "bakku-app", "tests", "test-1"),
			FileName:        "foo.jpg",
			Action:          types.FileModified,
		},
		{
			Name:            "scenario 3: file was deleted",
			DirectoryToWach: filepath.Join(os.TempDir(), "bakku-app", "tests", "test-1"),
			FileName:        "foo.jpg",
			Action:          types.FileRemoved,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			// create test dir
			if err := os.MkdirAll(tc.DirectoryToWach, 0744); err != nil {
				t.Fatalf("Cannot create test directory [%s]: %v\n", tc.DirectoryToWach, err)
			}
			// create test file
			to, err := os.OpenFile(filepath.Join(tc.DirectoryToWach, tc.FileName), os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				t.Fatalf("Cannot open file test directory [%s]: %v\n", tc.FileName, err)
				return
			}
			to.Close()
			changes := make(chan types.FileChangeNotification)

			proto := &ProtoNotification{}
			proto.action = func() types.Action { return tc.Action }
			proto.dir = func() string { return tc.FileName }

			go directoryChangeNotification(tc.DirectoryToWach, changes, proto)
			change := <-changes
			if change.Name != tc.FileName {
				t.Errorf("expected file name was [%s], got: [%s]", tc.FileName, change.Name)
			}
			if change.Action != tc.Action {
				t.Errorf("expected file action was [%d], got: [%d]", tc.Action, change.Action)
			}
			if change.DirectoryPath != tc.DirectoryToWach {
				t.Errorf("expected path was [%s], got: [%s]", tc.DirectoryToWach, change.DirectoryPath)
			}
		})
	}
}
