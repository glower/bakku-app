package snapshot

import (
	"log"

	"github.com/spf13/viper"
)

// Config is a struct for basic leveldb configuration
type Config struct {
	SameDir     bool
	SnapshotDir string
	Active      bool
}

const defaultSnapshotDir = ".snapshot"

// snapshot:
//   default: leveldb
//   leveldb:
//     sameDir: true
//     snapshotDir: .snapshot
//     active: true

// Conf ...
func Conf() *Config {
	var conf Config
	snapshotDir, ok := viper.Get("snapshot.leveldb.snapshotDir").(string)
	if !ok {
		log.Printf("SnapshotConf(): can't find [snapshot.leveldb.snapshotDir] in the config file, using default value [%s]\n", defaultSnapshotDir)
		snapshotDir = defaultSnapshotDir
	}
	conf.SnapshotDir = snapshotDir

	active, ok := viper.Get("snapshot.leveldb.active").(bool)
	if !ok {
		log.Printf("snapshot.Conf(): can't find [snapshot.leveldb.active] in the config file, using default value [true]\n")
		active = true
	}
	conf.Active = active

	sameDir, ok := viper.Get("snapshot.leveldb.sameDir").(bool)
	if !ok {
		log.Printf("snapshot.Conf(): can't find [snapshot.leveldb.sameDir] in the config file, using default value [true]\n")
		active = true
	}
	conf.SameDir = sameDir

	return &conf
}
