// ApplicationCard component for individual application management
class ApplicationCard {
    constructor(appName, app, onToggleDetails) {
        this.appName = appName;
        this.app = app;
        this.onToggleDetails = onToggleDetails;
        this.isExpanded = false;
        this.details = null;
        this.element = null;
    }

    // Create the application card DOM structure
    render() {
        const header = this.renderHeader();
        
        this.element = Utils.createElement('div', {
            className: 'app-accordion-item',
            dataset: { app: this.appName }
        }, [header]);

        return this.element;
    }

    // Render the card header
    renderHeader() {
        const nameSection = Utils.createElement('div', { className: 'app-name-section' }, [
            Utils.createElement('h2', { className: 'app-name' }, [this.appName]),
            this.renderStatus()
        ]);

        const summary = Utils.createElement('div', { className: 'app-summary' }, [
            Utils.createElement('code', {}, [
                `${this.app.config.path} ${this.app.config.args?.join(' ') || ''}`.trim()
            ])
        ]);

        const controls = this.renderControls();

        const headerContent = Utils.createElement('div', { className: 'app-header-content' }, [
            nameSection,
            summary,
            controls
        ]);

        return Utils.createElement('div', {
            className: 'app-accordion-header',
            onclick: (e) => {
                // Don't expand if clicking on action buttons (but allow expand button)
                if (e.target.closest('button') && !e.target.closest('.expand-btn')) {
                    return;
                }
                this.toggleDetails();
            }
        }, [headerContent]);
    }

    // Render application status
    renderStatus() {
        const statusClass = this.app.running ? 'running' : 'stopped';
        const statusText = this.app.running ? 'Running' : 'Stopped';
        const pidText = this.app.running && this.app.pid ? ` (PID: ${this.app.pid})` : '';

        return Utils.createElement('div', { className: `app-status ${statusClass}` }, [
            Utils.createElement('span', { className: `status-dot ${statusClass}` }),
            statusText + pidText
        ]);
    }

    // Render control buttons
    renderControls() {
        const actionButtons = this.renderActionButtons();
        const expandButton = this.renderExpandButton();

        return Utils.createElement('div', { className: 'app-controls' }, [
            actionButtons,
            expandButton
        ]);
    }

    // Render action buttons (start/stop/restart)
    renderActionButtons() {
        const buttons = [];

        if (this.app.running) {
            buttons.push(
                this.createActionButton('stop', Icons.stop(), 'Stop Application'),
                this.createActionButton('restart', Icons.restart(), 'Restart Application')
            );
        } else {
            buttons.push(
                this.createActionButton('start', Icons.play(), 'Start Application')
            );
        }

        return Utils.createElement('div', { className: 'btn-group' }, buttons);
    }

    // Create an action button
    createActionButton(action, icon, tooltip) {
        return Utils.createElement('button', {
            className: 'btn',
            onclick: (e) => {
                e.stopPropagation();
                this.performAction(action, e.target);
            },
            dataset: { tooltip }
        }, [icon]);
    }

    // Render expand/collapse button
    renderExpandButton() {
        const chevron = Icons.chevronDown();
        
        return Utils.createElement('button', {
            className: 'expand-btn',
            onclick: (e) => {
                e.stopPropagation();
                this.toggleDetails();
            },
            dataset: { tooltip: 'View Details' }
        }, [chevron]);
    }

    // Toggle details panel
    toggleDetails() {
        if (this.isExpanded) {
            this.collapse();
        } else {
            this.expand();
        }
        
        // Call the callback with the new state
        this.onToggleDetails(this.appName, this.isExpanded);
    }

    // Expand details panel
    expand() {
        if (this.isExpanded) return;

        this.isExpanded = true;
        this.element.classList.add('expanded');

        // Create and append details if not exists
        if (!this.details) {
            this.details = new ApplicationDetails(this.app, this.appName);
            this.element.appendChild(this.details.render());
        }

        // Start log streaming after animation
        setTimeout(() => {
            this.details.startLogStreaming();
        }, 300);
    }

    // Collapse details panel
    collapse() {
        if (!this.isExpanded) return;

        this.isExpanded = false;
        this.element.classList.remove('expanded');

        if (this.details) {
            this.details.stopLogStreaming();
        }
    }

    // Perform application action
    async performAction(action, button) {
        // Show confirmation for destructive actions
        if ((action === 'stop' || action === 'restart') && 
            !confirm(`Are you sure you want to ${action} ${this.appName}?`)) {
            return;
        }

        this.setButtonLoading(button, true);

        try {
            let result;
            switch (action) {
                case 'start':
                    result = await API.startApplication(this.appName);
                    break;
                case 'stop':
                    result = await API.stopApplication(this.appName);
                    break;
                case 'restart':
                    result = await API.restartApplication(this.appName);
                    break;
                default:
                    throw new Error(`Unknown action: ${action}`);
            }

            const actionPast = action === 'stop' ? 'stopped' : 
                              action === 'start' ? 'started' : 'restarted';
            
            Utils.showToast(`Successfully ${actionPast} ${this.appName}`, 'success');

            // Refresh the dashboard after a short delay
            setTimeout(() => {
                App.dashboard.refresh();
            }, 1000);

        } catch (error) {
            console.error(`Error ${action}ing application:`, error);
            Utils.showToast(`Failed to ${action} ${this.appName}: ${error.message}`, 'error');
        } finally {
            this.setButtonLoading(button, false);
        }
    }

    // Set button loading state
    setButtonLoading(button, loading) {
        if (loading) {
            button.disabled = true;
            button.style.opacity = '0.6';
        } else {
            button.disabled = false;
            button.style.opacity = '1';
        }
    }

    // Update application data
    updateApp(app) {
        this.app = app;
        if (this.details) {
            this.details.updateApp(app);
        }
        
        // Re-render the header to reflect changes
        const newHeader = this.renderHeader();
        const oldHeader = this.element.querySelector('.app-accordion-header');
        this.element.replaceChild(newHeader, oldHeader);
    }

    // Cleanup when component is destroyed
    destroy() {
        if (this.details) {
            this.details.destroy();
        }
    }
}
