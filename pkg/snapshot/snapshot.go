package snapshot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

const snapshotDirName = ".snapshot"
const appName = "bakku-app"

// Dir ...
func Dir() string {
	return filepath.Join(snapshotDirName)
}

// AppName ...
func AppName() string {
	return appName
}

// Exist ...
func Exist(snapshotPath string) bool {
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Update ...
func Update(dir, snapshotPath string) {
	log.Printf("UpdateSnapshot(): %s\n", snapshotPath)
	db, err := leveldb.OpenFile(snapshotPath, nil)
	if err != nil {
		log.Printf("watchers.UpdateSnapshot(): can not open snapshot file [%s]: %v\n", snapshotPath, err)
		return
	}
	defer db.Close()
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if strings.Contains(path, Dir()) {
			return nil
		}
		if !f.IsDir() {
			key := path
			value := fmt.Sprintf("%s:%d", f.ModTime(), f.Size())
			db.Put([]byte(key), []byte(value), nil)
		}
		return nil
	})
	log.Println("UpdateSnapshot(): done")
}

// CreateFirstBackup ...
func CreateFirstBackup(dir, snapshotPath string, changes chan storage.FileChangeNotification) {
	log.Printf("watchers.CreateFirstBackup(): for=[%s]\n", dir)
	db, err := leveldb.OpenFile(snapshotPath, nil)
	if err != nil {
		log.Printf("[ERROR] watchers.CreateFirstBackup(): can not open snapshot file [%s]: %v\n", snapshotPath, err)
		return
	}
	defer db.Close()

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		path := iter.Key()
		filePath := string(path)
		if strings.Contains(filePath, Dir()) {
			continue
		}
		fileName := filepath.Base(filePath)
		relativePath := strings.Replace(filePath, dir+string(os.PathSeparator), "", -1)
		// fmt.Printf("!!!!!!!!!!!! CreateFirstBackup(): relativePath=[%s]\n", relativePath)
		changes <- storage.FileChangeNotification{
			AbsolutePath:  filePath,
			RelativePath:  relativePath,
			DirectoryPath: dir,
			Name:          fileName,
		}
	}
	iter.Release()
	// go watch.DirectoryChangeNotification(dir, changes)
}
