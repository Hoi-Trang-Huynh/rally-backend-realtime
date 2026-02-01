# Rally Realtime Server

Real-time WebSocket server for Rally application with Redis pub/sub support.

## Architecture

```
rally-realtime/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── internal/
│   ├── socket/               # WebSocket management
│   │   ├── client.go         # Single connected user
│   │   ├── hub.go            # Connection switchboard
│   │   └── router.go         # Message routing
│   ├── features/             # Feature handlers
│   │   ├── chat/             # Chat logic
│   │   ├── location/         # Geo updates
│   │   └── planning/         # Collaboration
│   ├── pubsub/               # Redis pub/sub
│   └── storage/              # Database logic
├── Dockerfile
└── docker-compose.yml
```

## Quick Start

### Using Docker Compose (Recommended)

```bash
docker-compose up -d
```

This starts:
- Redis on port 6379
- Rally Realtime Server on port 8080

### Local Development

1. Start Redis:
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

2. Run the server:
```bash
go run ./cmd/server
```

## WebSocket API

### Connection

```
ws://localhost:8080/ws?room_id=<room>&user_id=<user>
```

### Message Format

```json
{
  "type": "chat|location|planning",
  "room_id": "optional-room-override",
  "payload": { ... }
}
```

### Message Types

#### Chat
```json
{
  "type": "chat",
  "payload": {
    "content": "Hello world!"
  }
}
```

#### Location
```json
{
  "type": "location",
  "payload": {
    "latitude": 13.7563,
    "longitude": 100.5018,
    "accuracy": 10.0
  }
}
```

#### Planning
```json
{
  "type": "planning",
  "payload": {
    "action": "lock|unlock|update",
    "item_id": "itinerary-item-id",
    "data": { ... }
  }
}
```

## Health Check

```bash
curl http://localhost:8080/health
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| REDIS_ADDR | localhost:6379 | Redis address |

## Related Jira Issues

- [RAL-274](https://rally-app.atlassian.net/browse/RAL-274) - Set up RL-Socket-Redis server
