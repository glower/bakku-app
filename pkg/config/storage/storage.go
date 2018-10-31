package storage

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config is a struct for basic storage configuration
type Config struct {
	Name   string
	Path   string
	Active bool
}

// ProviderConf ...
func ProviderConf(name string) *Config {
	var conf Config
	storagePathTpl := fmt.Sprintf("storage.%s.path", name)
	path, ok := viper.Get(storagePathTpl).(string)
	if !ok {
		log.Printf("ProviderConf(): can't find [%s]\n", storagePathTpl)
		path = ""
	}
	conf.Path = path

	storageActiveTpl := fmt.Sprintf("storage.%s.active", name)
	active, ok := viper.Get(storageActiveTpl).(bool)
	if !ok {
		log.Printf("ProviderConf(): can't find [%s]\n", storageActiveTpl)
		active = false
	}
	conf.Active = active

	return &conf
}
