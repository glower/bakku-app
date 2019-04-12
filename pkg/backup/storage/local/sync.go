package local

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/types"
)

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

	fileStoragePath := filepath.Dir(toPath)
	if err := os.MkdirAll(fileStoragePath, 0744); err != nil {
		log.Printf("[ERROR] storage.local.store():  MkdirAll for path: [%s] err: %v", fileStoragePath, err)
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
		var written int
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

func (s *Storage) reportProgress(written, totalSize, totalWritten int64, name string) {
	var percent float64
	if int64(written) == totalSize {
		percent = float64(100)
	} else {
		percent = float64(100 * int64(totalWritten) / totalSize)
	}

	s.fileStorageProgressCh <- types.BackupProgress{
		StorageName: storageName,
		FileName:    name,
		Percent:     percent,
	}
}
