# Goalfeed

Goalfeed is a service that provides real-time goal updates for NHL and MLB games. It's designed to be used with Home Assistant.

## Configuration

Goalfeed allows users to specify which NHL and MLB teams they want to watch. This can be done using command-line flags, a YAML configuration file, or environment variables.

### Using Command-Line Flags

You can specify the NHL and MLB teams you want to watch using flags:

\```bash
goalfeed --watch.nhl <team1,team2,...> --watch.mlb <team1,team2,...>
\```

For example:

\```bash
goalfeed --watch.nhl "WPG" --watch.mlb "TOR"
\```

### Using a YAML Configuration File

You can also use a YAML configuration file to specify the teams you want to watch. Here's an example of the structure:

\```yaml
watch:
nhl:
- WPG
mlb:
- TOR
\```

Save this configuration to a file, for example, `config.yaml`. Then, you can run Goalfeed with the configuration file:

\```bash
goalfeed --config /path/to/config.yaml
\```

Ensure that the path to the configuration file is correctly specified.

### Using Environment Variables

Goalfeed supports the use of environment variables for configuration. This is particularly useful for deployment scenarios where you might not want to use command-line flags or configuration files.

To specify teams using environment variables:

\```bash
export GOALFEED_WATCH_NHL="WPG"
export GOALFEED_WATCH_MLB="TOR"
\```

Then, simply run the `goalfeed` command without any flags:

\```bash
goalfeed
\```

Goalfeed will pick up the environment variables and use them for configuration.

### Configuring Home Assistant Integration

Goalfeed sends goal events to Home Assistant. To configure this integration, you'll need to provide the necessary details for Home Assistant, such as the endpoint and authentication details. This configuration might be in another part of the codebase or might require a separate configuration file.
