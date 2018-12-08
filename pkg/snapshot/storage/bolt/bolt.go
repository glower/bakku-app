package leveldb

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
	snapshotStorage "github.com/glower/bakku-app/pkg/snapshot/storage"
)

// Snapshot ...
type Snapshot struct {
	snapshotDBPath string // /foo/bar/.snapshot
	snapshotDBName string // .snapshot
	sameDir        bool
}

const snapshotStorageName = "boltdb"

func init() {
	log.Println("snapshot.boltdb.init()")
	snapshotStorage.Register(snapshotStorageName, &Snapshot{})
}

// Path ...
func (s *Snapshot) Path(path string) snapshotStorage.Snapshot {
	s.snapshotDBPath = filepath.Join(path, s.snapshotDBName)
	return s
}

// SnapshotStoragePath returns an absolute path where the snapshot data are stored
func (s *Snapshot) SnapshotStoragePath() string {
	return s.snapshotDBPath
}

// SnapshotStoragePathName ...
func (s *Snapshot) SnapshotStoragePathName() string {
	return s.snapshotDBName
}

// Exist ...
func (s *Snapshot) Exist() bool {
	if _, err := os.Stat(s.snapshotDBPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Setup new snapshot storage implementation
func (s *Snapshot) Setup() bool { // TODO: do we need bool here?
	s.snapshotDBName = ".snapshot"
	return true
}

// Add info about file to the snapshot
func (s *Snapshot) Add(filePath string, value []byte) error {
	// log.Printf("bolt.Add(): add [%s] to %s\n", filePath, s.snapshotDBPath)
	if strings.Contains(filePath, s.snapshotDBName) {
		return nil
	}
	db := s.openDB()
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("snapshot"))
		if err != nil {
			return err
		}
		return b.Put([]byte(filePath), value)
	})
	return nil
}

// Get information about the file from the snapshot
func (s *Snapshot) Get(file string) (string, error) {
	db := s.openDB()
	defer db.Close()
	var value []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("snapshot"))
		value = b.Get([]byte(file))
		return nil
	})
	return string(value), nil
}

// Remove  file from the snapshot storage
func (s *Snapshot) Remove(file string) error {
	log.Printf("bolt.Remove(): %s\n", s.snapshotDBPath)
	// db := s.openDB()
	// defer db.Close()

	// err := db.Delete([]byte(file), nil)
	// if err != nil {
	// 	return err
	// }
	return nil
}

// GetAll entries from the snapshot storage
func (s *Snapshot) GetAll() (map[string]string, error) {
	db := s.openDB()
	defer db.Close()
	result := make(map[string]string)

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("snapshot"))

		b.ForEach(func(k, v []byte) error {
			filePath := string(k)
			fileInfoJSON := string(v)
			result[filePath] = fileInfoJSON
			// fmt.Printf("key=%s, value=%s\n", k, v)
			return nil
		})
		return nil
	})

	return result, nil
}

func (s *Snapshot) openDB() *bolt.DB {
	db, err := bolt.Open(s.snapshotDBPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
