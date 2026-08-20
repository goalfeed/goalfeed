package main

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// newIsolatedLeagueFlagViper recreates, on a throwaway viper instance, the
// exact --nhl/--mlb/--cfl flag registration and watch.* key binding that
// main.go's init() performs on the global rootCmd/viper. Using an isolated
// instance (rather than mutating the package-level rootCmd/viper singletons)
// keeps these tests from leaking flag/env state into the many other tests in
// this package that read viper's watch.* keys.
func newIsolatedLeagueFlagViper() (*viper.Viper, *pflag.FlagSet) {
	v := viper.New()
	v.SetEnvPrefix("GOALFEED")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetConfigType("yaml")

	fs := pflag.NewFlagSet("goalfeed", pflag.ContinueOnError)
	fs.StringSlice("nhl", []string{}, "NHL teams to watch")
	fs.StringSlice("mlb", []string{}, "MLB teams to watch")
	fs.StringSlice("cfl", []string{}, "CFL teams to watch")
	_ = v.BindPFlag("watch.nhl", fs.Lookup("nhl"))
	_ = v.BindPFlag("watch.mlb", fs.Lookup("mlb"))
	_ = v.BindPFlag("watch.cfl", fs.Lookup("cfl"))
	return v, fs
}

// TestRootCmd_LeagueCLIFlags_OnlyNHLMLBCFL pins down which leagues get a CLI
// flag at all: the README documents --nhl/--mlb/--cfl as real flags, and NFL
// plus both Olympic hockey leagues as config.yaml/env-only (no CLI flag
// exists for them yet).
func TestRootCmd_LeagueCLIFlags_OnlyNHLMLBCFL(t *testing.T) {
	for _, name := range []string{"nhl", "mlb", "cfl"} {
		assert.NotNilf(t, rootCmd.PersistentFlags().Lookup(name), "--%s flag should be registered", name)
	}
	for _, name := range []string{"nfl", "olympic-men", "olympic_men", "olympic-women", "olympic_women"} {
		assert.Nilf(t, rootCmd.PersistentFlags().Lookup(name), "--%s flag should not exist; per the README, NFL and Olympic hockey teams are config.yaml/env-var only", name)
	}
}

// TestCLIFlags_NHLMLBCFLBindToWatchKeys pins down that --nhl/--mlb/--cfl
// parse as comma-separated team-code lists onto the watch.nhl/watch.mlb/
// watch.cfl viper keys the rest of the app reads (teamIsMonitoredByLeague).
func TestCLIFlags_NHLMLBCFLBindToWatchKeys(t *testing.T) {
	v, fs := newIsolatedLeagueFlagViper()

	assert.NoError(t, fs.Parse([]string{"--nhl=WPG,TOR", "--mlb=TOR", "--cfl=BC,OTT"}))

	assert.Equal(t, []string{"WPG", "TOR"}, v.GetStringSlice("watch.nhl"))
	assert.Equal(t, []string{"TOR"}, v.GetStringSlice("watch.mlb"))
	assert.Equal(t, []string{"BC", "OTT"}, v.GetStringSlice("watch.cfl"))
}

// TestConfigPrecedence_CLIFlagBeatsEnvBeatsConfigFile pins down the README's
// documented configuration precedence: "a config.yaml file in the working
// directory, GOALFEED_* environment variables, and CLI flags" in increasing
// priority. With nothing set beyond a config.yaml value, that value wins;
// adding an env var overrides it; explicitly setting the CLI flag overrides
// both.
func TestConfigPrecedence_CLIFlagBeatsEnvBeatsConfigFile(t *testing.T) {
	v, fs := newIsolatedLeagueFlagViper()

	// Lowest priority: a value from config.yaml.
	assert.NoError(t, v.ReadConfig(strings.NewReader("watch:\n  nhl:\n    - YAMLTEAM\n")))
	assert.Equal(t, []string{"YAMLTEAM"}, v.GetStringSlice("watch.nhl"), "config.yaml value should apply when nothing else is set")

	// Middle priority: GOALFEED_WATCH_NHL environment variable beats config.yaml.
	t.Setenv("GOALFEED_WATCH_NHL", "ENVTEAM")
	assert.Equal(t, []string{"ENVTEAM"}, v.GetStringSlice("watch.nhl"), "env var must beat config.yaml when no CLI flag is set")

	// Highest priority: explicitly setting --nhl beats both the env var and config.yaml.
	assert.NoError(t, fs.Parse([]string{"--nhl=FLAGTEAM"}))
	assert.Equal(t, []string{"FLAGTEAM"}, v.GetStringSlice("watch.nhl"), "CLI flag must beat env var and config.yaml once explicitly set")
}
