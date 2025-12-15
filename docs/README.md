# Goalfeed API Documentation

This directory contains automatically generated API documentation for the Goalfeed REST API and WebSocket API.

## REST API Documentation (OpenAPI/Swagger)

The REST API documentation is generated using [swaggo/swag](https://github.com/swaggo/swag) and follows the OpenAPI 2.0 specification.

### Viewing the Documentation

Once the server is running, you can view the interactive Swagger UI at:

- **Swagger UI**: http://localhost:8080/swagger/index.html

### Generated Files

- `swagger.json` - OpenAPI 2.0 specification in JSON format
- `swagger.yaml` - OpenAPI 2.0 specification in YAML format
- `docs.go` - Go package containing embedded documentation

### Regenerating Documentation

To regenerate the documentation after making changes to API endpoints:

```bash
# Install swag if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
export PATH=$PATH:$(go env GOPATH)/bin
swag init -g web/api/server.go -o docs --parseDependency --parseInternal
```

### Automatic Updates

The documentation is automatically updated via GitHub Actions when:
- API endpoints are modified in `web/api/server.go`
- Model definitions change in `models/`
- Documentation files are updated

See `.github/workflows/docs.yml` for the automation configuration.

## WebSocket API Documentation

The WebSocket API documentation is maintained manually in `WEBSOCKET.md`. This is separate from the OpenAPI documentation because WebSocket APIs cannot be fully described in OpenAPI 2.0.

See `WEBSOCKET.md` for:
- Connection details
- Message formats
- Event types
- Example client implementations

## API Endpoints

### Games
- `GET /api/games` - Get all active games
- `GET /api/upcoming` - Get upcoming games
- `POST /api/refresh` - Refresh active games
- `POST /api/clear` - Clear all games

### Leagues
- `GET /api/leagues` - Get league configurations
- `POST /api/leagues` - Update league configuration

### Events
- `GET /api/events` - Get filtered events

### Teams
- `GET /api/teams` - Get teams for a league

### Logs
- `GET /api/logs` - Get application logs

### Home Assistant
- `GET /api/homeassistant/status` - Get connection status
- `GET /api/homeassistant/config` - Get configuration
- `POST /api/homeassistant/config` - Update configuration

### WebSocket
- `WS /ws` - Real-time updates

## Contributing

When adding new API endpoints:

1. Add Swagger annotations to the handler function
2. Regenerate documentation: `swag init -g web/api/server.go -o docs`
3. Test the Swagger UI to ensure the endpoint appears correctly
4. Commit both the code changes and the generated documentation

For WebSocket changes, update `WEBSOCKET.md` manually.

