package storage

import "github.com/spf13/viper"

// LDriveConfig is a struct for local drive storage configuration
type LDriveConfig struct {
	Config
	AddLatency bool
}

// LocalDriveConfig ...
func LocalDriveConfig() *LDriveConfig {
	localConf := &LDriveConfig{}
	conf := ProviderConf("local")

	localConf.Config = *conf

	addLatencyTpl := "storage.local.addLatency"
	addLatency, ok := viper.Get(addLatencyTpl).(bool)
	if !ok {
		addLatency = false
	}
	localConf.AddLatency = addLatency

	return localConf
}
