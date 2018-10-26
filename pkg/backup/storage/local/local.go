package local

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/otiai10/copy"
	"github.com/spf13/viper"
)

// Storage local
type Storage struct {
	name                          string // storage name
	fileChangeNotificationChannel chan *storage.FileChangeNotification
	fileStorageProgressCannel     chan *storage.Progress
	ctx                           context.Context
	storagePath                   string
}

const storageName = "local"
const bufferSize = 1024 * 1024

func init() {
	storage.Register(storageName, &Storage{})
}

// StoreOptions ...
type StoreOptions struct {
	reportProgress bool
}

// Setup local storage
func (s *Storage) Setup(fileStorageProgressCannel chan *storage.Progress) bool {
	if isStorageConfigured() {
		log.Println("storage.local.Setup()")
		s.name = storageName
		s.fileChangeNotificationChannel = make(chan *storage.FileChangeNotification)
		s.fileStorageProgressCannel = fileStorageProgressCannel
		storagePath := filepath.Clean(viper.Get("backup.local.path").(string))
		s.storagePath = storagePath

		go s.SyncLocalFilesToBackup()

		return true
	}
	return false
}

// SyncLocalFilesToBackup ...
func (s *Storage) SyncLocalFilesToBackup() {

	// TODO: move this to some utils or config class, so we don't work with viper direct
	dirs, ok := viper.Get("watch").([]interface{})
	if !ok {
		log.Println("SyncLocalFilesToBackup(): nothing to sync")
		return
	}

	for _, path := range dirs {
		path, ok := path.(string)
		if !ok {
			log.Println("SyncLocalFilesToBackup(): invalid path")
			continue
		}

		log.Printf("SyncLocalFilesToBackup(): [%s]\n", path)
		remoteSnapshotPath := fmt.Sprintf("%s%s%s/.snapshot", s.storagePath, string(os.PathSeparator), filepath.Base(path))
		localTMPPath := fmt.Sprintf("%s%s%s%s%s%s%s/.snapshot",
			os.TempDir(), string(os.PathSeparator),
			"bakku-app", string(os.PathSeparator),
			storageName, string(os.PathSeparator),
			filepath.Base(path))

		// s.get(remoteSnapshotPath, os.TempDir)
		log.Printf("SyncLocalFilesToBackup(): copy snapshot for [%s] from [%s] to [%s]\n",
			path, remoteSnapshotPath, localTMPPath)
		if err := copy.Copy(remoteSnapshotPath, localTMPPath); err != nil {
			log.Printf("[ERROR] SyncLocalFilesToBackup(): cannot copy snapshot for [%s]: %v\n", path, err)
			return
		}

	}
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
	absolutePath := fileChange.AbsolutePath
	relativePath := fileChange.RelativePath
	directoryPath := fileChange.DirectoryPath

	from := absolutePath
	to := fmt.Sprintf("%s%s%s%s%s",
		s.storagePath, string(os.PathSeparator),
		filepath.Base(directoryPath), string(os.PathSeparator),
		relativePath)

	storage.BackupStarted(absolutePath, storageName)
	s.store(from, to, StoreOptions{reportProgress: true})
	storage.BackupFinished(absolutePath, storageName)
}

// get remote file from the storage
func (s *Storage) get(fromPath, toPath string) {
	s.store(fromPath, toPath, StoreOptions{reportProgress: false})
}

func (s *Storage) store(fromPath, toPath string, opt StoreOptions) {
	log.Printf(">>> Copy file from [%s] to [%s]\n", fromPath, toPath)
	from, err := os.Open(fromPath)
	if err != nil {
		log.Printf("[ERROR] storage.local.handleFileChanges(): Cannot open file  [%s]: %v\n", fromPath, err)
		return
	}
	defer from.Close()
	fromStrats, _ := from.Stat()
	readBuffer := bufio.NewReader(from)
	totalSize := fromStrats.Size()

	// func MkdirAll(path string, perm FileMode) error
	if err := os.MkdirAll(filepath.Dir(toPath), 0744); err != nil {
		log.Printf("[ERROR] storage.local.handleFileChanges():  MkdirAll for [%s], %v", filepath.Dir(toPath), err)
		return
	}

	to, err := os.OpenFile(toPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("[ERROR] storage.local.handleFileChanges(): Cannot open file [%s] to write: %v\n", toPath, err)
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
			var percent float64
			if int64(written) == totalSize {
				percent = float64(100)
			} else {
				percent = float64(100 * int64(totalWritten) / totalSize)
			}
			progress := &storage.Progress{
				StorageName: storageName,
				FileName:    from.Name(),
				Percent:     percent,
			}
			s.fileStorageProgressCannel <- progress
		}
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
