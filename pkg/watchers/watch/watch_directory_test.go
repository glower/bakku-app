package watch

import (
	"testing"

	"github.com/glower/bakku-app/pkg/types"
)

type DirectoryChangeWacherFakeImplementer struct{}

func (i *DirectoryChangeWacherFakeImplementer) SetupDirectoryChangeNotification(path string) {
	fileChangeNotifier(path, "watch_directory_test.go", types.Action(int(1)))
}

func TestTileChangeNotifier(t *testing.T) {
	changes := make(chan types.FileChangeNotification)
	fakeNotifier := &DirectoryChangeWacherFakeImplementer{}
	go directoryChangeNotification(".", changes, fakeNotifier)
	change := <-changes
	if change.Name != "watch_directory_test.go" {
		t.Errorf("expected file name [%s] got: [%s]", "watch_directory_test.go", change.Name)
	}
}
