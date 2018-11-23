package storage

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

const configName = "googleDrive"

// GDriveConfig is a struct for Google drive storage configuration
type GDriveConfig struct {
	Config
	ClientID     string
	ClientSecret string
}

// GoogleDriveConfig ...
func GoogleDriveConfig() *GDriveConfig {
	gConfig := &GDriveConfig{}
	conf := ProviderConf(configName)
	if !conf.Active {
		gConfig.Config = *conf
		return gConfig

	}

	clientIDTpl := fmt.Sprintf("storage.%s.clientID", configName)
	clientID, ok := viper.Get(clientIDTpl).(string)
	if !ok {
		log.Printf("GoogleDriveConfig(): can't find [%s]\n", clientID)
		clientID = ""
	}
	gConfig.ClientID = clientIDTpl

	clientSecretTpl := fmt.Sprintf("storage.%s.clientID", configName)
	clientSecret, ok := viper.Get(clientSecretTpl).(string)
	if !ok {
		log.Printf("GoogleDriveConfig(): can't find [%s]\n", clientSecret)
		clientSecret = ""
	}
	gConfig.ClientSecret = clientSecret

	return gConfig
}
