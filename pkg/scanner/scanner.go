package scanner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/syndtr/goleveldb/leveldb"
)

// Scanner ...
type Scanner struct {
}

// UpdateSnapshot ...
func (s *Scanner) UpdateSnapshot(path string) {
	// TODO: put leveldb to Scanner
	db, _ := leveldb.OpenFile(path+".snapshot", nil)
	defer db.Close()
	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			key := path
			value := fmt.Sprintf("%s:%d", f.ModTime(), f.Size())
			db.Put([]byte(key), []byte(value), nil)
		}
		return nil
	})
}
