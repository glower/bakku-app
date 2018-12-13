package storage

import (
	"fmt"

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

	gConfig.Config = *conf

	clientIDTpl := fmt.Sprintf("storage.%s.clientID", configName)
	clientID, ok := viper.Get(clientIDTpl).(string)
	if !ok {
		// log.Printf("GoogleDriveConfig(): can't find [%s] in the config\n", clientIDTpl)
		clientID = ""
	}
	gConfig.ClientID = clientID

	clientSecretTpl := fmt.Sprintf("storage.%s.clientSecret", configName)
	clientSecret, ok := viper.Get(clientSecretTpl).(string)
	if !ok {
		// log.Printf("GoogleDriveConfig(): can't find [%s] in the config\n", clientSecretTpl)
		clientSecret = ""
	}
	gConfig.ClientSecret = clientSecret
	return gConfig
}
