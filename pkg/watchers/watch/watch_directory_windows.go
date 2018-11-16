// +build windows

package watch

/*
#include <stdio.h>
#include <windows.h>
#include <stdlib.h>

#define BUFFER_SIZE 1024

extern void goCallbackFileChange(char* path, char* file, int action);

// Install https://sourceforge.net/projects/mingw-w64/ to compile (x86_64-8.1.0-posix-seh-rt_v6-rev0)
// For the API documentation see:
// https://msdn.microsoft.com/de-de/library/windows/desktop/aa365261(v=vs.85).aspx
// https://docs.microsoft.com/en-us/windows/desktop/api/fileapi/nf-fileapi-findfirstchangenotificationa
static inline void WatchDirectory(char* dir) {
	printf("[CGO] [INFO] WatchDirectory(): %s\n" ,dir);
	HANDLE handle;
	size_t  count;
	DWORD waitStatus;
	DWORD dw;
	OVERLAPPED ovl = { 0 };
	char buffer[1024];

	// FILE_NOTIFY_CHANGE_FILE_NAME  – File creating, deleting and file name changing
	// FILE_NOTIFY_CHANGE_DIR_NAME   – Directories creating, deleting and file name changing
	// FILE_NOTIFY_CHANGE_ATTRIBUTES – File or Directory attributes changing
	// FILE_NOTIFY_CHANGE_SIZE       – File size changing
	// FILE_NOTIFY_CHANGE_LAST_WRITE – Changing time of write of the files
	// FILE_NOTIFY_CHANGE_SECURITY   – Changing in security descriptors
	handle = FindFirstChangeNotification(
  	dir,   		// directory to watch
		TRUE,  		// do watch subtree
		FILE_NOTIFY_CHANGE_FILE_NAME | FILE_NOTIFY_CHANGE_SIZE | FILE_NOTIFY_CHANGE_DIR_NAME
	);
	ovl.hEvent = CreateEvent(
		NULL,  		// default security attribute
		TRUE,  		// manual reset event
		FALSE, 		// initial state = signaled
		NULL); 		// unnamed event object

  if (handle == INVALID_HANDLE_VALUE){
    printf("[CGO] [ERROR] WatchDirectory(): FindFirstChangeNotification function failed for directroy [%s] with error [%s]\n", dir, GetLastError());
    ExitProcess(GetLastError());
  }

  if ( handle == NULL ) {
    printf("[CGO] ERROR WatchDirectory(): Unexpected NULL from CreateFile for directroy [%s]\n", dir);
    ExitProcess(GetLastError());
  }

	ReadDirectoryChangesW(handle, buffer, sizeof(buffer), FALSE, FILE_NOTIFY_CHANGE_LAST_WRITE, NULL, &ovl, NULL);

	while (TRUE) {
		waitStatus = WaitForSingleObject(ovl.hEvent, INFINITE);
		switch (waitStatus) {
      case WAIT_OBJECT_0:
				// printf("[CGO] [INFO] A file was created, renamed, or deleted\n");
				GetOverlappedResult(
					handle,  // pipe handle
					&ovl, 	 // OVERLAPPED structure
					&dw,     // bytes transferred
					FALSE);  // does not wait

				char fileName[MAX_PATH] = "";
				FILE_NOTIFY_INFORMATION *fni = NULL;
				DWORD offset = 0;

				do {
					fni = (FILE_NOTIFY_INFORMATION*)(&buffer[offset]);
					wcstombs_s(&count, fileName, sizeof(fileName),  fni->FileName, (size_t)fni->FileNameLength/sizeof(WCHAR));
					// printf("[CGO] [INFO] file=[%s] action=[%d] offset=[%ld]\n", fileName, fni->Action, offset);
					goCallbackFileChange(dir, fileName, fni->Action);
					memset(fileName, '\0', sizeof(fileName));
        	offset += fni->NextEntryOffset;
				} while (fni->NextEntryOffset != 0);

				ResetEvent(ovl.hEvent);
				if( ReadDirectoryChangesW( handle, buffer, sizeof(buffer), FALSE, FILE_NOTIFY_CHANGE_LAST_WRITE, NULL, &ovl, NULL) == 0) {
					printf("Reading Directory Change");
				}
        break;
      case WAIT_TIMEOUT:
        printf("\nNo changes in the timeout period.\n");
        break;
      default:
        printf("\n ERROR: Unhandled status.\n");
        ExitProcess(GetLastError());
        break;
    }
  }
}
*/
import "C"
import (
	"log"
	"strings"
	"unsafe"

	"github.com/glower/bakku-app/pkg/types"
)

type DirectoryChangeWacherImplementer struct{}

// SetupDirectoryChangeNotification ...
func (i *DirectoryChangeWacherImplementer) SetupDirectoryChangeNotification(path string) {
	cpath := C.CString(path)
	defer func() {
		C.free(unsafe.Pointer(cpath))
		unregister(path)
	}()
	C.WatchDirectory(cpath)
}

//export goCallbackFileChange
func goCallbackFileChange(cpath, cfile *C.char, caction C.int) {
	path := strings.TrimSpace(C.GoString(cpath))
	file := strings.TrimSpace(C.GoString(cfile))
	fileChangeNotifier(path, file, types.Action(int(caction)))
}

func lookup(path string) CallbackData {
	// log.Printf("watch.lookup(): %s\n", path)
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	data, ok := callbackFuncs[path]
	if !ok {
		log.Printf("watch.lookup(): callback data for path=%s are not found!!!\n", path)
	}
	return data
}
