# Goalfeed

[![Build Status](https://github.com/goalfeed/goalfeed/workflows/PR%20Test/badge.svg)](https://github.com/goalfeed/goalfeed/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/goalfeed/goalfeed/branch/main/graph/badge.svg)](https://codecov.io/gh/goalfeed/goalfeed)
[![Coverage Check](https://github.com/goalfeed/goalfeed/workflows/Coverage%20Check/badge.svg)](https://github.com/goalfeed/goalfeed/actions/workflows/coverage-check.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/goalfeed/goalfeed)](https://goreportcard.com/report/github.com/goalfeed/goalfeed)
[![Release](https://img.shields.io/github/release/goalfeed/goalfeed.svg)](https://github.com/goalfeed/goalfeed/releases/latest)

Goalfeed is a service that provides real-time goal updates for NHL and MLB games. It's designed for Home Assistant but can be used as a standalone application.

## Quickstart with Home Assistant

Our wiki contains instructions for quickly getting started with Goalfeed and Home Assistant.
- [First check out the Hassio installation page](https://github.com/goalfeed/goalfeed/wiki/Hassio-Add%E2%80%90on-Installation)
- [Next, check out the automation tutorial](https://github.com/goalfeed/goalfeed/wiki/Goal-Automation)

## Configuration

### Test Coverage

Goalfeed maintains high test coverage across all packages. The project includes comprehensive unit tests and integration tests to ensure reliability and correctness.

**Current Coverage Status:**
- 🎯 **7 out of 12 packages at 100% coverage**
- 📈 **Overall coverage significantly improved across all packages**
- ✅ **All tests passing with comprehensive error scenario coverage**

| Package | Coverage | Status |
|---------|----------|---------|
| `clients/leagues/nhl` | 100% | ✅ |
| `clients/leagues/mlb` | 100% | ✅ |
| `clients/leagues/iihf` | 100% | ✅ |
| `config` | 100% | ✅ |
| `models` | 100% | ✅ |
| `services/leagues/nhl` | 100% | ✅ |
| `services/leagues/iihf` | 100% | ✅ |
| `services/leagues/mlb` | 88% | 📈 |
| `targets/homeassistant` | 84% | 📈 |
| `targets/memoryStore` | 97.5% | 📈 |
| `utils` | 87% | 📈 |
| `main` | ~89% | 📈 |

Coverage is automatically checked on every pull request to prevent regressions.

**Local Coverage Testing:**
```bash
# Generate and view coverage report locally
./scripts/coverage.sh

# View detailed HTML coverage report
go tool cover -html=combined_coverage.out
```

## Configuration

Goalfeed allows users to specify which NHL and MLB teams they want to watch. This can be done using command-line flags, a YAML configuration file, or environment variables.

### Using Command-Line Flags

You can specify the NHL and MLB teams you want to watch using flags:

```bash
goalfeed --watch.nhl <team1,team2,...> --watch.mlb <team1,team2,...>
```

For example:

```bash
goalfeed --watch.nhl "WPG" --watch.mlb "TOR"
```

### Using a YAML Configuration File

You can also use a YAML configuration file to specify the teams you want to watch. Here's an example of the structure:

```yaml
watch:
  nhl:
  - WPG
  mlb:
  - TOR
```

Save this configuration to a file, for example, `config.yaml`. Then, you can run Goalfeed with the configuration file:

```bash
goalfeed --config /path/to/config.yaml
```

Ensure that the path to the configuration file is correctly specified.

### Using Environment Variables

Goalfeed supports the use of environment variables for configuration. This is particularly useful for deployment scenarios where you might not want to use command-line flags or configuration files.

To specify teams using environment variables:

```bash
export GOALFEED_WATCH_NHL="WPG"
export GOALFEED_WATCH_MLB="TOR"
```

Then, simply run the `goalfeed` command without any flags:

```bash
goalfeed
```

Goalfeed will pick up the environment variables and use them for configuration.

### Configuring Home Assistant Integration

Goalfeed sends goal events to Home Assistant. To configure this integration, you'll need to provide the necessary details for Home Assistant, such as the endpoint and authentication details. This configuration might be in another part of the codebase or might require a separate configuration file.
