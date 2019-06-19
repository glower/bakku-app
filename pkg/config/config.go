package config

import (
	"encoding/json"
	"fmt"
	"io"
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
func DirectoriesToWatch() (*WatchConfig, error) {
	result := &WatchConfig{}

	err := viper.Unmarshal(&result)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %v", err)
	}

	return result, nil
}

func (c *WatchConfig) ToJSON() (string, error) {
	jsonConf, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("unable to marshal config: %v", err)
	}
	return string(jsonConf), nil
}

func FromJSON(input io.ReadCloser) (*WatchConfig, error) {
	conf := &WatchConfig{}
	decoder := json.NewDecoder(input)
	err := decoder.Decode(&conf)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %v", err)
	}
	return conf, nil
}
