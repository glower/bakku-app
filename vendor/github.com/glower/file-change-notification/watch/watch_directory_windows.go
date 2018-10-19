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
  	dir,   // directory to watch
		TRUE,  // do watch subtree
		FILE_NOTIFY_CHANGE_FILE_NAME | FILE_NOTIFY_CHANGE_SIZE | FILE_NOTIFY_CHANGE_DIR_NAME
	);
	ovl.hEvent = CreateEvent(NULL, TRUE, FALSE, NULL);

  if (handle == INVALID_HANDLE_VALUE){
    printf("\n ERROR: CreateFile function failed.\n");
    ExitProcess(GetLastError());
  }

  if ( handle == NULL ) {
    printf("\n ERROR: Unexpected NULL from CreateFile.\n");
    ExitProcess(GetLastError());
  }

	ReadDirectoryChangesW(handle, buffer, sizeof(buffer), FALSE, FILE_NOTIFY_CHANGE_LAST_WRITE, NULL, &ovl, NULL);

	while (TRUE) {
		waitStatus = WaitForSingleObject(ovl.hEvent, INFINITE);
		switch (waitStatus) {
      case WAIT_OBJECT_0:
				// printf("A file was created, renamed, or deleted in the directory\n");
				GetOverlappedResult(handle, &ovl, &dw, FALSE);

				char fileName[MAX_PATH] = "";
				FILE_NOTIFY_INFORMATION *fni = NULL;
				fni = (FILE_NOTIFY_INFORMATION*)(&buffer[0]);

				if (fni->Action != 0) {
					wcstombs_s(&count, fileName, sizeof(fileName),  fni->FileName, (size_t)fni->FileNameLength/sizeof(WCHAR));
					goCallbackFileChange(dir, fileName, fni->Action);
				}
				memset(fileName, '\0', sizeof(fileName));
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
	"os"
	"strings"
	"unsafe"
)

// DirectoryChangeNotification expected path to the directory to watch as string
// and a FileInfo channel for the callback notofications
// Notofication is fired each time file in the directory is changed or some new
// file (or sub-directory) is created
func DirectoryChangeNotification(path string, callbackChan chan FileChangeInfo) {

	data := CallbackData{
		CallbackChan: callbackChan,
		Path:         path,
	}

	register(data, path)

	cpath := C.CString(path)
	defer func() {
		C.free(unsafe.Pointer(cpath))
		unregister(path)
	}()
	C.WatchDirectory(cpath)
}

//export goCallbackFileChange
func goCallbackFileChange(cpath, cfile *C.char, action C.int) {

	path := strings.TrimSpace(C.GoString(cpath))
	file := strings.TrimSpace(C.GoString(cfile))
	log.Printf("goCallbackFileChange(): [%s %s], action: %d\n", path, file, action)

	filePath := strings.TrimSpace(path + file)
	fi, err := os.Stat(filePath)
	if err != nil {
		log.Printf("[ERROR] Can not stat file [%s]: %v\n", filePath, err)
		return
	}
	callbackData := lookup(path)

	// TODO: check for double events
	if fi != nil {
		callbackData.CallbackChan <- FileChangeInfo{
			Action:   Action(int(action)),
			FileInfo: fi,
		}
	}
}
