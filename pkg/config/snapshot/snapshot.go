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

// Config is a struct for basic leveldb configuration
type Config struct {
	SameDir    bool
	BucketName string
	FileName   string
}

const defaultSnapshotDBName = ".snapshot"
const defaultBucketName = "snapshot"

// snapshot:
//   sameDir:    true
//   bucketName: snapshot
//   fileName:   .snapshot

// Conf ...
func Conf() *Config {
	var conf Config
	fileName, ok := viper.Get("snapshot.fileName").(string)
	if !ok {
		log.Printf("SnapshotConf(): can't find [snapshot.fileNamer] in the config file, using default value [%s]\n", defaultSnapshotDBName)
		fileName = defaultSnapshotDBName
	}
	conf.FileName = fileName

	bucketName, ok := viper.Get("snapshot.bucketName").(string)
	if !ok {
		log.Printf("snapshot.Conf(): can't find [snapshot.leveldb.active] in the config file, using default value [%s]\n", defaultBucketName)
		bucketName = defaultBucketName
	}
	conf.BucketName = bucketName

	sameDir, ok := viper.Get("snapshot.leveldb.sameDir").(bool)
	if !ok {
		log.Printf("snapshot.Conf(): can't find [snapshot.sameDir] in the config file, using default value [true]\n")
		sameDir = true
	}
	conf.SameDir = sameDir

	return &conf
}
