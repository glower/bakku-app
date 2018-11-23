package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	home "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Config is an interface for a configuration provider
type Config interface {
	ReadDefaultConfig()
	DirectoriesToWatch() []string
}

const defaultConfigName = "config"
const defaultCofigPath = ".bakkuapp"

// search first in the ENV variable, then in the user home
func getConfigPath() string {
	configPath := os.Getenv("BAKKUAPPCONF")
	if configPath != "" {
		return configPath
	}

	homeDir, err := home.Dir()
	if err != nil {
		log.Printf("getConfigPath(): cannot read home dir: %v\n", err)
		return defaultCofigPath
	}

	configPath = filepath.Join(homeDir, defaultCofigPath)
	return configPath
}

// ReadDefaultConfig reads default config from the default path
func ReadDefaultConfig() {
	path := getConfigPath()
	log.Printf("config.ReadDefaultConfig(): read config file [%s] from [%s]\n", defaultConfigName, path)
	viper.SetConfigName(defaultConfigName) // name of config file (without extension)
	viper.AddConfigPath(path)
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
