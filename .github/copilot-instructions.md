# Goalfeed Development Instructions

Goalfeed is a Go application that provides real-time NHL and MLB goal updates for Home Assistant integration. The application fetches data from external sports APIs and sends goal events to Home Assistant.

**Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.**

## Working Effectively

### Bootstrap and Build
Execute these commands in sequence to set up the development environment:

```bash
# Verify Go version (requires 1.21+)
go version

# Download and verify dependencies (~1 second after first run)
go mod tidy

# Build the application (~1 second after dependencies ready)
go build -o goalfeed .
```

**NEVER CANCEL**: Initial dependency download may take longer on slow networks. Build commands typically complete in 1-15 seconds. Always set timeout to 60+ seconds when using build commands.

### Testing
Run tests using these specific commands:

```bash
# Run unit tests for core packages (~1 second total)
go test ./models ./targets/homeassistant ./targets/memoryStore ./services/leagues/nhl ./services/leagues/mlb ./services/leagues/iihf

# Run specific main package tests that work without network (~1 second)
go test -v -run="TestTeamIsMonitored|TestGameIsMonitored|TestSendTestGoal|TestFireGoalEvents" .

# DO NOT run "go test ./..." - it fails due to network dependencies and nil pointer panics
```

**NEVER CANCEL**: Test suites complete in under 1 second when cached. Set timeout to 30+ seconds for safety.

### Code Quality
Always run these before committing changes:

```bash
# Format code (may modify files)
go fmt ./...

# Check for common issues (reports struct tag error in nhl models - known issue)
go vet ./...
```

### Application Usage
```bash
# Show help and available flags
./goalfeed --help

# Example configurations (NOTE: app will panic in restricted networks)
./goalfeed --nhl "WPG" --mlb "TOR" --test-goals
GOALFEED_WATCH_NHL="WPG" ./goalfeed
```

## Critical Limitations

### Network Dependencies
**WARNING**: The application immediately attempts to connect to external APIs on startup:
- NHL: `api-web.nhle.com`
- MLB: `statsapi.mlb.com`

In restricted network environments, the application will panic with "server misbehaving" or nil pointer dereference errors. This is expected behavior.

### Docker Build
Docker builds fail in sandboxed environments due to certificate verification issues. The Dockerfile is valid but requires proper network access.

### Test Suite Limitations  
- Full test suite (`go test ./...`) fails due to network calls and panic in `utils/utils.go:33`
- Use the specific test commands listed above instead
- Main package tests that require network access will fail

## Configuration Options

### Command Line Flags
- `--nhl strings`: NHL teams to watch (e.g., "WPG,TOR")
- `--mlb strings`: MLB teams to watch (e.g., "TOR,BOS") 
- `--test-goals`: Enable test goal events every minute

### Configuration File
Create `config.yaml` in the working directory:
```yaml
home_assistant:
  url: "http://yourhomeassistanturl"
  access_token: "yourhomeassistantaccesstoken"
watch:
  nhl:
    - WPG
  mlb:
    - TOR
```

### Environment Variables
All config options support environment variables with `GOALFEED_` prefix:
- `GOALFEED_WATCH_NHL="WPG"`
- `GOALFEED_WATCH_MLB="TOR"`
- `GOALFEED_TEST_GOALS=true`

## Project Structure

### Key Directories
- `/models` - Data models (game.go, team.go, event.go, league.go)
- `/services/leagues` - League-specific services (nhl/, mlb/, iihf/)
- `/targets` - Output targets (homeassistant/, memoryStore/)
- `/clients/leagues` - API clients with mock implementations
- `/config` - Configuration management (config.go)
- `/utils` - Utility functions (utils.go, logger.go)

### Important Files
- `main.go` - Application entry point and core logic
- `main_test.go` - Main package tests with mock implementations
- `go.mod` - Go module definition (requires Go 1.21+)
- `Dockerfile` - Multi-stage Docker build
- `.goreleaser.yml` - Release configuration for multiple platforms

## Common Development Tasks

### Adding New Team Support
1. Modify the appropriate service in `/services/leagues/`
2. Update mock clients in `/clients/leagues/` for testing
3. Add test cases following existing patterns
4. Update team monitoring logic in `main.go` if needed

### Testing Changes
1. Always run the working test subset first
2. Use mock clients for testing (see existing test files)
3. Test configuration via environment variables or config files
4. Verify formatting with `go fmt ./...`

### Release Process
- CI automatically tests, builds, and releases via GoReleaser
- Supports multiple platforms: Linux, Windows, macOS (arm, arm64, amd64, 386)
- Release workflow triggered on pushes to main branch

## Validation Scenarios

**ALWAYS run these validation steps after making changes:**

1. **Build Validation**: 
   ```bash
   go mod tidy && go build -o goalfeed .
   ```

2. **Test Validation**:
   ```bash
   go test ./models ./targets/homeassistant ./targets/memoryStore ./services/leagues/nhl ./services/leagues/mlb ./services/leagues/iihf
   ```

3. **Code Quality**:
   ```bash
   go fmt ./... && go vet ./...
   ```

4. **Basic Functionality** (if network available):
   ```bash
   ./goalfeed --help
   # Test configuration parsing (will fail in restricted networks)
   ```

**Do not attempt to run full end-to-end testing without proper network access to external sports APIs.**

## Troubleshooting

### Common Issues
- **Panic on startup**: Expected in restricted networks due to immediate API calls
- **"go test ./..." fails**: Use specific package tests instead
- **Struct tag error in go vet**: Known issue in nhl models, does not affect functionality
- **Docker build fails**: Requires proper network access for Go module downloads

### Known Bugs
- `utils/utils.go:33`: GetByte function doesn't handle nil response properly, causing panics when network calls fail

### Working Around Network Issues
- Use unit tests instead of integration tests
- Focus on mock-based testing for development
- Validate configuration parsing separately from runtime execution