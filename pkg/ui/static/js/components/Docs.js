// Docs component for API documentation view
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
            'API Documentation'
        ]);

        // Create docs container
        this.container = Utils.createElement('div', { className: 'docs-container' });

        // Create Swagger UI header
        const header = this.createSwaggerHeader();
        
        // Create Swagger UI container
        const swaggerContainer = Utils.createElement('div', { id: 'swagger-ui' });

        this.container.appendChild(header);
        this.container.appendChild(swaggerContainer);

        content.appendChild(title);
        content.appendChild(this.container);

        // Load Swagger UI
        this.loadSwaggerUI();
    }

    // Create docs header
    createHeader() {
        const header = Utils.createElement('div', { className: 'docs-header' }, [
            Utils.createElement('h1', {}, ['tail-on API']),
            Utils.createElement('p', {}, ['A web service for managing applications over Tailscale']),
            Utils.createElement('div', { className: 'base-url' }, [`Base URL: ${this.baseURL}`])
        ]);

        return header;
    }

    // Create Swagger UI header
    createSwaggerHeader() {
        const header = Utils.createElement('div', { className: 'swagger-header' }, [
            Utils.createElement('h1', {}, ['tail-on API Documentation']),
            Utils.createElement('p', {}, ['Interactive API Explorer powered by Swagger UI']),
            this.createNavigation()
        ]);

        return header;
    }

    // Load Swagger UI
    loadSwaggerUI() {
        // Check if Swagger UI is already loaded
        if (window.SwaggerUIBundle) {
            this.initializeSwaggerUI();
            return;
        }

        // Load Swagger UI CSS
        const cssLink = document.createElement('link');
        cssLink.rel = 'stylesheet';
        cssLink.href = 'https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui.css';
        document.head.appendChild(cssLink);

        // Load Swagger UI JS
        const script1 = document.createElement('script');
        script1.src = 'https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui-bundle.js';
        script1.crossOrigin = 'anonymous';
        
        const script2 = document.createElement('script');
        script2.src = 'https://unpkg.com/swagger-ui-dist@5.10.3/swagger-ui-standalone-preset.js';
        script2.crossOrigin = 'anonymous';

        // Initialize Swagger UI after scripts load
        script2.onload = () => {
            this.initializeSwaggerUI();
        };

        document.head.appendChild(script1);
        document.head.appendChild(script2);
    }

    // Initialize Swagger UI
    initializeSwaggerUI() {
        if (!window.SwaggerUIBundle) {
            console.error('SwaggerUIBundle not loaded');
            return;
        }

        window.ui = SwaggerUIBundle({
            url: `${this.baseURL}/docs/openapi.yaml`,
            dom_id: '#swagger-ui',
            deepLinking: true,
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIStandalonePreset
            ],
            plugins: [
                SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "StandaloneLayout",
            tryItOutEnabled: true,
            requestInterceptor: (request) => {
                // Ensure requests go to the correct base URL
                if (request.url.startsWith('/')) {
                    request.url = this.baseURL + request.url;
                }
                return request;
            },
            onComplete: () => {
                console.log('Swagger UI loaded successfully');
            },
            onFailure: (error) => {
                console.error('Failed to load Swagger UI:', error);
            }
        });
    }

    // Create navigation tabs
    createNavigation() {
        const nav = Utils.createElement('div', { className: 'swagger-nav' });

        const links = [
            { label: 'OpenAPI Spec', href: '/docs/openapi.yaml' },
            { label: 'GitHub Repository', href: 'https://github.com/SierraSoftworks/tail-on' },
            { label: 'Tailscale', href: 'https://tailscale.com/' }
        ];

        links.forEach(link => {
            const linkEl = Utils.createElement('a', {
                href: link.href,
                target: '_blank',
                className: 'swagger-nav-link'
            }, [link.label]);

            nav.appendChild(linkEl);
        });

        return nav;
    }

    // Render error state
    renderError(message) {
        const content = document.getElementById('content');
        content.innerHTML = '';

        const title = Utils.createElement('div', { className: 'page-title' }, [
            'API Documentation'
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
