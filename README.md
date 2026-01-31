# OpenAI API Gateway

A high-performance, multi-channel OpenAI API gateway with intelligent routing, session stickiness, and comprehensive monitoring.

## Features

- **OpenAI API Compatible**: Full compatibility with OpenAI API (Chat Completions, Models)
- **Multi-Channel Support**: Configure multiple backend channels with different weights
- **Intelligent Routing**: Multi-factor scoring (weight, latency, error rate) for optimal channel selection
- **Session Stickiness**: Same user always routes to same channel for cache hit optimization
- **Health Checking**: Active and passive health monitoring of all channels
- **Prometheus Metrics**: Comprehensive metrics export for monitoring
- **Web Admin Interface**: Built-in web UI for channel and user management
- **Hot Reload**: Configuration changes without restart

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/X0Ken/openai-gateway.git
cd openai-gateway

# Install dependencies
go mod tidy

# Build
go build -o gateway main.go
```

### Configuration

Create a `config.yaml` file:

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30
  write_timeout: 30

database:
  path: "./gateway.db"

health_check:
  interval: 30
  timeout: 5

session:
  idle_timeout: 30

metrics:
  enabled: true
  port: 9090
```

### Run

```bash
./gateway
```

The server will start on port 8080.

## API Usage

### OpenAI Compatible Endpoints

#### Chat Completions

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

#### List Models

```bash
curl http://localhost:8080/v1/models
```

### Admin API

#### Create Channel

```bash
curl -X POST http://localhost:8080/api/channels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "openai-channel",
    "base_url": "https://api.openai.com",
    "api_key": "sk-your-key",
    "weight": 10,
    "enabled": true,
    "models": ["gpt-3.5-turbo", "gpt-4"]
  }'
```

#### Create User

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "user-api-key",
    "name": "Test User"
  }'
```

### Model-Channel Associations

The gateway supports associating multiple channels with a single model, enabling intelligent load balancing and failover.

#### Create Model

```bash
curl -X POST http://localhost:8080/api/models \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gpt-4"
  }'
```

#### Create Model-Channel Mapping

Associate a channel with a model and specify the backend model name:

```bash
curl -X POST http://localhost:8080/api/models/1/channels \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": 1,
    "backend_model_name": "gpt-4",
    "weight": 10
  }'
```

**Parameters:**
- `channel_id`: ID of the channel to associate
- `backend_model_name`: The model name to use on the backend (e.g., "gpt-4", "gpt-3.5-turbo")
- `weight`: Routing weight for this channel (default: 10)

#### List Channels for a Model

```bash
curl http://localhost:8080/api/models/1/channels
```

Response:
```json
[
  {
    "id": 1,
    "model_id": 1,
    "channel_id": 1,
    "backend_model_name": "gpt-4",
    "weight": 10,
    "created_at": "2026-01-31T10:00:00Z"
  }
]
```

#### Remove Model-Channel Mapping

```bash
curl -X DELETE http://localhost:8080/api/models/1/channels/1
```

**Architecture Note:**
- One model can be associated with multiple channels
- Channels are selected based on weight, latency, and error rate
- When a model is deleted, all associated channel mappings are automatically removed
- When a channel is deleted, all associated model mappings are automatically removed

## Web Interface

Access the web admin interface at: http://localhost:8080/

## Monitoring

Prometheus metrics are available at: http://localhost:8080/metrics

Key metrics:
- `gateway_requests_total`: Total requests
- `gateway_request_duration_seconds`: Request latency
- `gateway_channel_latency_seconds`: Channel response time
- `gateway_channel_error_rate`: Channel error rate

## Architecture

```
┌─────────────────┐
│   Client        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  API Gateway    │
│  (Gin Server)   │
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌────────┐ ┌────────┐
│ Router │ │ Metrics│
└────┬───┘ └────────┘
     │
     ▼
┌─────────────────┐
│ Channel Manager │
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌────────┐ ┌────────┐
│Backend1│ │Backend2│
└────────┘ └────────┘
```

## Development

### Run Tests

```bash
go test ./... -v
```

### Project Structure

```
.
├── cmd/server/         # Server entry point
├── internal/
│   ├── api/           # OpenAI API handlers
│   ├── admin/         # Admin API handlers
│   ├── auth/          # Authentication middleware
│   ├── channel/       # Channel management
│   ├── config/        # Configuration management
│   ├── metrics/       # Prometheus metrics
│   ├── model/         # Model management
│   ├── router/        # Smart routing engine
│   ├── session/       # Session management
│   └── web/           # Web UI
├── pkg/
│   ├── database/      # SQLite database layer
│   └── health/        # Health checking
└── config/            # Configuration files
```

## License

MIT
