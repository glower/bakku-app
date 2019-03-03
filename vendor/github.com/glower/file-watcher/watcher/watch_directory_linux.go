// +build linux,!integration

package watcher

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

extern void goCallbackFileChange(char* root, char* path, char* file, int action);

static inline void *WatchDirectory(char* root, char* dir) {
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
			goCallbackFileChange(root, dir, event->name, event->mask);
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

	"github.com/glower/file-watcher/notification"
	fileinfo "github.com/glower/file-watcher/util"
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
func convertMaskToAction(mask int) notification.ActionType {
	switch mask {
	case 2, 8: // File was modified
		return notification.ActionType(notification.FileModified)
	case 256: // Subfile was created
		return notification.ActionType(notification.FileAdded)
	case 512: // Subfile was deleted
		return notification.ActionType(notification.FileRemoved)
	case 64: // File was moved from X
		return notification.ActionType(notification.FileRenamedOldName)
	case 128: // File was moved to Y
		return notification.ActionType(notification.FileRenamedNewName)
	default:
		return notification.ActionType(notification.Invalid)
	}
}

// StartWatching starts a CGO function for getting the notifications
func (i *DirectoryWatcher) StartWatching(root string) {
	log.Printf("linux.StartWatching(): for [%s]\n", root)
	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			go watchDir(root, path)
		}
		return nil
	})
	if err != nil {
		fileError("ERROR", err)
	}
}

func watchDir(rootDirToWatch string, subDir string) {
	croot := C.CString(rootDirToWatch)
	cdir := C.CString(subDir)
	defer func() {
		C.free(unsafe.Pointer(croot))
		C.free(unsafe.Pointer(cdir))
	}()
	C.WatchDirectory(croot, cdir)
}

//export goCallbackFileChange
func goCallbackFileChange(croot, cpath, cfile *C.char, caction C.int) {
	root := strings.TrimSpace(C.GoString(croot))
	path := strings.TrimSpace(C.GoString(cpath))
	file := strings.TrimSpace(C.GoString(cfile))
	action := convertMaskToAction(int(caction))

	absoluteFilePath := filepath.Join(path, file)
	relativeFilePath, err := filepath.Rel(root, absoluteFilePath)
	if err != nil {
		fileError("ERROR", err)
		return
	}

	fi, err := fileinfo.GetFileInformation(absoluteFilePath)
	if err != nil {
		fileError("WARN", err)
		return
	}

	fileChangeNotifier(root, relativeFilePath, fi, action)
}
