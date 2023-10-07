package config

import (
	"github.com/spf13/viper"
	"strings"
)

func init() {
	// Set the file name of the configurations file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml") // Type of the config file
	// Set the path to look for the configurations file
	viper.AddConfigPath(".")

	// Enable environment variable overriding
	viper.AutomaticEnv()

	// Read in the configuration file
	viper.SetEnvPrefix("GOALFEED")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.ReadInConfig()
}

// Get a configuration value as string
func GetString(key string) string {
	return viper.GetString(key)
}
func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}
