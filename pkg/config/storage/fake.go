package storage

import (
	"log"

	"github.com/spf13/viper"
)

type FakeConfig struct {
	Active bool
}

func FakeDriverConfig() *FakeConfig {
	c := &FakeConfig{}

	storageActiveTpl := "storage.fake.active"
	active, ok := viper.Get(storageActiveTpl).(bool)
	if !ok {
		log.Printf("ProviderConf(): can't find [%s]\n", storageActiveTpl)
		active = false
	}
	c.Active = active

	return c
}
