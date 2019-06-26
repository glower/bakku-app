package boltdb

import (
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

// BoltDB ...
type BoltDB struct {
	DBFilePath string
}

// Exist ...
func (s BoltDB) Exist() bool {
	if _, err := os.Stat(s.DBFilePath); os.IsNotExist(err) {
		log.Printf("[INFO] snapshot.storage.boltdb.Exist(): snapshot for [%s] is not present!\n", s.DBFilePath)
		return false
	}
	return true
}

// Add info about file to the snapshot, filePath is the key and bucketName is the name of the backup storage
func (s BoltDB) Add(filePath, bucketName string, value []byte) error {
	fmt.Printf("storage.Add(): [key=%s][bucket=%s][value=%q]\n", filePath, bucketName, string(value))
	db, err := s.openDB()
	if err != nil {
		return err
	}
	defer func() {
		err := db.Close() // handle this error
		if err != nil {
			fmt.Printf("[ERROR] db.Close() faild: %v\n", err)
		}
	}()
	return db.Update(func(tx *bolt.Tx) error {
		fmt.Printf("CreateBucketIfNotExists b=%s\n", bucketName)
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		fmt.Printf("Put k=%s\n", filePath)
		return b.Put([]byte(filePath), value)
	})
}

// Get information about the file from the snapshot
func (s BoltDB) Get(filePath, bucketName string) ([]byte, error) {
	fmt.Printf("storage.Get(): [key=%s][bucket=%s]\n", filePath, bucketName)
	if filePath == "" {
		return nil, fmt.Errorf("bolt.Get(): the key(file path) is empty")
	}
	db, err := s.openDB()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := db.Close() // handle this error
		if err != nil {
			fmt.Printf("[ERROR] db.Close() faild: %v\n", err)
		}
	}()
	var value []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		value = b.Get([]byte(filePath))
		return nil
	})
	// if err != nil {
	// 	return nil, err
	// }
	fmt.Printf("Get done: v=%s\n", value)
	return value, nil
}

// Remove file from the snapshot storage
func (s BoltDB) Remove(filePath, bucketName string) error {
	db, err := s.openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		err := b.Delete([]byte(filePath))
		return err
	})
}

// GetAll return all entries from the snapshot storage for one bucket
func (s BoltDB) GetAll(bucketName string) (map[string]string, error) {
	db, _ := s.openDB()
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

func (s BoltDB) openDB() (*bolt.DB, error) {
	fmt.Printf("openDB: %s\n", s.DBFilePath)
	db, err := bolt.Open(s.DBFilePath, 0600, nil)
	if err != nil {
		fmt.Printf("[ERROR] bolt.openDB(): can't open boltDB file [%s]: %v\n", s.DBFilePath, err)
		return nil, err
	}
	return db, nil
}
