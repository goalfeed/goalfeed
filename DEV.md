# Goalfeed Development Guide

## Quick Start

### Development Environment
```bash
# Start full development environment with auto-rebuild
./dev.sh dev

# Or using npm scripts
npm run dev
```

### Manual Commands
```bash
# Rebuild everything
./dev.sh rebuild

# Rebuild only frontend
./dev.sh frontend

# Rebuild only backend  
./dev.sh backend

# Restart server
./dev.sh restart

# Stop server
./dev.sh stop
```

## Auto-Compilation Rules

The development environment automatically rebuilds when files change:

### Frontend Changes
- **Triggers**: Any file in `web/frontend/src/`
- **Action**: Runs `npm run build` in `web/frontend/`
- **Result**: Updates the React build in `web/frontend/build/`

### Backend Changes  
- **Triggers**: Any `.go` file or files in `web/api/`
- **Action**: Runs `go build -o goalfeed .`
- **Result**: Creates new Goalfeed binary

### Server Restart
- **Triggers**: After successful builds
- **Action**: Stops old server and starts new one
- **Result**: Serves updated code at http://localhost:8080

## File Watching

The development script uses `fswatch` for file monitoring:

```bash
# Install fswatch (macOS)
brew install fswatch

# Install fswatch (Linux)
sudo apt-get install fswatch
```

If `fswatch` is not available, the script falls back to manual rebuild mode.

## Development Workflow

1. **Start Development Environment**
   ```bash
   ./dev.sh dev
   ```

2. **Make Changes**
   - Edit React components in `web/frontend/src/`
   - Edit Go backend code in `web/api/` or other `.go` files
   - Changes are automatically detected and rebuilt

3. **Test Changes**
   - Web interface: http://localhost:8080
   - API: http://localhost:8080/api/games
   - WebSocket: ws://localhost:8080/ws

4. **Stop Development**
   ```bash
   ./dev.sh stop
   ```

## Project Structure

```
goalfeed/
├── .cursorrules          # Cursor IDE rules for auto-compilation
├── dev.sh               # Development helper script
├── package.json         # Root package.json with dev scripts
├── web/
│   ├── api/             # Go backend API server
│   └── frontend/        # React frontend
│       ├── src/         # React source code
│       └── build/      # Built React app (auto-generated)
├── services/            # League service implementations
├── models/             # Data models
└── clients/            # API client implementations
```

## Troubleshooting

### Frontend Build Fails
```bash
# Check for TypeScript errors
cd web/frontend && npm run build

# Clear build cache
cd web/frontend && rm -rf build && npm run build
```

### Backend Build Fails
```bash
# Check Go compilation errors
go build -o goalfeed .

# Run tests
go test ./...
```

### Server Won't Start
```bash
# Check if port is in use
lsof -i :8080

# Kill existing processes
pkill -f goalfeed
```

## Cursor IDE Integration

The `.cursorrules` file provides automatic compilation rules for Cursor IDE:

- **Frontend Changes**: Auto-runs `npm run build` in `web/frontend/`
- **Backend Changes**: Auto-runs `go build` and restarts server
- **API Testing**: Provides curl commands for testing endpoints
- **Development Workflow**: Guides proper file organization

## Environment Variables

```bash
# Optional: Set custom port
export GOALFEED_PORT=8080

# Optional: Enable debug logging
export GOALFEED_DEBUG=true
```

## API Endpoints

- `GET /api/games` - Get all active games
- `GET /api/leagues` - Get available leagues  
- `GET /api/events` - Get recent events
- `GET /ws` - WebSocket connection for real-time updates

