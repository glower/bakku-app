package leveldb

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/glower/bakku-app/pkg/config/snapshot"
	snapshotStorage "github.com/glower/bakku-app/pkg/snapshot/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

// Snapshot ...
type Snapshot struct {
	snapshotPath    string // /foo/bar/.snapshot
	snapshotDirName string // .snapshot
	sameDir         bool
}

const snapshotStorageName = "leveldb"

func init() {
	snapshotStorage.Register(snapshotStorageName, &Snapshot{})
}

// Path ...
func (s *Snapshot) Path(path string) snapshotStorage.Snapshot {
	s.snapshotPath = filepath.Join(path, s.snapshotDirName)
	return s
}

// SnapshotStoragePath returns an absolute path where the snapshot data are stored
func (s *Snapshot) SnapshotStoragePath() string {
	return s.snapshotPath
}

// SnapshotStoragePathName ...
func (s *Snapshot) SnapshotStoragePathName() string {
	return s.snapshotDirName
}

// Exist ...
func (s *Snapshot) Exist() bool {
	if _, err := os.Stat(s.snapshotPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Setup new snapshot storage implementation
func (s *Snapshot) Setup() bool { // TODO: do we need bool here?
	config := snapshot.Conf()
	if config.Active {
		s.snapshotDirName = config.SnapshotDir
		s.sameDir = config.SameDir
		return true
	}
	return false
}

// Add info about file to the snapshot
func (s *Snapshot) Add(filePath string, value []byte) error {
	// log.Printf("leveldb.Add(): add [%s] to %s\n", filePath, s.snapshotPath)
	if strings.Contains(filePath, s.snapshotDirName) {
		return nil
	}
	// log.Printf("Add(): leveldb.OpenFile(): %s\n", s.snapshotPath)
	db, err := leveldb.OpenFile(s.snapshotPath, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Put([]byte(filePath), value, nil)
	if err != nil {
		return err
	}
	return nil
}

// Get information about the file from the snapshot
func (s *Snapshot) Get(file string) (string, error) {
	// log.Printf("Get(): leveldb.OpenFile(): %s\n", s.snapshotPath)
	db, err := leveldb.OpenFile(s.snapshotPath, nil)
	if err != nil {
		return "", err
	}
	defer db.Close()

	value, err := db.Get([]byte(file), nil)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

// Remove  file from the snapshot storage
func (s *Snapshot) Remove(file string) error {
	// log.Printf("Remove(): leveldb.OpenFile(): %s\n", s.snapshotPath)
	db, err := leveldb.OpenFile(s.snapshotPath, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Delete([]byte(file), nil)
	if err != nil {
		return err
	}
	return nil
}

// GetAll entries from the snapshot storage
func (s *Snapshot) GetAll() (map[string]string, error) {
	log.Printf("GetAll(): leveldb.OpenFile(): %s\n", s.snapshotPath)
	db, err := leveldb.OpenFile(s.snapshotPath, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	result := make(map[string]string)
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		filePath := string(iter.Key())
		fileInfoJSON := string(iter.Value())
		result[filePath] = fileInfoJSON
	}
	iter.Release()
	return result, nil
}
