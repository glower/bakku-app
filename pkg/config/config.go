package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config is an interface for a configuration provider
type Config interface {
	ReadDefaultConfig()
	DirectoriesToWatch() []string
}

const defaultConfigName = "config"
const defaultCofigPath = "."

// ReadDefaultConfig reads default config from the default path
func ReadDefaultConfig() {
	log.Printf("config.ReadDefaultConfig(): read config file [%s] from [%s]\n", defaultConfigName, defaultCofigPath)
	viper.SetConfigName(defaultConfigName) // name of config file (without extension)
	viper.AddConfigPath(defaultCofigPath)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
}

// DirectoriesToWatch returns a list of directories to watch for the file changes
func DirectoriesToWatch() []string {
	var result []string
	dirs, ok := viper.Get("watch").([]interface{})
	if !ok {
		log.Println("config.DirectoriesToWatch(): nothing to watch")
		return result
	}
	for _, dir := range dirs {
		path, ok := dir.(string)
		if !ok {
			log.Println("SetupWatchers(): invalid path")
		}
		result = append(result, path)
	}
	return result
}
