package config

import (
	"github.com/spf13/viper"
)

func init() {
	// Set the file name of the configurations file
	viper.SetConfigName("config")

	// Set the path to look for the configurations file
	viper.AddConfigPath(".")

	// Enable environment variable overriding
	viper.AutomaticEnv()

	// Read in the configuration file
	viper.ReadInConfig()
}

// Get a configuration value as string
func GetString(key string) string {
	return viper.GetString(key)
}
