# WebSocket API Documentation

The Goalfeed WebSocket API provides real-time updates for game state changes, events, and logs.

## Connection

**Endpoint:** `ws://localhost:8080/ws` (or `wss://` for HTTPS)

**Protocol:** WebSocket (RFC 6455)

## Connection Flow

1. Client establishes WebSocket connection to `/ws`
2. Server immediately sends initial `games_list` message with all active games
3. Server broadcasts updates as they occur
4. Client can keep connection alive by reading messages (no ping/pong required)

## Message Format

All messages follow this JSON structure:

```json
{
  "type": "message_type",
  "data": { ... }
}
```

## Message Types

### `games_list`

Sent immediately upon connection and whenever the full games list changes.

**Data Structure:**
```json
{
  "type": "games_list",
  "data": [
    {
      "gameCode": "string",
      "leagueId": 1,
      "currentState": {
        "home": {
          "team": { "teamCode": "TOR", "teamName": "Toronto Maple Leafs" },
          "score": 2
        },
        "away": {
          "team": { "teamCode": "MTL", "teamName": "Montreal Canadiens" },
          "score": 1
        },
        "status": "active",
        "period": 2,
        "clock": "15:30"
      }
    }
  ]
}
```

### `game_update`

Sent whenever a game's state changes (score, period, status, etc.).

**Data Structure:**
```json
{
  "type": "game_update",
  "data": {
    "gameCode": "string",
    "leagueId": 1,
    "currentState": { ... }
  }
}
```

### `event`

Sent when a new game event occurs (goal, score, period change, etc.).

**Data Structure:**
```json
{
  "type": "event",
  "data": {
    "id": "event-id",
    "type": "goal",
    "timestamp": "2024-01-15T20:30:00Z",
    "description": "Goal scored by Player Name",
    "teamCode": "TOR",
    "teamName": "Toronto Maple Leafs",
    "leagueId": 1,
    "gameCode": "game-code",
    "period": 2,
    "time": "15:30",
    "score": {
      "homeScore": 2,
      "awayScore": 1,
      "homeTeam": "TOR",
      "awayTeam": "MTL"
    }
  }
}
```

### `log`

Sent for application log entries (debugging and monitoring).

**Data Structure:**
```json
{
  "type": "log",
  "data": {
    "timestamp": "2024-01-15T20:30:00Z",
    "level": "info",
    "type": "event",
    "message": "Game event processed",
    "leagueId": 1,
    "teamCode": "TOR"
  }
}
```

## Event Types

The `event.type` field can be one of the following:

- `goal` - Goal scored (NHL)
- `touchdown` - Touchdown scored (NFL/CFL)
- `home_run` - Home run (MLB)
- `penalty` - Penalty called
- `power_play` - Power play situation
- `period_start` - Period/quarter/inning started
- `period_end` - Period/quarter/inning ended
- `game_start` - Game started
- `game_end` - Game ended
- `shot` - Shot on goal
- `save` - Goalkeeper save
- `strikeout` - Strikeout (MLB)
- `walk` - Walk (MLB)
- `error` - Error occurred
- `turnover` - Turnover (NFL/CFL)
- `fumble` - Fumble (NFL/CFL)
- `interception` - Interception (NFL/CFL)
- `field_goal` - Field goal (NFL/CFL)
- `safety` - Safety (NFL/CFL)

## Example Client Implementation

### JavaScript/TypeScript

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('WebSocket connected');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  switch (message.type) {
    case 'games_list':
      console.log('Games list:', message.data);
      break;
    case 'game_update':
      console.log('Game updated:', message.data);
      break;
    case 'event':
      console.log('New event:', message.data);
      break;
    case 'log':
      console.log('Log entry:', message.data);
      break;
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket disconnected');
};
```

### Python

```python
import websocket
import json

def on_message(ws, message):
    data = json.loads(message)
    msg_type = data.get('type')
    
    if msg_type == 'games_list':
        print(f"Games list: {data['data']}")
    elif msg_type == 'game_update':
        print(f"Game updated: {data['data']}")
    elif msg_type == 'event':
        print(f"New event: {data['data']}")
    elif msg_type == 'log':
        print(f"Log entry: {data['data']}")

def on_error(ws, error):
    print(f"WebSocket error: {error}")

def on_close(ws, close_status_code, close_msg):
    print("WebSocket closed")

def on_open(ws):
    print("WebSocket connected")

ws = websocket.WebSocketApp("ws://localhost:8080/ws",
                            on_open=on_open,
                            on_message=on_message,
                            on_error=on_error,
                            on_close=on_close)

ws.run_forever()
```

## Error Handling

- If the connection is lost, the client should attempt to reconnect
- The server will automatically send the current games list upon reconnection
- No authentication is required for WebSocket connections
- All origins are allowed (CORS is handled at the HTTP level)

## Rate Limiting

The WebSocket server does not implement rate limiting. However, clients should:

- Avoid sending unnecessary messages (the server doesn't process client messages currently)
- Handle reconnection backoff to avoid overwhelming the server
- Implement exponential backoff for reconnection attempts

## Connection Lifecycle

1. **Connect**: Client opens WebSocket connection
2. **Initial Data**: Server sends `games_list` immediately
3. **Updates**: Server broadcasts updates as they occur
4. **Disconnect**: Client closes connection or connection times out

The server maintains the connection until:
- Client closes the connection
- Network error occurs
- Server shuts down

## Notes

- The WebSocket endpoint does not require authentication
- All messages are sent as JSON
- The connection is unidirectional (server → client) for now
- Client messages are currently ignored but the connection must remain open

