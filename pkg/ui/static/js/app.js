// Main application entry point
class TailOnApp {
    constructor() {
        this.router = new Router();
        this.dashboard = new Dashboard();
        this.isInitialized = false;
    }

    // Initialize the application
    async init() {
        if (this.isInitialized) {
            return;
        }

        try {
            // Initialize API
            API.init();

            // Register routes
            this.setupRoutes();

            // Start router
            this.router.start();

            this.isInitialized = true;
            console.log('Tail-On SPA initialized successfully');

        } catch (error) {
            console.error('Failed to initialize application:', error);
            this.showError('Failed to initialize application');
        }
    }

    // Setup application routes
    setupRoutes() {
        // Dashboard route
        this.router.register('/', async () => {
            await this.showDashboard();
        });

        // API docs route
        this.router.register('/docs', () => {
            this.showApiDocs();
        });
    }

    // Show dashboard
    async showDashboard() {
        try {
            await this.dashboard.init();
        } catch (error) {
            console.error('Failed to load dashboard:', error);
            this.showError('Failed to load dashboard');
        }
    }

    // Show API documentation
    showApiDocs() {
        window.location.href = '/docs';
    }

    // Show error message
    showError(message) {
        const content = document.getElementById('content');
        content.innerHTML = '';

        const errorDiv = Utils.createElement('div', { className: 'error-message' }, [
            Utils.createElement('strong', {}, ['Error: ']),
            message
        ]);

        const retryButton = Utils.createElement('button', {
            className: 'btn btn-primary',
            onclick: () => {
                this.isInitialized = false;
                this.init();
            }
        }, ['Retry']);

        const container = Utils.createElement('div', {}, [
            errorDiv,
            Utils.createElement('div', { style: 'margin-top: 1rem;' }, [retryButton])
        ]);

        content.appendChild(container);
    }

    // Auto-refresh dashboard every 30 seconds
    startAutoRefresh() {
        setInterval(() => {
            if (this.router.getCurrentPath() === '/' && this.dashboard) {
                this.dashboard.refresh();
            }
        }, 30000);
    }
}

// Global app instance
window.App = new TailOnApp();

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', async () => {
    await App.init();
    App.startAutoRefresh();
});
