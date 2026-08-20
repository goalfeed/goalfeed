package config

import (
	"strings"

	"github.com/spf13/viper"
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
	// Defaults
	viper.SetDefault("nfl.fastcast.enabled", true)
	// NFL Fastcast keepalive/reconnect defaults
	viper.SetDefault("nfl.fastcast.ping_interval_sec", 20)
	viper.SetDefault("nfl.fastcast.pong_wait_sec", 60)
	viper.SetDefault("nfl.fastcast.reconnect_base_ms", 2000)
	viper.SetDefault("nfl.fastcast.reconnect_max_ms", 30000)
	// Security defaults: reject remote (non-private) Home Assistant URLs and
	// don't let the runtime API persist config changes to disk unless the
	// operator opts in.
	viper.SetDefault("home_assistant.allow_remote_url", false)
	viper.SetDefault("web.allow_config_writes", false)
	viper.ReadInConfig()
}

// Get a configuration value as string
func GetString(key string) string {
	return viper.GetString(key)
}
func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

// GetBool returns a configuration value as a bool.
func GetBool(key string) bool {
	return viper.GetBool(key)
}
