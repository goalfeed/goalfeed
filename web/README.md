# Goalfeed Web Interface

A modern React/TypeScript web interface for Goalfeed that provides real-time scoreboards and team management.

## Features

- **Real-time Scoreboard**: Live updates of active games across all supported leagues (NHL, MLB, CFL, EPL, IIHF)
- **Team Management**: Configure which teams to monitor for each league
- **Event Feed**: Real-time notifications when teams score
- **WebSocket Integration**: Live updates without page refresh
- **Responsive Design**: Works on desktop and mobile devices

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- npm or yarn

### Development Setup

1. **Build the Go backend**:
   ```bash
   go build -o goalfeed .
   ```

2. **Start the web server (single command)**:
   ```bash
   ./goalfeed --web --web-port 8080 --cfl "*" --nhl "*" --mlb "*"
   ```

3. **Access the web interface**:
   - Web Interface: http://localhost:8080
   - API: http://localhost:8080/api

### Using the Development Script

The `web-dev.sh` script provides convenient commands:

```bash
# Start complete web server (recommended)
./web-dev.sh web

# Start frontend only (for development)
./web-dev.sh frontend

# Build frontend for production
./web-dev.sh build

# Show help
./web-dev.sh help
```

## Production Deployment

**Single Command Deployment**:
```bash
./goalfeed --web --web-port 8080
```

The server will automatically:
1. Install frontend dependencies
2. Build the React frontend
3. Serve both the API and frontend
4. Start monitoring configured teams

**Access the application**:
- Web Interface: http://localhost:8080
- API: http://localhost:8080/api

## API Endpoints

- `GET /api/games` - Get all active games
- `GET /api/leagues` - Get league configurations
- `POST /api/leagues` - Update league team configurations
- `GET /api/events` - Get recent events
- `GET /ws` - WebSocket connection for real-time updates

## WebSocket Events

The WebSocket connection receives the following event types:

- `game_update` - Game state changes (scores, status)
- `event` - New goal/score events
- `games_list` - Initial list of active games

## Configuration

Team monitoring can be configured through:

1. **Command line flags**:
   ```bash
   ./goalfeed --web --nhl TOR --mlb TOR --cfl BC
   ```

2. **Web interface**: Use the "Manage Teams" tab to configure teams

3. **Configuration file**: Update `config.yaml`:
   ```yaml
   watch:
     nhl:
       - TOR
       - MTL
     mlb:
       - TOR
     cfl:
       - BC
       - OTT
   ```

## Architecture

### Frontend (React/TypeScript)
- **Components**: Scoreboard, TeamManager, EventFeed
- **Hooks**: useWebSocket for real-time updates
- **Styling**: Tailwind CSS for responsive design
- **State Management**: React hooks for local state

### Backend (Go)
- **HTTP API**: Gin framework with CORS support
- **WebSocket**: Gorilla WebSocket for real-time communication
- **Integration**: Broadcasts game updates and events to web clients

### Real-time Updates
- Game state changes are broadcast to all connected web clients
- Events (goals/scores) are pushed immediately to the frontend
- WebSocket maintains persistent connections for live updates

## Development Notes

- The web interface runs alongside the normal Goalfeed functionality
- All existing features (Home Assistant integration, etc.) continue to work
- Web mode is enabled with the `--web` flag
- The frontend is served statically from the Go backend in production
