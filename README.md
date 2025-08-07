# tail-on

**Easily manage and monitor applications on your machine via [Tailscale](https://tailscale.com/).**

This project provides a web-based application manager that runs on your [Tailscale](https://tailscale.com/) network, allowing you to start, stop, and monitor configured applications from anywhere on your tailnet. It features real-time log streaming, a modern web interface, and RESTful API access.

## Installation

```bash
go install github.com/sierrasoftworks/tail-on@latest
```

## Usage

At its simplest, you can start managing applications by creating a configuration file and running the `tail-on` command:

```bash
# Start tail-on with the default configuration file
tail-on

# Start with a custom configuration file
tail-on --config /path/to/config.yaml

# Enable verbose logging
tail-on --verbose
```

### Using with Tailscale

To expose tail-on on your Tailscale network, you'll need to configure the Tailscale integration in your configuration file. Here's how to set it up:

#### Basic Tailscale Configuration

```yaml
tailscale:
  name: "my-tail-on-server"  # The hostname that will appear on your tailnet
  state_dir: "/var/lib/tailscale"  # Directory to store Tailscale state

applications:
  - name: "my-app"
    path: "/path/to/executable"
```

#### With Authentication Key (Headless Setup)

For automated deployments or headless servers, you can authenticate using an authkey:

```bash
# Set your Tailscale authkey as an environment variable
export TS_AUTHKEY="tskey-auth-your-key-here"

# Start tail-on (it will automatically authenticate on first run)
tail-on
```

#### Running in Ephemeral Mode

For temporary deployments or testing, you can run in ephemeral mode where the service is removed from your tailnet when stopped:

```yaml
tailscale:
  ephemeral: true
```

#### Accessing Your Service

Once configured and running, your tail-on service will be available at:

- `https://my-tail-on-server.your-tailnet.ts.net` (if HTTPS certificates are available)
- `http://my-tail-on-server.your-tailnet.ts.net` (fallback to HTTP)

**Note**: Replace `my-tail-on-server` with the name you specified in your configuration, and `your-tailnet` with your actual Tailscale tailnet name.

## Configuration

The service is configured via a YAML file (default: `config.yaml`) that defines the applications you want to manage:

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

### Managing Applications

Once configured, you can manage your applications through the web interface or API:

- **Web Interface**: Navigate to your tail-on service in a browser to access the modern web UI
- **RESTful API**: Use HTTP requests to programmatically manage applications
- **Real-time Logs**: Stream application output in real-time through the web interface or API

### Tailscale Integration Details

The service integrates deeply with Tailscale to provide seamless network access:

- **Automatic TLS**: Attempts to obtain HTTPS certificates when available
- **MagicDNS**: Uses your Tailscale hostname for easy discovery
- **Network Security**: Inherits Tailscale's zero-trust network model
- **State Persistence**: Saves Tailscale configuration for reliable restarts

#### Access Control with Tailscale ACLs

You can use Tailscale's Access Control Lists (ACLs) to restrict who can access your tail-on service. This provides fine-grained control over which users or devices can manage your applications:

```json
{
  "acls": [
    {
      "action": "accept",
      "src": ["group:admins"],
      "dst": ["my-tail-on-server:80", "my-tail-on-server:443"]
    },
    {
      "action": "accept", 
      "src": ["user:alice@example.com"],
      "dst": ["my-tail-on-server:*"]
    }
  ],
  "groups": {
    "group:admins": ["user:admin@example.com", "user:devops@example.com"]
  }
}
```

This example configuration:

- Allows members of the `admins` group to access the service on ports 80 and 443
- Grants `alice@example.com` full access to all ports on the service
- Blocks all other users from accessing the tail-on service

For more information on configuring Tailscale ACLs, see the [Tailscale ACL documentation](https://tailscale.com/kb/1018/acls/).

### Log Management

Applications are monitored continuously with comprehensive logging:

- Circular buffer of up to 1,000 log lines per application
- Combined stdout and stderr with timestamps
- Real-time streaming via Server-Sent Events
- In-memory storage (logs are not persisted to disk)

## API Examples

The tail-on service provides a RESTful API for programmatic access to application management:

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

### CLI Options

- `--config, -c`: Path to configuration file (default: "config.yaml")
- `--verbose, -v`: Enable verbose logging

### Exposing TailOn on non-Tailscale Interfaces

If you wish to allow people to access your TailOn server without needing to go via Tailscale,
you can set the `listen` option in your configuration. We strongly recommend binding this to
`localhost:*` or `127.0.0.1:*` to avoid the risk of bad actors with network access to your device
being able to manage your application remotely.

```yaml
listen: "localhost:8080"  # Optional: also bind to local interface
```
