package boltdb

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/glower/bakku-app/pkg/config/snapshot"
	"github.com/glower/bakku-app/pkg/snapshot/storage"
)

const snapshotStorageName = "boltdb"

// Storage ...
type Storage struct {
	Path         string
	DBFilePath   string // /foo/bar/.snapshot
	DBFileName   string // .snapshot
	DBBucketName []byte
	SameDir      bool
}

// New returns new snapshot storage implementation
func New(path string) storage.Storager {
	conf := snapshot.Conf()
	return &Storage{
		Path:         path,
		DBFilePath:   filepath.Join(path, conf.FileName),
		DBFileName:   conf.FileName,
		DBBucketName: []byte(conf.BucketName),
		SameDir:      conf.SameDir,
	}
}

// Exist ...
func (s *Storage) Exist() bool {
	if _, err := os.Stat(s.DBFilePath); os.IsNotExist(err) {
		log.Printf("snapshot.storage.boltdb.Exist(): snapshot for [%s] is not present!\n", s.DBFilePath)
		return false
	}
	log.Printf("snapshot.storage.boltdb.Exist(): snapshot for [%s] rxist\n", s.DBFilePath)
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

// Add info about file to the snapshot
func (s *Storage) Add(filePath string, value []byte) error {
	// log.Printf("bolt.Add(): add [%s] to %s\n", filePath, s.snapshotDBPath)
	if strings.Contains(filePath, s.DBFileName) {
		return nil
	}
	db := s.openDB()
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(s.DBBucketName)
		if err != nil {
			return err
		}
		return b.Put([]byte(filePath), value)
	})
	return nil
}

// Get information about the file from the snapshot
func (s *Storage) Get(file string) (string, error) {
	db := s.openDB()
	defer db.Close()
	var value []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.DBBucketName)
		value = b.Get([]byte(file))
		return nil
	})
	return string(value), nil
}

// Remove  file from the snapshot storage
func (s *Storage) Remove(file string) error {
	// log.Printf("bolt.Remove(): remove file [%s] from snapshot\n", filePath)
	db := s.openDB()
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.DBBucketName)
		err := b.Delete([]byte(file))
		return err
	})
	return nil
}

// GetAll entries from the snapshot storage
func (s *Storage) GetAll() (map[string]string, error) {
	db := s.openDB()
	defer db.Close()
	result := make(map[string]string)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.DBBucketName)

		b.ForEach(func(k, v []byte) error {
			filePath := string(k)
			fileInfoJSON := string(v)
			result[filePath] = fileInfoJSON
			return nil
		})
		return nil
	})

	return result, nil
}

func (s *Storage) openDB() *bolt.DB {
	db, err := bolt.Open(s.DBFilePath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
