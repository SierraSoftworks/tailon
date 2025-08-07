// ApplicationCard component for individual application management
class ApplicationCard {
    constructor(appName, app, onToggleDetails) {
        this.appName = appName;
        this.app = app;
        this.onToggleDetails = onToggleDetails;
        this.isExpanded = false;
        this.details = null;
        this.element = null;
        this.confirmationStates = new Map(); // Track button confirmation states by action
        this.confirmationTimeouts = new Map(); // Track confirmation timeouts by action
        this.runtimeUpdateTimer = null; // Timer for updating runtime display
    }

    // Create the application card DOM structure
    render() {
        const header = this.renderHeader();
        
        this.element = Utils.createElement('div', {
            className: 'app-accordion-item',
            dataset: { app: this.appName }
        }, [header]);

        // Start the runtime update timer
        this.startRuntimeTimer();

        return this.element;
    }

    // Start timer to update runtime display every second
    startRuntimeTimer() {
        // Clear any existing timer
        this.stopRuntimeTimer();
        
        // Only start timer if app has a state change time
        if (this.app.state_changed_at) {
            this.runtimeUpdateTimer = setInterval(() => {
                this.updateRuntimeDisplay();
            }, 1000);
        }
    }

    // Stop the runtime update timer
    stopRuntimeTimer() {
        if (this.runtimeUpdateTimer) {
            clearInterval(this.runtimeUpdateTimer);
            this.runtimeUpdateTimer = null;
        }
    }

    // Update just the runtime display without full re-render
    updateRuntimeDisplay() {
        if (!this.element || !this.app.state_changed_at) return;
        
        const runtimeInfoElement = this.element.querySelector('.app-runtime-info');
        if (runtimeInfoElement) {
            const newRuntimeInfo = this.renderRuntimeInfo();
            runtimeInfoElement.parentNode.replaceChild(newRuntimeInfo, runtimeInfoElement);
        }
    }

    // Render the card header
    renderHeader() {
        const nameAndStatusRow = Utils.createElement('div', { className: 'app-name-status-row' }, [
            Utils.createElement('h2', { className: 'app-name' }, [this.appName]),
            this.renderStatus()
        ]);

        const runtimeInfo = this.renderRuntimeInfo();
        
        const leftColumn = Utils.createElement('div', { className: 'app-info-column' }, [
            nameAndStatusRow,
            runtimeInfo
        ]);

        const controls = this.renderControls();

        const headerContent = Utils.createElement('div', { className: 'app-header-content' }, [
            leftColumn,
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
        let statusClass, statusText;
        
        switch (this.app.state) {
            case 'running':
                statusClass = 'running';
                statusText = 'Running';
                break;
            case 'stopping':
                statusClass = 'stopping';
                statusText = 'Stopping';
                break;
            case 'not_running':
            default:
                statusClass = 'stopped';
                statusText = 'Stopped';
                break;
        }
        
        const pidText = this.app.state === 'running' && this.app.pid ? ` (PID: ${this.app.pid})` : '';
        const exitCodeText = this.app.state === 'not_running' && this.app.last_exit_code !== undefined && this.app.last_exit_code !== 0 
            ? ` (${this.app.last_exit_code})` : '';

        return Utils.createElement('div', { className: `app-status ${statusClass}` }, [
            Utils.createElement('span', { className: `status-dot ${statusClass}` }),
            statusText + pidText + exitCodeText
        ]);
    }

    // Render runtime information (duration and user who made the state change)
    renderRuntimeInfo() {
        const info = [];

        if (this.app.state_changed_at) {
            const stateChangeTime = new Date(this.app.state_changed_at);
            const now = new Date();
            const duration = this.formatDuration(now - stateChangeTime);

            if (this.app.state === 'running') {
                info.push(`Running for ${duration}`);
            } else if (this.app.state === 'not_running') {
                info.push(`Stopped ${duration} ago`);
            } else if (this.app.state === 'stopping') {
                info.push(`Stopping for ${duration}`);
            }
        }

        if (this.app.state_changed_by && !this.app.state_changed_by.is_anonymous) {
            const user = this.app.state_changed_by;
            const userName = user.display_name || user.login_name || user.id;
            
            if (this.app.state === 'running') {
                info.push(`Started by ${userName}`);
            } else if (this.app.state === 'not_running') {
                info.push(`Stopped by ${userName}`);
            } else if (this.app.state === 'stopping') {
                info.push(`Being stopped by ${userName}`);
            }
        }

        // If no information is available, show a placeholder
        if (info.length === 0) {
            if (this.app.state === 'not_running') {
                info.push('Not running');
            } else {
                info.push('No runtime information available');
            }
        }

        return Utils.createElement('div', { className: 'app-runtime-info' }, [
            Utils.createElement('div', { className: 'runtime-text' }, [
                Utils.createElement('span', { className: 'runtime-line' }, 
                    info.map((text, index) => {
                        const separator = index > 0 ? ' • ' : '';
                        return [
                            separator,
                            Utils.createElement('span', { 
                                className: 'runtime-item' 
                            }, [text])
                        ];
                    }).flat().filter(item => item !== '') // Remove empty separators
                )
            ])
        ]);
    }

    // Format duration in human readable format
    formatDuration(milliseconds) {
        const seconds = Math.floor(milliseconds / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (days > 0) {
            return `${days}d ${hours % 24}h ${minutes % 60}m`;
        } else if (hours > 0) {
            return `${hours}h ${minutes % 60}m`;
        } else if (minutes > 0) {
            return `${minutes}m ${seconds % 60}s`;
        } else {
            return `${seconds}s`;
        }
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

        if (this.app.state === 'running') {
            buttons.push(
                this.createActionButtonWithState('stop', Icons.stop(), 'Stop Application'),
                this.createActionButtonWithState('restart', Icons.restart(), 'Restart Application')
            );
        } else if (this.app.state === 'stopping') {
            // Show force stop button when app is stopping
            buttons.push(
                this.createActionButton('force-stop', Icons.stop(), 'Force Stop Application')
            );
        } else {
            // not_running state
            buttons.push(
                this.createActionButtonWithState('start', Icons.play(), 'Start Application')
            );
        }

        return Utils.createElement('div', { className: 'btn-group' }, buttons);
    }

    // Create an action button that respects confirmation state
    createActionButtonWithState(action, icon, tooltip) {
        const confirmationState = this.confirmationStates.get(action);
        
        if (confirmationState === 'confirming') {
            const confirmText = action === 'stop' ? 'Confirm Stop' : 'Confirm Restart';
            return this.createConfirmationButton(action, 'confirm', confirmText);
        } else {
            return this.createActionButton(action, icon, tooltip);
        }
    }

    // Create a confirmation state button
    createConfirmationButton(action, state, text) {
        const icon = this.getIconForAction(action);
        
        return Utils.createElement('button', {
            className: `btn btn-${state}`,
            onclick: (e) => {
                e.stopPropagation();
                this.performAction(action, e.target);
            },
            dataset: { 
                tooltip: text,
                action
            }
        }, [icon, text]);
    }

    // Get the appropriate icon for an action
    getIconForAction(action) {
        switch (action) {
            case 'start':
                return Icons.play();
            case 'stop':
                return Icons.stop();
            case 'restart':
                return Icons.restart();
            default:
                return Icons.play();
        }
    }

    // Create an action button
    createActionButton(action, icon, tooltip) {
        const button = Utils.createElement('button', {
            className: 'btn',
            onclick: (e) => {
                e.stopPropagation();
                this.performAction(action, e.target);
            },
            dataset: { 
                tooltip,
                action
            }
        }, [icon]);
        
        return button;
    }

    // Render expand/collapse button
    renderExpandButton() {
        const chevron = Icons.chevronDown();
        chevron.classList.add('chevron'); // Add chevron class for CSS rotation
        
        return Utils.createElement('button', {
            className: 'expand-btn',
            onclick: (e) => {
                e.stopPropagation();
                this.toggleDetails();
            },
            dataset: { tooltip: this.isExpanded ? 'Hide Details' : 'View Details' }
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
        const isDangerous = action === 'stop' || action === 'restart' || action === 'force-stop';
        
        if (isDangerous && action !== 'force-stop') {
            const confirmationState = this.confirmationStates.get(action);
            
            if (!confirmationState) {
                // First click - start confirmation process
                this.startConfirmation(action, button);
                return;
            } else if (confirmationState === 'confirming') {
                // Second click - proceed with action
                this.clearConfirmation(action);
                await this.executeAction(action, button);
                return;
            }
        }
        
        // Non-dangerous action, force-stop, or already confirmed - execute immediately
        await this.executeAction(action, button);
    }

    // Start the confirmation process for dangerous actions
    startConfirmation(action, button) {
        // Set confirmation state immediately
        this.confirmationStates.set(action, 'confirming');
        this.refreshActionButtons();
        
        // Set 15-second timeout to reset
        const timeoutId = setTimeout(() => {
            if (this.confirmationStates.get(action) === 'confirming') {
                this.resetButton(action);
            }
        }, 15000);
        
        this.confirmationTimeouts.set(action, timeoutId);
    }

    // Clear confirmation state and timeout
    clearConfirmation(action) {
        this.confirmationStates.delete(action);
        const timeoutId = this.confirmationTimeouts.get(action);
        if (timeoutId) {
            clearTimeout(timeoutId);
            this.confirmationTimeouts.delete(action);
        }
    }

    // Reset button to original state
    resetButton(action) {
        this.clearConfirmation(action);
        this.refreshActionButtons();
    }

    // Refresh action buttons to reflect current state
    refreshActionButtons() {
        const btnGroup = this.element.querySelector('.btn-group');
        if (btnGroup) {
            const newActionButtons = this.renderActionButtons();
            btnGroup.parentNode.replaceChild(newActionButtons, btnGroup);
        }
    }

    // Execute the actual action
    async executeAction(action, button) {
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
                case 'force-stop':
                    result = await API.stopApplication(this.appName, true); // force stop
                    break;
                case 'restart':
                    result = await API.restartApplication(this.appName);
                    break;
                default:
                    throw new Error(`Unknown action: ${action}`);
            }

            const actionPast = action === 'stop' ? 'stopped' : 
                              action === 'force-stop' ? 'force stopped' :
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
        
        // Clear any ongoing confirmations before re-rendering
        for (const timeoutId of this.confirmationTimeouts.values()) {
            clearTimeout(timeoutId);
        }
        this.confirmationTimeouts.clear();
        this.confirmationStates.clear();
        
        // Re-render the header to reflect changes
        const newHeader = this.renderHeader();
        const oldHeader = this.element.querySelector('.app-accordion-header');
        this.element.replaceChild(newHeader, oldHeader);
        
        // Restart the runtime timer with new app data
        this.startRuntimeTimer();
    }

    // Cleanup when component is destroyed
    destroy() {
        // Clear all confirmation timeouts
        for (const timeoutId of this.confirmationTimeouts.values()) {
            clearTimeout(timeoutId);
        }
        this.confirmationTimeouts.clear();
        this.confirmationStates.clear();
        
        // Stop the runtime update timer
        this.stopRuntimeTimer();
        
        if (this.details) {
            this.details.destroy();
        }
    }
}
