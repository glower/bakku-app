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
		log.Printf("SnapshotConf(): can't find [%s]\n", snapshotDir)
		snapshotDir = defaultSnapshotDir
	}
	conf.SnapshotDir = snapshotDir

	active, ok := viper.Get("snapshot.leveldb.active").(bool)
	if !ok {
		log.Printf("snapshot.Conf(): can't find [%v]\n", active)
		active = true
	}
	conf.Active = active

	sameDir, ok := viper.Get("snapshot.leveldb.sameDir").(bool)
	if !ok {
		log.Printf("snapshot.Conf(): can't find [%v]\n", sameDir)
		active = true
	}
	conf.SameDir = sameDir

	return &conf
}
