package local

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/glower/bakku-app/pkg/types"
)

func (s *Storage) store(fromPath, toPath string, opt StoreOptions) error {
	log.Printf("storage.local.store(): Copy file from [%s] to [%s]\n", fromPath, toPath)
	from, err := os.Open(fromPath)
	if err != nil {
		return fmt.Errorf("cannot open file  [%s]: %v", fromPath, err)
	}
	defer from.Close()
	fromStrats, _ := from.Stat()
	readBuffer := bufio.NewReader(from)
	totalSize := fromStrats.Size()

	fileStoragePath := filepath.Dir(toPath)
	if err := os.MkdirAll(fileStoragePath, 0744); err != nil {
		return fmt.Errorf("mkdirAll for path: [%s] err: %v", fileStoragePath, err)
	}

	to, err := os.OpenFile(toPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("cannot open file [%s] to write: %v", toPath, err)
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

		if s.addLatency {
			sleepRandom()
		}

		if opt.reportProgress {
			s.reportProgress(int64(written), int64(totalSize), int64(totalWritten), from.Name())
		}
	}

	if err = writeBuffer.Flush(); err != nil {
		return fmt.Errorf("cannot write buffer: %v", err)
	}
	return nil
}

func sleepRandom() {
	r := 500000 + rand.Intn(2000000)
	time.Sleep(time.Duration(r) * time.Microsecond)
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
