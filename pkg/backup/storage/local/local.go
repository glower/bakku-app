package local

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"

	"github.com/glower/bakku-app/pkg/backup/storage"
)

// Storage local
type Storage struct {
	name                          string // storage name
	fileChangeNotificationChannel chan *storage.FileChangeNotification
	fileStorageProgressCannel     chan *storage.Progress
	ctx                           context.Context
	path                          string
}

const storageName = "local"
const bufferSize = 1024 * 1024

func init() {
	log.Println("storage.local.init()")
	storage.Register(storageName, &Storage{})
}

// Setup local storage
func (s *Storage) Setup(fileStorageProgressCannel chan *storage.Progress) bool {
	log.Println("storage.local.Setup()")
	s.name = storageName
	s.fileChangeNotificationChannel = make(chan *storage.FileChangeNotification)
	s.fileStorageProgressCannel = fileStorageProgressCannel
	s.path = "/home/igor/storage/" // TODO: need to read this from the config or something
	return true
}

// FileChangeNotification returns channel for notifications
func (s *Storage) FileChangeNotification() chan *storage.FileChangeNotification {
	return s.fileChangeNotificationChannel
}

// Start local storage
func (s *Storage) Start(ctx context.Context) error {
	log.Println("storage.local.Start()")
	s.ctx = ctx
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case fileChange := <-s.fileChangeNotificationChannel:
				go s.handleFileChanges(fileChange)
			}
		}
	}()
	return nil
}

func (s *Storage) handleFileChanges(fileChange *storage.FileChangeNotification) {
	// log.Printf("storage.local.handleFileChanges(): File [%#v] has been changed\n", fileChange.File)
	file := fileChange.File.Name()
	path := fileChange.Path
	storage.BackupStarted(file, storageName)
	s.store(path, file)
	storage.BackupFinished(file, storageName)
}

func (s *Storage) store(path, file string) {
	// log.Printf(">>> Copy file from p=%s to %s%s", path, s.path, file)
	from, err := os.Open(path)
	if err != nil {
		log.Fatalf("Cannot open file  [%s]: %v\n", path, err)
	}
	defer from.Close()
	fromStrats, _ := from.Stat()
	readBuffer := bufio.NewReader(from)
	totalSize := fromStrats.Size()

	to, err := os.OpenFile(s.path+file, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Cnnot open file [%s] to write: %v\n", s.path+file, err)
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
		if _, err := writeBuffer.Write(buf[:n]); err != nil {
			panic(err)
		}
		totalWritten = totalWritten + bufferSize
		progress := &storage.Progress{
			StorageName: storageName,
			FileName:    file,
			Percent:     float64(100 * int64(totalWritten) / totalSize),
		}
		s.fileStorageProgressCannel <- progress
	}

	if err = writeBuffer.Flush(); err != nil {
		panic(err)
	}
}
