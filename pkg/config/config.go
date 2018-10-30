package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

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
