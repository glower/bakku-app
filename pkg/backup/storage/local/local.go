package local

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"

	"github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/spf13/viper"
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
	storage.Register(storageName, &Storage{})
}

// Setup local storage
func (s *Storage) Setup(fileStorageProgressCannel chan *storage.Progress) bool {
	if isStorageConfigured() {
		log.Println("storage.local.Setup()")
		s.name = storageName
		s.fileChangeNotificationChannel = make(chan *storage.FileChangeNotification)
		s.fileStorageProgressCannel = fileStorageProgressCannel
		s.path = viper.Get("backup.local.path").(string)
		return true
	}
	return false
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
	// log.Printf("storage.local.handleFileChanges(): File [%#v] has been changed\n", fileChange)
	file := fileChange.Name
	path := fileChange.AbsolutePath
	storage.BackupStarted(file, storageName)
	s.store(path, file)
	storage.BackupFinished(file, storageName)
}

func (s *Storage) store(path, file string) {
	log.Printf(">>> Copy file from p=%s to %s%s", path, s.path, file)
	from, err := os.Open(path)
	if err != nil {
		log.Printf("[ERROR] storage.local.handleFileChanges(): Cannot open file  [%s]: %v\n", path, err)
		return
	}
	defer from.Close()
	fromStrats, _ := from.Stat()
	readBuffer := bufio.NewReader(from)
	totalSize := fromStrats.Size()

	to, err := os.OpenFile(s.path+file, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("[ERROR] storage.local.handleFileChanges(): Cannot open file [%s] to write: %v\n", s.path+file, err)
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
		var percent float64
		if int64(written) == totalSize {
			percent = float64(100)
		} else {
			percent = float64(100 * int64(totalWritten) / totalSize)
		}
		progress := &storage.Progress{
			StorageName: storageName,
			FileName:    file,
			Percent:     percent,
		}
		s.fileStorageProgressCannel <- progress
	}

	if err = writeBuffer.Flush(); err != nil {
		panic(err)
	}
}

func isStorageConfigured() bool {
	isActive, ok := viper.Get("backup.local.active").(bool)
	if !ok {
		log.Printf("isStorageConfigured(): is not active: %v\n", isActive)
		return false
	}
	_, ok = viper.Get("backup.local.path").(string)
	if !ok {
		return false
	}
	return isActive
}
