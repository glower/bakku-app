package leveldb

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/glower/bakku-app/pkg/config/snapshot"
	"github.com/glower/bakku-app/pkg/snapshot/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

// Snapshot ...
type Snapshot struct {
	snapshotPath    string
	snapshotDirName string // .snapshot
	sameDir         bool
}

const snapshotStorageName = "leveldb"

func init() {
	storage.Register(snapshotStorageName, &Snapshot{})
}

// Path ...
func (s *Snapshot) Path(path string) storage.Snapshot {
	s.snapshotPath = filepath.Join(path, s.snapshotDirName)
	return s
}

// SnapshotStoragePath returns a path where the snapshot data are stored
func (s *Snapshot) SnapshotStoragePath() string {
	return filepath.Join(s.snapshotPath, s.snapshotDirName)
}

// Exist ...
func (s *Snapshot) Exist() bool {
	if _, err := os.Stat(s.snapshotPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Register new snapshot storage implementation
func (s *Snapshot) Register() bool {
	// config := snapshot.Config("leveldb")
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
	// we don't add snapshot files to the snapshot
	if strings.Contains(filePath, s.snapshotDirName) {
		return nil
	}
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
	db, err := leveldb.OpenFile(s.snapshotPath, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// x := map[type]type
	result := make(map[string]string)
	// dir := s.snapshotPath
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		filePath := string(iter.Key())
		fileInfoJSON := string(iter.Value())
		result[filePath] = fileInfoJSON
	}
	iter.Release()
	return result, nil
}
