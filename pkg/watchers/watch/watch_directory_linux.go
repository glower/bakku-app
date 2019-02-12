// +build linux,!integration

package watch

/*
#include <stdlib.h>
#include <stdio.h>
#include <sys/inotify.h>
#include <limits.h>
#include <unistd.h>
#include <dirent.h>
#include <string.h>
#include <pthread.h>

#define BUF_LEN (10 * (sizeof(struct inotify_event) + NAME_MAX + 1))

extern void goCallbackFileChange(char* path, char* file, int action);

static inline void *WatchDirectory(char* dir) {
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

   	wd = inotify_add_watch(inotifyFd, dir, IN_CLOSE_WRITE | IN_DELETE);
   	if (wd == -1) {
		printf("[CGO] [ERROR] WatchDirectory(): inotify_add_watch()");
		exit(-1);
	}

  	printf("[CGO] [INFO] WatchDirectory(): watching %s\n", dir);
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
			printf("[INFO] CGO: file was changed: mask=%x, len=%d\n", event->mask, event->len);
			goCallbackFileChange(dir, event->name, event->mask);
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
	"strings"
	"unsafe"

	fileinfo "github.com/glower/bakku-app/pkg/file"
	"github.com/glower/bakku-app/pkg/types"
)

// #define IN_ACCESS		0x00000001	/* File was accessed */
// #define IN_MODIFY		0x00000002	/* File was modified */
// #define IN_ATTRIB		0x00000004	/* Metadata changed */
// #define IN_CLOSE_WRITE	0x00000008	/* Writtable file was closed */
// #define IN_CLOSE_NOWRITE	0x00000010	/* Unwrittable file closed */
// #define IN_OPEN			0x00000020	/* File was opened */
// #define IN_MOVED_FROM	0x00000040	/* File was moved from X */
// #define IN_MOVED_TO		0x00000080	/* File was moved to Y */
// #define IN_CREATE		0x00000100	/* Subfile was created */
// #define IN_DELETE		0x00000200	/* Subfile was deleted */
// #define IN_DELETE_SELF	0x00000400	/* Self was deleted */
func convertMaskToAction(mask int) types.Action {
	switch mask {
	case 2, 8: // File was modified
		return types.Action(types.FileModified)
	case 256: // Subfile was created
		return types.Action(types.FileAdded)
	case 512: // Subfile was deleted
		return types.Action(types.FileRemoved)
	case 64: // File was moved from X
		return types.Action(types.FileRenamedOldName)
	case 128: // File was moved to Y
		return types.Action(types.FileRenamedNewName)
	default:
		return types.Action(types.Invalid)
	}
}

// StartWatching starts a CGO function for getting the notifications
func (i *DirectoryWatcher) StartWatching(dir string) {
	log.Printf("linux.SetupDirectoryChangeNotification(): for [%s]\n", dir)
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			go watchDir(path)
		}
		return nil
	})
}

func watchDir(path string) {
	cpath := C.CString(path)
	defer func() {
		C.free(unsafe.Pointer(cpath))
		// unregister(path)
	}()
	C.WatchDirectory(cpath)
}

//export goCallbackFileChange
func goCallbackFileChange(cpath, cfile *C.char, caction C.int) {
	path := strings.TrimSpace(C.GoString(cpath))
	file := strings.TrimSpace(C.GoString(cfile))
	action := types.Action(int(caction))

	absoluteFilePath := filepath.Join(path, file)
	fi, err := fileinfo.GetFileInformation(absoluteFilePath)

	if err == nil {
		fileChangeNotifier(path, file, fi, action)
	}
}
