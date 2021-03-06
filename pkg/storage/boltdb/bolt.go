package boltdb

import (
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

// BoltDB ...
type BoltDB struct {
	DBFilePath string
}

// Exist checks if the database file exists
func (s BoltDB) Exist() bool {
	if _, err := os.Stat(s.DBFilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Add info about file to the snapshot, filePath is the key and bucketName is the name of the backup storage
func (s BoltDB) Add(filePath, bucketName string, value []byte) error {
	// fmt.Printf("storage.Add(): [key=%s][bucket=%s][value=%q]\n", filePath, bucketName, string(value))
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
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return b.Put([]byte(filePath), value)
	})
}

// Get information about the file from the snapshot
func (s BoltDB) Get(filePath, bucketName string) (string, error) {
	// fmt.Printf("storage.Get(): [key=%s][bucket=%s]\n", filePath, bucketName)
	if filePath == "" {
		return "", fmt.Errorf("bolt.Get(): the key(file path) is empty")
	}

	db, err := s.openDB()
	if err != nil {
		return "", err
	}

	defer func() {
		err := db.Close()
		if err != nil {
			fmt.Printf("[ERROR] db.Close() faild: %v\n", err)
		}
	}()

	var value []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket not found")
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
	db, err := bolt.Open(s.DBFilePath, 0600, nil)
	if err != nil {
		fmt.Printf("[ERROR] bolt.openDB(): can't open boltDB file [%s]: %v\n", s.DBFilePath, err)
		return nil, err
	}
	return db, nil
}
