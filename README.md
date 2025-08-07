# tailon

A web service for managing applications over Tailscale, built with Go, Gorilla Mux, logrus, and Cobra.

## Features

- **Application Management**: Start, stop, and monitor configured applications
- **Tailscale Integration**: Exposed directly over your Tailscale network
- **Real-time Logs**: Stream application logs via Server-Sent Events
- **RESTful API**: JSON-based API for all operations
- **Circular Log Buffer**: In-memory log storage with automatic rotation

## API Endpoints

- `GET /api/v1/apps` - List all configured applications
- `GET /api/v1/apps/{app_name}` - Get details of a specific application
- `POST /api/v1/apps/{app_name}/start` - Start an application
- `POST /api/v1/apps/{app_name}/stop` - Stop an application
- `POST /api/v1/apps/{app_name}/restart` - Restart an application (stop then start)
- `GET /api/v1/apps/{app_name}/logs` - Get application logs (JSON format)
- `GET /api/v1/apps/{app_name}/logs` (with `Accept: text/event-stream`) - Stream logs via SSE

## Configuration

The service is configured via a YAML file (default: `config.yaml`):

```yaml
applications:
  - name: "my-app"
    path: "/path/to/executable"
    args: ["--flag", "value"]
    env:
      - "ENV_VAR=value"

listen: "localhost:8080"  # Optional: local interface binding

tailscale:
  name: "my-tail-on-server"  # Hostname on Tailscale network
  state_dir: "/tmp/tailscale-state"  # Tailscale state directory
```

## Usage

### Building

```bash
go build -o tail-on
```

### Running

```bash
# With default config file (config.yaml)
./tail-on

# With custom config file
./tail-on --config /path/to/config.yaml

# With verbose logging
./tail-on --verbose
```

### CLI Options

- `--config, -c`: Path to configuration file (default: "config.yaml")
- `--verbose, -v`: Enable verbose logging

## API Examples

### List all applications
```bash
curl http://localhost:8080/api/v1/apps
```

### Start an application
```bash
curl -X POST http://localhost:8080/api/v1/apps/my-app/start
```

### Stop an application
```bash
curl -X POST http://localhost:8080/api/v1/apps/my-app/stop
```

### Restart an application
```bash
curl -X POST http://localhost:8080/api/v1/apps/my-app/restart
```

### Get application logs
```bash
curl http://localhost:8080/api/v1/apps/my-app/logs
```

### Stream logs via Server-Sent Events
```bash
curl -H "Accept: text/event-stream" http://localhost:8080/api/v1/apps/my-app/logs
```

## Tailscale Integration

The service automatically configures itself to listen on your Tailscale network:

1. It attempts to listen on port 443 first (for HTTPS if certificates are available)
2. Falls back to port 80 (HTTP) if HTTPS is not available
3. Uses the configured Tailscale hostname and state directory
4. Optionally also listens on a local interface if configured

## Log Management

- Each application maintains a circular buffer of up to 1,000 log lines
- Logs include both stdout and stderr with timestamps
- Real-time log streaming is available via Server-Sent Events
- Logs are stored in memory and not persisted to disk

## Dependencies

- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP routing
- [logrus](https://github.com/sirupsen/logrus) - Structured logging
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Tailscale](https://tailscale.com/) - Network connectivity
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML configuration parsing
