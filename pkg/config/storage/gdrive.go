package storage

import (
	"fmt"

	"github.com/spf13/viper"
)

const configName = "googleDrive"

// GDriveConfig is a struct for Google drive storage configuration
type GDriveConfig struct {
	Config
	TokenFile       string
	CredentialsFile string
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

	tokenFileTpl := fmt.Sprintf("storage.%s.tokenFile", configName)
	tokenFile, ok := viper.Get(tokenFileTpl).(string)
	if !ok {
		// log.Printf("GoogleDriveConfig(): can't find [%s] in the config\n", tokenFileTpl)
		tokenFile = ""
	}
	gConfig.TokenFile = tokenFile

	credentialsFileTpl := fmt.Sprintf("storage.%s.credentialsFile", configName)
	credentialsFile, ok := viper.Get(credentialsFileTpl).(string)
	if !ok {
		// log.Printf("GoogleDriveConfig(): can't find [%s] in the config\n", credentialsFile)
		credentialsFile = ""
	}
	gConfig.CredentialsFile = credentialsFile
	return gConfig
}
