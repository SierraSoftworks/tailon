// Docs component for main documentation view
class Docs {
    constructor() {
        this.baseURL = window.location.origin;
        this.container = null;
    }

    // Initialize the docs view
    async init() {
        try {
            this.render();
        } catch (error) {
            console.error('Failed to initialize docs:', error);
            this.renderError('Failed to load documentation');
        }
    }

    // Render the docs view
    render() {
        const content = document.getElementById('content');
        content.innerHTML = '';

        // Page title
        const title = Utils.createElement('div', { className: 'page-title' }, [
            'TailOn Documentation'
        ]);

        // Create docs container
        this.container = Utils.createElement('div', { className: 'docs-container' });

        // Create header
        const header = this.createHeader();
        
        // Create main content sections
        const quickStart = this.createQuickStartSection();
        const configuration = this.createConfigurationSection();
        const security = this.createSecuritySection();
        const examples = this.createExamplesSection();
        const navigation = this.createNavigationSection();

        this.container.appendChild(header);
        this.container.appendChild(quickStart);
        this.container.appendChild(configuration);
        this.container.appendChild(security);
        this.container.appendChild(examples);
        this.container.appendChild(navigation);

        content.appendChild(title);
        content.appendChild(this.container);
    }

    // Create docs header
    createHeader() {
        const header = Utils.createElement('div', { className: 'docs-header' }, [
            Utils.createElement('h1', {}, ['TailOn Documentation']),
            Utils.createElement('p', { className: 'lead' }, [
                'TailOn is a web-based application manager that runs on your Tailscale network, allowing you to start, stop, and monitor configured applications from anywhere on your tailnet.'
            ])
        ]);

        return header;
    }

    // Create quick start section
    createQuickStartSection() {
        const section = Utils.createElement('div', { className: 'docs-section' });
        
        const title = Utils.createElement('h2', {}, ['Quick Start']);
        
        const steps = Utils.createElement('ol', { className: 'setup-steps' }, [
            Utils.createElement('li', {}, [
                Utils.createElement('strong', {}, ['Install TailOn:']),
                Utils.createElement('pre', {}, ['go install github.com/sierrasoftworks/tailon@latest'])
            ]),
            Utils.createElement('li', {}, [
                Utils.createElement('strong', {}, ['Create a configuration file:']),
                Utils.createElement('pre', {}, [`# config.yaml
applications:
  - name: "my-app"
    path: "/path/to/executable"
    args: ["--flag", "value"]
    
listen: "localhost:8080"

tailscale:
  name: "my-tailon-server"
  state_dir: "/tmp/tailscale-state"`])
            ]),
            Utils.createElement('li', {}, [
                Utils.createElement('strong', {}, ['Start TailOn:']),
                Utils.createElement('pre', {}, ['tailon --config config.yaml'])
            ])
        ]);

        section.appendChild(title);
        section.appendChild(steps);
        
        return section;
    }

    // Create configuration section
    createConfigurationSection() {
        const section = Utils.createElement('div', { className: 'docs-section' });
        
        const title = Utils.createElement('h2', {}, ['Configuration']);
        
        const intro = Utils.createElement('p', {}, [
            'TailOn is configured via a YAML file that defines applications, security settings, and network configuration.'
        ]);

        const basicConfig = this.createSubSection('Basic Application Configuration', `applications:
  - name: "web-server"
    path: "/usr/bin/python3"
    args: ["-m", "http.server", "8000"]
    working_dir: "/var/www/html"
    env:
      - "PYTHONPATH=/app"
      - "PORT=8000"`);

        const tailscaleConfig = this.createSubSection('Tailscale Integration', `tailscale:
  name: "my-tailon-server"      # Hostname on your tailnet
  state_dir: "/var/lib/tailscale"  # State directory
  ephemeral: false              # Remove from tailnet on shutdown`);

        const networkConfig = this.createSubSection('Network Configuration', `# Bind to local interface (optional)
listen: "localhost:8080"

# For development/testing only - avoid binding to 0.0.0.0
# listen: "0.0.0.0:8080"  # ⚠️ Security risk!`);

        section.appendChild(title);
        section.appendChild(intro);
        section.appendChild(basicConfig);
        section.appendChild(tailscaleConfig);
        section.appendChild(networkConfig);
        
        return section;
    }

    // Create security section
    createSecuritySection() {
        const section = Utils.createElement('div', { className: 'docs-section' });
        
        const title = Utils.createElement('h2', {}, ['Security Configuration']);
        
        const intro = Utils.createElement('p', {}, [
            'TailOn includes comprehensive security features to control access and protect sensitive information.'
        ]);

        const securityConfig = this.createSubSection('Security Settings', `security:
  # Allow anonymous users when Tailscale is disabled
  allow_anonymous: true
  
  # Restrict anonymous access to specific IP ranges
  allowed_ips:
    - "127.0.0.1"        # localhost only
    - "192.168.1.0/24"   # local network
  
  # Hide environment variables in API responses
  hide_env_vars: true`);

        const prodConfig = this.createSubSection('Production Recommendations', `# Secure production setup
security:
  allow_anonymous: false    # Require Tailscale authentication
  hide_env_vars: true      # Hide sensitive environment variables

tailscale:
  enabled: true            # Use Tailscale for secure access
  name: "prod-tailon-server"`);

        const auditInfo = Utils.createElement('div', { className: 'info-box' }, [
            Utils.createElement('h4', {}, ['Audit Logging']),
            Utils.createElement('p', {}, [
                'All user actions are logged with user identification, IP addresses, and timestamps. Anonymous users are tracked by IP address in the format ',
                Utils.createElement('code', {}, ['$anonymous-192.168.1.100$']),
                '.'
            ])
        ]);

        section.appendChild(title);
        section.appendChild(intro);
        section.appendChild(securityConfig);
        section.appendChild(prodConfig);
        section.appendChild(auditInfo);
        
        return section;
    }

    // Create examples section
    createExamplesSection() {
        const section = Utils.createElement('div', { className: 'docs-section' });
        
        const title = Utils.createElement('h2', {}, ['Common Examples']);

        const webServer = this.createSubSection('Web Server', `applications:
  - name: "static-site"
    path: "/usr/bin/python3"
    args: ["-m", "http.server", "8000"]
    working_dir: "/var/www/html"
    env:
      - "PYTHONPATH=/app"`);

        const nodeApp = this.createSubSection('Node.js Application', `applications:
  - name: "node-api"
    path: "/usr/bin/node"
    args: ["server.js"]
    working_dir: "/app"
    env:
      - "NODE_ENV=production"
      - "PORT=3000"`);

        const database = this.createSubSection('Database Service', `applications:
  - name: "postgres"
    path: "/usr/bin/postgres"
    args: ["-D", "/var/lib/postgresql/data"]
    working_dir: "/var/lib/postgresql"
    env:
      - "POSTGRES_DB=myapp"
      - "POSTGRES_USER=appuser"`);

        section.appendChild(title);
        section.appendChild(webServer);
        section.appendChild(nodeApp);
        section.appendChild(database);
        
        return section;
    }

    // Create navigation section
    createNavigationSection() {
        const section = Utils.createElement('div', { className: 'docs-section docs-navigation' });
        
        const title = Utils.createElement('h2', {}, ['Additional Resources']);
        
        const links = Utils.createElement('div', { className: 'docs-links' }, [
            this.createDocLink('/docs/api', 'API Documentation', 'Complete REST API reference with interactive examples'),
            this.createDocLink('https://github.com/SierraSoftworks/tailon', 'GitHub Repository', 'Source code, issues, and contributions'),
            this.createDocLink('https://tailscale.com/kb/', 'Tailscale Documentation', 'Learn more about Tailscale network setup and ACLs')
        ]);

        section.appendChild(title);
        section.appendChild(links);
        
        return section;
    }

    // Helper method to create subsections
    createSubSection(title, code) {
        const subsection = Utils.createElement('div', { className: 'docs-subsection' });
        
        const subTitle = Utils.createElement('h3', {}, [title]);
        const codeBlock = Utils.createElement('pre', { className: 'config-example' }, [code]);
        
        subsection.appendChild(subTitle);
        subsection.appendChild(codeBlock);
        
        return subsection;
    }

    // Helper method to create documentation links
    createDocLink(href, title, description) {
        const isExternal = href.startsWith('http');
        
        const link = Utils.createElement('a', {
            href: href,
            className: 'docs-link',
            ...(isExternal ? { target: '_blank', rel: 'noopener noreferrer' } : {}),
            ...(isExternal ? {} : { 
                onclick: (e) => {
                    e.preventDefault();
                    window.router.navigate(href);
                }
            })
        }, [
            Utils.createElement('h4', {}, [title]),
            Utils.createElement('p', {}, [description])
        ]);

        return link;
    }

    // Render error state
    renderError(message) {
        const content = document.getElementById('content');
        content.innerHTML = '';

        const title = Utils.createElement('div', { className: 'page-title' }, [
            'Documentation'
        ]);

        const errorDiv = Utils.createElement('div', { className: 'error-message' }, [
            message
        ]);

        const retryButton = Utils.createElement('button', {
            className: 'btn btn-primary',
            onclick: () => this.init()
        }, ['Retry']);

        content.appendChild(title);
        content.appendChild(errorDiv);
        content.appendChild(retryButton);
    }

    // Cleanup when component is destroyed
    destroy() {
        this.container = null;
    }
}
