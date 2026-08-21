package config

import (
	"errors"
	"fmt"
	"os"
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
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			configMissing = true
		}
	}
}

// configMissing records whether startup found a config.yaml. Goalfeed can run
// without one (flags and GOALFEED_* env vars are enough), so this is not fatal
// — but a run with no config and no watched teams does nothing useful while
// still logging every scheduled game, which reads as a broken install rather
// than a missing file.
var configMissing bool

// MissingConfigNotice returns a one-line message to print at startup when no
// config.yaml was found AND no teams were configured by any other means, or ""
// when there is nothing to say. Callers print it once, before the log opens.
func MissingConfigNotice() string {
	if !configMissing {
		return ""
	}
	for _, k := range []string{
		"watch.nhl", "watch.mlb", "watch.cfl", "watch.nfl",
	} {
		if len(viper.GetStringSlice(k)) > 0 {
			return ""
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "the current directory"
	}
	return fmt.Sprintf(
		"No config.yaml found in %s, and no teams set via flags or GOALFEED_* env vars, "+
			"so Goalfeed has nothing to watch. Create a config.yaml next to the binary "+
			"(see config.example.yaml in the release archive), or pass --nhl WPG to try it "+
			"right now. Full reference: https://goalfeed.ca/docs/configuration/",
		cwd,
	)
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
