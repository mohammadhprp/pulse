# Pulse - Distributed Log Collection System

Pulse is a distributed log collection and storage system that uses Kafka for message queuing and ClickHouse for efficient storage and querying of log data.

## Architecture

The system consists of two main components:

1. **Agent**: Collects logs from standard input and sends them to a Kafka topic.
2. **Collector**: Consumes logs from Kafka and stores them in ClickHouse.

## Prerequisites

- Docker and Docker Compose
- Go 1.23+ (for local development)

## Quick Start

1. Copy the example environment file and update as needed:

   ```bash
   cp .env.example .env
   ```

2. Build and start the services:

   ```bash
   make build
   ```

## Components

### Agent

The agent component reads JSON-formatted log events from stdin, parses them into the Event model, and produces messages to Kafka. This makes it easy to pipe logs from any application into the Pulse system.

### Collector

The collector consumes log events from Kafka and stores them in ClickHouse for efficient querying and analysis.

### Storage

Logs are stored in ClickHouse with a TTL of 30 days. The schema includes:

- EventTimeMs (UInt64)
- Timestamp (DateTime, materialized from EventTimeMs)
- Service (String)
- Level (Enum: DEBUG, INFO, WARN, ERROR)
- Message (String)
- Host (String)
- RequestID (UUID)

The data is partitioned by day for optimal query performance.

## Event Format

Log events should be in JSON format with the following structure:

```json
{
  "event_time_ms": 1651234567890,
  "service": "my-service",
  "level": "INFO",
  "message": "User logged in",
  "host": "server-1",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Development

### Project Structure

```bash
.
├── cmd/
│   ├── agent/       # Agent application entry point
│   └── collector/   # Collector application entry point
├── internal/
│   ├── agent/       # Agent specific code
│   ├── collector/   # Collector specific code
│   └── storage/     # Storage layer (ClickHouse)
├── pkg/
│   └── models/      # Shared data models
└── scripts/
    ├── entrypoint.sh       # Container entrypoint script
    └── init-clickhouse.sql # ClickHouse initialization script
```

### Available Commands

The following Make commands are available for development:

- `make build` - Build and start containers in detached mode
- `make start` - Start existing containers
- `make stop` - Stop and remove containers
- `make restart` - Restart containers
- `make logs` - Tail container logs
- `make clean` - Stop containers and remove volumes, images
- `make help` - Show available commands

## Configuration

Configure the application using environment variables (see `.env.example`):

- `KAFKA_BROKER`: Kafka broker address (default: kafka:9092)
- `KAFKA_TOPIC`: Kafka topic for logs (default: logs)
- `CLICKHOUSE_ADDR`: ClickHouse server address (default: clickhouse:9000)
- `CLICKHOUSE_DB`: ClickHouse database name (default: gologcentral)
- `CLICKHOUSE_USER`: ClickHouse username (default: default)
- `CLICKHOUSE_PASS`: ClickHouse password
- `LOG_LEVEL`: Logging verbosity (options: debug, info, warn, error, default: info)

## Logging

Pulse uses structured JSON logging powered by Zap. This provides:

- High-performance logging with minimal allocations
- Structured JSON output for easy parsing by log aggregation tools
- Different log levels (debug, info, warn, error) configurable via environment variables
- Context-rich logs with consistent fields across components

You can control the verbosity of logging using the `LOG_LEVEL` environment variable.

## License

[MIT License](LICENSE)
