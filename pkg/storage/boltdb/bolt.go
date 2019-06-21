package boltdb

import (
	"fmt"
	"log"
	"os"

	bolt "go.etcd.io/bbolt"
)

// Storage ...
type Storage struct {
	DBFilePath string // /foo/bar/.snapshot
}

// Exist ...
func (s *Storage) Exist() bool {
	if _, err := os.Stat(s.DBFilePath); os.IsNotExist(err) {
		log.Printf("[INFO] snapshot.storage.boltdb.Exist(): snapshot for [%s] is not present!\n", s.DBFilePath)
		return false
	}
	return true
}

// Add info about file to the snapshot, filePath is the key and bucketName is the name of the backup storage
func (s *Storage) Add(filePath, bucketName string, value []byte) error {
	db := s.openDB() // TODO: refactor me
	defer db.Close() // handle this error
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

// GetAll return all entries from the snapshot storage for one bucket
func (s *Storage) GetAll(bucketName string) (map[string]string, error) {
	db := s.openDB()
	defer db.Close()
	result := make(map[string]string)

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bolt.Get(): bucket [%s] not found", bucketName)
		}
		return b.ForEach(func(k, v []byte) error {
			filePath := string(k)
			fileInfoJSON := string(v)
			result[filePath] = fileInfoJSON
			return nil
		})
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
