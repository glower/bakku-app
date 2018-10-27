package local

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/otiai10/copy"
	"github.com/syndtr/goleveldb/leveldb"
)

// SyncSnapshot syncs the snapshot dir to the storage
func (s *Storage) SyncSnapshot(from, to string) {
	// log.Printf("storage.local.SyncSnapshot(): copy snapshot form [%s] to [%s]\n", from, to)
	if err := copy.Copy(from, to); err != nil {
		log.Printf("[ERROR] storage.local.SyncSnapshot(): cannot copy snapshot for [%s]: %v\n", from, err)
		return
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
			s.syncFile(localInfo)
			continue
		}
		if string(localInfo) != string(remoteInfo) {
			log.Printf("storage.local.syncFiles(): values are different for the key [%s]\n", string(localFile))
			s.syncFile(localInfo)
			continue
		}
		if err != nil {
			log.Printf("[ERROR] storage.local.syncFiles(): can not get key=[%s]: dbRemote.Get(): %v\n", string(localFile), err)
		}
	}
	iter.Release()
}

func (s *Storage) syncFile(localInfo []byte) {
	log.Printf("storage.local.syncFile()\n")
	change := types.FileChangeNotification{}
	if err := json.Unmarshal(localInfo, &change); err != nil {
		log.Printf("storage.local.syncFile(): cannot unmarshal data [%s]: %v\n", string(localInfo), err)
	}
	s.fileChangeNotificationChannel <- &change
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

func (s *Storage) reportProgress(written, totalSize, totalWritten int64, name string) {
	var percent float64
	if int64(written) == totalSize {
		percent = float64(100)
	} else {
		percent = float64(100 * int64(totalWritten) / totalSize)
	}

	progress := &storage.Progress{
		StorageName: storageName,
		FileName:    name,
		Percent:     percent,
	}
	s.fileStorageProgressCannel <- progress
}
