package local

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/otiai10/copy"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
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
		return true
	}
	return false
}

// SyncLocalFilesToBackup ...
func (s *Storage) SyncLocalFilesToBackup() {
	log.Println("storage.local.SyncLocalFilesToBackup()")
	// TODO: move this to some utils or config class, so we don't work with viper direct
	dirs, ok := viper.Get("watch").([]interface{})
	if !ok {
		log.Println("[ERROR] storage.local.SyncLocalFilesToBackup(): nothing to sync")
		return
	}

	for _, path := range dirs {
		path, ok := path.(string)
		if !ok {
			log.Println("[ERROR] storage.local.SyncLocalFilesToBackup(): invalid path")
			continue
		}

		// TODO: use filepath.Join(...)!!!
		remoteSnapshotPath := fmt.Sprintf("%s%s%s%s.snapshot",
			s.storagePath, string(os.PathSeparator),
			filepath.Base(path), string(os.PathSeparator))

		// TODO: use filepath.Join(...)!!!
		localTMPPath := fmt.Sprintf("%s%s%s%s%s%s%s%s.snapshot",
			os.TempDir(), string(os.PathSeparator),
			"bakku-app", string(os.PathSeparator),
			storageName, string(os.PathSeparator),
			filepath.Base(path), string(os.PathSeparator))

		log.Printf("storage.local.SyncLocalFilesToBackup(): copy snapshot for [%s] from [%s] to [%s]\n",
			path, remoteSnapshotPath, localTMPPath)

		if err := copy.Copy(remoteSnapshotPath, localTMPPath); err != nil {
			log.Printf("[ERROR] storage.local.SyncLocalFilesToBackup(): cannot copy snapshot for [%s]: %v\n", path, err)
			return
		}

		// TODO: use filepath.Join(...)!!!
		snapshotPath := fmt.Sprintf("%s%s.snapshot", path, string(os.PathSeparator))

		s.syncFiles(localTMPPath, snapshotPath)
	}
}

func (s *Storage) syncFiles(remoteSnapshotPath, localSnapshotPath string) {
	dbRemote, err := leveldb.OpenFile(remoteSnapshotPath, nil)
	if err != nil {
		log.Printf("[ERROR] storage.local.syncFiles(): cannot open snapshot file [%s]: leveldb.OpenFile():%v\n", remoteSnapshotPath, err)
		return
	}
	defer dbRemote.Close()

	dbLocal, err := leveldb.OpenFile(localSnapshotPath, nil)
	if err != nil {
		log.Printf("[ERROR] storage.local.syncFiles(): can not open snapshot file [%s]: leveldb.OpenFile():%v\n", localSnapshotPath, err)
		return
	}
	defer dbLocal.Close()

	iter := dbLocal.NewIterator(nil, nil)
	for iter.Next() {
		localFile := iter.Key()
		localInfo := iter.Value()
		remoteInfo, err := dbRemote.Get(localFile, nil)
		if strings.Contains(string(localFile), ".snapshot") {
			continue
		}
		if err != nil && err.Error() == "leveldb: not found" {
			log.Printf("storage.local.syncFiles(): key [%s] not found in the remote snapshot\n", string(localFile))
			continue
		}
		if string(localInfo) != string(remoteInfo) {
			log.Printf("storage.local.syncFiles(): values are different for the key [%s]: local=[%s], remote=[%s]\n",
				string(localFile), string(localInfo), string(remoteInfo))
			continue
		}
		if err != nil {
			log.Printf("[ERROR] storage.local.syncFiles(): can not get key=[%s]: dbRemote.Get(): %v\n", string(localFile), err)
		}
	}
	iter.Release()
	err = iter.Error()
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

	snapshotPath := fmt.Sprintf("%s%s%s", directoryPath, string(os.PathSeparator), ".snapshot")
	from := absolutePath
	to := fmt.Sprintf("%s%s%s%s%s",
		s.storagePath, string(os.PathSeparator),
		filepath.Base(directoryPath), string(os.PathSeparator),
		relativePath)

	// don't backup file if it is in progress
	if ok := storage.BackupStarted(absolutePath, storageName); ok {
		s.store(from, to, StoreOptions{reportProgress: true})
		storage.BackupFinished(absolutePath, storageName)
		storage.UpdateSnapshot(snapshotPath, absolutePath)
	}
}

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

	log.Printf("storage.local.store(): MkdirAll for [%s]\n", filepath.Dir(toPath))
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
