# tailon

**Easily manage and monitor applications on your machine via [Tailscale](https://tailscale.com/).**

This project provides a web-based application manager that runs on your [Tailscale](https://tailscale.com/) network, allowing you to start, stop, and monitor configured applications from anywhere on your tailnet. It features real-time log streaming, a modern web interface, and RESTful API access.

## Installation

```bash
go install github.com/sierrasoftworks/tailon@latest
```

## Usage

At its simplest, you can start managing applications by creating a configuration file and running the `tailon` command:

```bash
# Start tailon with the default configuration file
tailon

# Start with a custom configuration file
tailon --config /path/to/config.yaml

# Enable verbose logging
tailon --verbose
```

### Using with Tailscale

To expose tailon on your Tailscale network, you'll need to configure the Tailscale integration in your configuration file. Here's how to set it up:

#### Basic Tailscale Configuration

```yaml
tailscale:
  name: "my-tailon-server"  # The hostname that will appear on your tailnet
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

# Start tailon (it will automatically authenticate on first run)
tailon
```

#### Running in Ephemeral Mode

For temporary deployments or testing, you can run in ephemeral mode where the service is removed from your tailnet when stopped:

```yaml
tailscale:
  ephemeral: true
```

#### Accessing Your Service

Once configured and running, your tailon service will be available at:

- `https://my-tailon-server.your-tailnet.ts.net` (if HTTPS certificates are available)
- `http://my-tailon-server.your-tailnet.ts.net` (fallback to HTTP)

**Note**: Replace `my-tailon-server` with the name you specified in your configuration, and `your-tailnet` with your actual Tailscale tailnet name.

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
  name: "my-tailon-server"  # Hostname on Tailscale network
  state_dir: "/tmp/tailscale-state"  # Tailscale state directory
```

### Working Directory Configuration

Applications can specify a working directory where they will run. This is useful for applications that expect to run from a specific location or need access to files in a particular directory:

```yaml
applications:
  - name: "web-server"
    path: "/usr/bin/python3"
    args: ["-m", "http.server", "8000"]
    working_dir: "/var/www/html"  # Serve files from this directory
    env:
      - "PYTHONPATH=/app"
      
  - name: "file-processor"
    path: "/usr/bin/python3"
    args: ["process.py"]
    working_dir: "/data/input"    # Process files from this directory
    env:
      - "DATA_PATH=/data/input"
```

### Security Configuration

Tailon includes comprehensive security features to control access and protect sensitive information:

```yaml
# Security configuration
security:
  # Default role for anonymous users (when Tailscale is disabled)
  default_role: "admin"     # Options: admin, operator, viewer, or "" (none)

applications:
  - name: "secure-app"
    path: "/app/server"
    env:
      - "API_KEY=secret123"    # Will be hidden if hide_env_vars: true
      - "DATABASE_URL=postgres://..."
```

#### Role-Based Access Control

TailOn implements a flexible role-based authorization system with four permission levels:

- **`admin`**: Full access - can view, start, stop, and restart applications
- **`operator`**: Control access - can view, start, stop, and restart applications (environment variables may be hidden)
- **`viewer`**: Read-only access - can view application status and logs only
- **`none` or `""`**: No access - cannot access applications

#### Tailscale User Capabilities

When using Tailscale, you can grant users specific roles for applications using [Tailscale's capabilities feature](https://tailscale.com/kb/1537/grants-app-capabilities). Add the `sierrasoftworks/cap/tailon` capability to your Tailscale ACL policy:

```json
{
  "acls": [
    {
      "action": "accept",
      "src": ["group:admins"],
      "dst": ["my-tailon-server:80", "my-tailon-server:443"]
    }
  ],
  "groups": {
    "group:admins": ["user:admin@example.com"]
  },
  "grants": [
    {
      "src": ["user:alice@example.com"],
      "dst": ["my-tailon-server"],
      "app": {
        "sierrasoftworks/cap/tailon": [
          {
            "role": "operator",
            "applications": ["web-server", "api-service"]
          }
        ]
      }
    },
    {
      "src": ["group:admins"],
      "dst": ["my-tailon-server"],
      "app": {
        "sierrasoftworks/cap/tailon": [
          {
            "role": "admin",
            "applications": ["*"]
          }
        ]
      }
    }
  ]
}
```

In this example:

- `alice@example.com` gets `operator` role for `web-server` and `api-service` applications
- Members of `group:admins` get `admin` role for all applications (`*` wildcard)

Note that if Alice is a member of the `admins` group then the most specific rule will win and
she will **NOT** have `admin` on `web-server` or `api-service` (instead she will be limited to `operator`
access).

#### Security Recommendations

**For Production Environments:**

```yaml
security:
  default_role: ""          # No default access for anonymous users

tailscale:
  enabled: true             # Use Tailscale for secure access
```

**For Development/Internal Use:**

```yaml
listen: "localhost:8080"    # Bind to localhost only
security:
  default_role: "admin"     # Allow full access for development
```

### Managing Applications

Once configured, you can manage your applications through the web interface or API:

- **Web Interface**: Navigate to your tailon service in a browser to access the modern web UI
- **RESTful API**: Use HTTP requests to programmatically manage applications
- **Real-time Logs**: Stream application output in real-time through the web interface or API

### Tailscale Integration Details

The service integrates deeply with Tailscale to provide seamless network access:

- **Automatic TLS**: Attempts to obtain HTTPS certificates when available
- **MagicDNS**: Uses your Tailscale hostname for easy discovery
- **Network Security**: Inherits Tailscale's zero-trust network model
- **State Persistence**: Saves Tailscale configuration for reliable restarts

#### Access Control with Tailscale ACLs

You can use Tailscale's Access Control Lists (ACLs) to restrict who can access your tailon service. This provides fine-grained control over which users or devices can manage your applications:

```json
{
  "acls": [
    {
      "action": "accept",
      "src": ["group:admins"],
      "dst": ["my-tailon-server:80", "my-tailon-server:443"]
    },
    {
      "action": "accept", 
      "src": ["user:alice@example.com"],
      "dst": ["my-tailon-server:*"]
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
- Blocks all other users from accessing the tailon service

For more information on configuring Tailscale ACLs, see the [Tailscale ACL documentation](https://tailscale.com/kb/1018/acls/).

### Log Management

Applications are monitored continuously with comprehensive logging:

- Circular buffer of up to 1,000 log lines per application
- Combined stdout and stderr with timestamps
- Real-time streaming via Server-Sent Events
- In-memory storage (logs are not persisted to disk)

### Audit Logging

Tailon provides comprehensive audit logging for security and compliance:

- **User Tracking**: All actions are logged with user identification
- **IP Address Logging**: Anonymous users are tracked by IP address (`$anonymous-192.168.1.100$`)
- **Action Logging**: Start, stop, restart operations are recorded with timestamps
- **Enhanced Context**: Logs include user display names, IP addresses, and detailed action context

Example audit log entries:

```log
INFO  User started application  action=start target=web-server user_id=$anonymous-127.0.0.1$ ip_address=127.0.0.1
INFO  Alice Smith stopped application (Gracefully stopping application (SIGTERM)) action=stop target=web-server user_id=alice@company.com
```

## API Examples

The tailon service provides a RESTful API for programmatic access to application management:

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
listen: "localhost:8080"  # Bind to local interface

# Configure security settings for non-Tailscale access
security:
  allow_anonymous: true      # Allow access without Tailscale auth
```

**Security Warning**: When binding to non-localhost addresses (e.g., `0.0.0.0:8080`), anyone with network access to your machine can control your applications. Always use appropriate security configuration and consider using Tailscale for secure remote access instead.
