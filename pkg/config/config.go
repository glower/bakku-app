package config

import (
	"encoding/json"
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

// GetConfigPath returns a path to the configs of the app: search first in the ENV variable, then in the user home
func GetConfigPath() string {
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
	path := GetConfigPath()
	log.Printf("config.ReadDefaultConfig(): read config file [%s] from [%s]\n", defaultConfigName, path)
	viper.SetConfigName(defaultConfigName) // name of config file (without extension)
	viper.AddConfigPath(path)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
}

type WatchConfig struct {
	DirsToWatch []Watch `json:"dirs_to_watch"`
}

// Watch ...
type Watch struct {
	Path   string `json:"path"`
	Active bool   `json:"active"`
}

// DirectoriesToWatch returns a list of directories to watch for the file changes
func DirectoriesToWatch() *WatchConfig {
	result := &WatchConfig{}

	err := viper.Unmarshal(&result)
	if err != nil {
		panic("Unable to unmarshal config")
	}

	return result
}

func (c *WatchConfig) ToJSON() string {
	fmt.Printf("WatchConfig.WatchConfig(): %v\n", c)
	jsonConf, err := json.Marshal(c)
	if err != nil {
		panic("Unable to unmarshal config")
	}
	fmt.Printf("WatchConfig.WatchConfig(): json=%s\n", jsonConf)
	return string(jsonConf)
}
