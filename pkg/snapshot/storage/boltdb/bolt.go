package boltdb

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/glower/bakku-app/pkg/config/snapshot"
	"github.com/glower/bakku-app/pkg/snapshot/storage"
)

// Storage ...
type Storage struct {
	path       string // /foo/bar
	DBFilePath string // /foo/bar/.snapshot
	DBFileName string // .snapshot
}

// New returns new snapshot storage implementation
func New(path string) storage.Storage {
	conf := snapshot.Conf()
	return &Storage{
		path:       path,
		DBFilePath: filepath.Join(path, conf.FileName),
		DBFileName: conf.FileName,
	}
}

// Exist ...
func (s *Storage) Exist() bool {
	if _, err := os.Stat(s.DBFilePath); os.IsNotExist(err) {
		log.Printf("snapshot.storage.boltdb.Exist(): snapshot for [%s] is not present!\n", s.DBFilePath)
		return false
	}
	return true
}

// FilePath ...
func (s *Storage) FilePath() string {
	return s.DBFilePath
}

// FileName ...
func (s *Storage) FileName() string {
	return s.DBFileName
}

// Path ...
func (s *Storage) Path() string {
	return s.path
}

// Add info about file to the snapshot, filePath is the key and bucketName is the name of the backup storage
func (s *Storage) Add(filePath, bucketName string, value []byte) error {
	// log.Printf("bolt.Add(): file=[%s], bucketName=[%s] DBFilePath=[%s]\n", filePath, bucketName, s.DBFilePath)
	if strings.Contains(filePath, s.DBFileName) {
		return nil
	}
	db := s.openDB()
	defer db.Close()
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		err = b.Put([]byte(filePath), value)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

// Get information about the file from the snapshot
func (s *Storage) Get(filePath, bucketName string) (string, error) {
	// log.Printf("bolt.Get(): file=[%s], bucketName=[%s] DBFilePath=[%s]\n", filePath, bucketName, s.DBFilePath)
	if filePath == "" {
		return "", fmt.Errorf("bolt.Get(): the key(file path) is empty")
	}
	db := s.openDB()
	defer db.Close()
	var value []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bolt.Get(): bucket [%s] not found", bucketName)
		}
		value = b.Get([]byte(filePath))
		return nil
	})
	if err != nil {
		return "", err
	}
	return string(value), nil
}

// Remove file from the snapshot storage
func (s *Storage) Remove(filePath, bucketName string) error {
	db := s.openDB()
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		err := b.Delete([]byte(filePath))
		return err
	})
}

// GetAll entries from the snapshot storage
func (s *Storage) GetAll(bucketName string) (map[string]string, error) {
	db := s.openDB()
	defer db.Close()
	result := make(map[string]string)

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bolt.Get(): bucket [%s] not found", bucketName)
		}
		b.ForEach(func(k, v []byte) error {
			filePath := string(k)
			fileInfoJSON := string(v)
			result[filePath] = fileInfoJSON
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Storage) openDB() *bolt.DB {
	db, err := bolt.Open(s.DBFilePath, 0600, nil)
	if err != nil {
		log.Printf("[ERROR] bolt.openDB(): can't open boltDB file [%s]\n", s.DBFilePath)
		log.Fatal(err)
	}
	return db
}
