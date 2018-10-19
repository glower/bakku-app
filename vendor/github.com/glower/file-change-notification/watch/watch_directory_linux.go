// +build linux

package watch

/*
#include <stdlib.h>
#include <stdio.h>
#include <sys/inotify.h>
#include <limits.h>
#include <unistd.h>

extern void goCallbackFileChange(void);

#define BUF_LEN (10 * (sizeof(struct inotify_event) + NAME_MAX + 1))

static inline void WatchDirectory(const char* dir) {
  int inotifyFd, wd, j;
  char buf[BUF_LEN] __attribute__ ((aligned(8)));
  ssize_t numRead;
  char *p;
  struct inotify_event *event;

  inotifyFd = inotify_init();
  if (inotifyFd == -1) {
		printf("[ERROR] CGO: inotify_init()");
		exit(-1);
	}

  wd = inotify_add_watch(inotifyFd, dir, IN_CREATE | IN_DELETE | IN_MODIFY);
  if (wd == -1) {
		printf("[ERROR] CGO: inotify_add_watch()");
		exit(-1);
	}

  printf("[INFO] CGO: Watching %s\n", dir);

  for (;;) {
    numRead = read(inotifyFd, buf, BUF_LEN);
    if (numRead == 0) {
			printf("[ERROR] CGO: read() from inotify fd returned 0!");
			exit(-1);
		}

    if (numRead == -1) {
			printf("[ERROR] CGO: read()");
			exit(-1);
		}

    for (p = buf; p < buf + numRead; ) {
			event = (struct inotify_event *) p;
			printf("[INFO] CGO: file was changed\n");
			goCallbackFileChange();
      p += sizeof(struct inotify_event) + event->len;
    }
  }
}
*/
import "C"
import (
	"log"
	"os"
	"path/filepath"
	"unsafe"
)

// TODO: inotify is not recursive!!!

// DirectoryChangeNotification ...
func DirectoryChangeNotification(path string, callbackChan chan os.FileInfo) {
	snap = Snapshot{
		CallbackChan: callbackChan,
		Root:         path,
		DirSnapshot:  createSnapshot(path),
	}

	go watchDirectoryAPI(path)
	// cpath := C.CString(path)
	// defer C.free(unsafe.Pointer(cpath))
	// C.WatchDirectory(cpath)
}

func watchDirectoryAPI(path string) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	C.WatchDirectory(cpath)
}

//export goCallbackFileChange
func goCallbackFileChange() {
	// log.Println("goCallbackFileChange")
	fileInfo := findChange()
	if fileInfo != nil {
		snap.CallbackChan <- fileInfo
	}
}

func createSnapshot(path string) map[string]os.FileInfo {
	defer track(runningtime("createSnapshot()"))
	fileList := make(map[string]os.FileInfo)
	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			// log.Printf("createSnapshot(): %s is a dir\n", path)
			go watchDirectoryAPI(path)
		}
		fileList[path] = f
		return nil
	})
	log.Printf("createSnapshot(): found %d files\n", len(fileList))
	return fileList
}
