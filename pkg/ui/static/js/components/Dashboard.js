// Dashboard component for managing the main application view
class Dashboard {
    constructor() {
        this.apps = {};
        this.applicationCards = new Map();
        this.currentExpandedApp = null;
        this.container = null;
    }

    // Initialize the dashboard
    async init() {
        try {
            await this.loadApplications();
            this.render();
        } catch (error) {
            console.error('Failed to initialize dashboard:', error);
            this.renderError('Failed to load applications');
        }
    }

    // Load applications from API
    async loadApplications() {
        this.apps = await API.getApplications();
    }

    // Render the dashboard
    render() {
        const content = document.getElementById('content');
        content.innerHTML = '';

        // Page title
        const title = Utils.createElement('div', { className: 'page-title' }, [
            'Applications Dashboard'
        ]);

        if (Object.keys(this.apps).length === 0) {
            // No applications message
            const noAppsCard = Utils.createElement('div', { className: 'card text-center' }, [
                Utils.createElement('div', { className: 'card-header' }, [
                    Utils.createElement('h2', { className: 'card-title' }, [
                        'No Applications Configured'
                    ])
                ]),
                Utils.createElement('p', {}, [
                    'There are no applications configured in your TailOn instance.'
                ]),
                Utils.createElement('p', {}, [
                    'To add applications, update your configuration file and restart the service.'
                ])
            ]);

            content.appendChild(title);
            content.appendChild(noAppsCard);
            return;
        }

        // Create accordion container
        this.container = Utils.createElement('div', { className: 'app-accordion' });

        // Create application cards
        Object.entries(this.apps).forEach(([appName, app]) => {
            const card = new ApplicationCard(
                appName, 
                app, 
                (name, isExpanded) => this.handleCardToggle(name, isExpanded)
            );
            
            this.applicationCards.set(appName, card);
            this.container.appendChild(card.render());
        });

        content.appendChild(title);
        content.appendChild(this.container);
    }

    // Handle card expand/collapse
    handleCardToggle(appName, isExpanded) {
        // Close currently expanded app if different
        if (this.currentExpandedApp && this.currentExpandedApp !== appName) {
            const currentCard = this.applicationCards.get(this.currentExpandedApp);
            if (currentCard) {
                currentCard.collapse();
            }
        }

        this.currentExpandedApp = isExpanded ? appName : null;
    }

    // Refresh dashboard data
    async refresh() {
        try {
            await this.loadApplications();
            
            // Update existing cards or recreate if apps changed
            const currentApps = new Set(this.applicationCards.keys());
            const newApps = new Set(Object.keys(this.apps));

            // Check if app list changed
            const appsChanged = currentApps.size !== newApps.size || 
                               ![...currentApps].every(app => newApps.has(app));

            if (appsChanged) {
                // Complete re-render if apps added/removed
                this.destroy();
                this.render();
            } else {
                // Update existing cards
                Object.entries(this.apps).forEach(([appName, app]) => {
                    const card = this.applicationCards.get(appName);
                    if (card) {
                        card.updateApp(app);
                    }
                });
            }
        } catch (error) {
            console.error('Failed to refresh dashboard:', error);
            Utils.showToast('Failed to refresh applications', 'error');
        }
    }

    // Render error state
    renderError(message) {
        const content = document.getElementById('content');
        content.innerHTML = '';

        const title = Utils.createElement('div', { className: 'page-title' }, [
            'Applications Dashboard'
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
        this.applicationCards.forEach(card => card.destroy());
        this.applicationCards.clear();
        this.currentExpandedApp = null;
        this.container = null;
    }
}
