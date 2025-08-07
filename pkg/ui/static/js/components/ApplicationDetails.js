// ApplicationDetails component for showing expanded app information
class ApplicationDetails {
    constructor(app, appName) {
        this.app = app;
        this.appName = appName;
        this.logViewer = new LogViewer(appName);
    }

    // Create the application details DOM structure
    render() {
        const detailsSection = Utils.createElement('div', { className: 'app-details-section' }, [
            Utils.createElement('h3', {}, ['Application Details']),
            this.renderDetailItem('Command', this.renderCommand()),
            ...this.renderEnvironmentVariables(),
            ...this.renderProcessInfo()
        ]);

        const grid = Utils.createElement('div', { className: 'app-details-grid' }, [
            detailsSection,
            this.logViewer.render()
        ]);

        return Utils.createElement('div', { className: 'app-accordion-content' }, [grid]);
    }

    // Render command detail item
    renderCommand() {
        const command = `${this.app.config.path} ${this.app.config.args?.join(' ') || ''}`.trim();
        return Utils.createElement('code', {}, [command]);
    }

    // Render environment variables if they exist
    renderEnvironmentVariables() {
        if (!this.app.config.env || this.app.config.env.length === 0) {
            return [];
        }

        const envVars = Utils.createElement('div', { className: 'env-vars' }, 
            this.app.config.env.map(envVar => 
                Utils.createElement('div', {}, [
                    Utils.createElement('code', {}, [envVar])
                ])
            )
        );

        return [this.renderDetailItem('Environment Variables', envVars)];
    }

    // Render process information if running
    renderProcessInfo() {
        if (!this.app.running) {
            return [];
        }

        return [this.renderDetailItem('Process ID', this.app.pid.toString())];
    }

    // Helper to create detail items
    renderDetailItem(label, content) {
        return Utils.createElement('div', { className: 'detail-item' }, [
            Utils.createElement('strong', {}, [label + ':']),
            Utils.createElement('br'),
            content
        ]);
    }

    // Start log streaming when expanded
    startLogStreaming() {
        this.logViewer.startStreaming();
    }

    // Stop log streaming when collapsed
    stopLogStreaming() {
        this.logViewer.stopStreaming();
    }

    // Update app data
    updateApp(app) {
        this.app = app;
    }

    // Cleanup when component is destroyed
    destroy() {
        this.logViewer.destroy();
    }
}
