package snapshot

import (
	"log"

	"github.com/spf13/viper"
)

const defaultStorageImplementationName = "boltdb"

// DefaultStorage returns name of a default storage implementation as a string
func DefaultStorage() string {
	defaultStorage, ok := viper.Get("snapshot.default").(string)
	if !ok {
		log.Printf("DefaultStorage(): can't find [%s]\n", defaultStorage)
		defaultStorage = defaultStorageImplementationName
	}
	return defaultStorage
}
