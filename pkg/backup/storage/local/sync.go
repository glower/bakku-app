package local

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/types"
)

// // SyncSnapshot syncs the file to the storage
// func (s *Storage) SyncSnapshot(fileChange *types.FileChangeNotification) {
// 	directoryPath := fileChange.DirectoryPath
// 	remoteSnapshotPath := filepath.Join(s.storagePath, fileChange.WatchDirectoryName, snapshot.FileName(directoryPath))
// 	localSnapshotPath := snapshot.FilePath(directoryPath)
// 	s.store(localSnapshotPath, remoteSnapshotPath, StoreOptions{reportProgress: false})
// }

// func (s *Storage) syncFiles(remoteSnapshotPath, localSnapshotPath string) {
// 	log.Printf("syncFiles(): from remote: [%s] to local [%s]\n", remoteSnapshotPath, localSnapshotPath)
// 	files, err := snapshot.Diff(remoteSnapshotPath, localSnapshotPath)
// 	if err != nil {
// 		log.Printf("[ERROR] storage.local.syncFiles(): %v\n", err)
// 		return
// 	}
// 	for _, file := range *files {
// 		s.fileChangeNotificationChannel <- &file
// 	}
// }

// get remote file from the storage
func (s *Storage) get(fromPath, toPath string) {
	s.store(fromPath, toPath, StoreOptions{reportProgress: false})
}

func (s *Storage) store(fromPath, toPath string, opt StoreOptions) {
	log.Printf("storage.local.store(): Copy file from [%s] to [%s]\n", fromPath, toPath)
	from, err := os.Open(fromPath)
	if err != nil {
		log.Printf("[ERROR] storage.local.store(): Cannot open file  [%s]: %v\n", fromPath, err)
		return
	}
	defer from.Close()
	fromStrats, _ := from.Stat()
	readBuffer := bufio.NewReader(from)
	totalSize := fromStrats.Size()

	if err := os.MkdirAll(filepath.Dir(toPath), 0744); err != nil {
		log.Printf("[ERROR] storage.local.handleFileChanges():  MkdirAll for [%s], %v", filepath.Dir(toPath), err)
		return
	}

	to, err := os.OpenFile(toPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("[ERROR] storage.local.store(): Cannot open file [%s] to write: %v\n", toPath, err)
		return
	}
	defer to.Close()
	writeBuffer := bufio.NewWriter(to)

	totalWritten := 0
	buf := make([]byte, bufferSize)
	for {
		// read a chunk
		n, err := readBuffer.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		// write a chunk
		var written = 0
		if written, err = writeBuffer.Write(buf[:n]); err != nil {
			panic(err)
		}
		totalWritten = totalWritten + written

		if opt.reportProgress {
			s.reportProgress(int64(written), int64(totalSize), int64(totalWritten), from.Name())
		}
	}

	if err = writeBuffer.Flush(); err != nil {
		panic(err)
	}
}

// // BackupProgress represents a moment of progress.
// type BackupProgress struct {
// 	StorageName string  `json:"storage"`
// 	FileName    string  `json:"file"`
// 	Percent     float64 `json:"percent"`
// }
func (s *Storage) reportProgress(written, totalSize, totalWritten int64, name string) {
	var percent float64
	if int64(written) == totalSize {
		percent = float64(100)
	} else {
		percent = float64(100 * int64(totalWritten) / totalSize)
	}

	s.fileStorageProgressCannel <- types.BackupProgress{
		StorageName: storageName,
		FileName:    name,
		Percent:     percent,
	}
}
