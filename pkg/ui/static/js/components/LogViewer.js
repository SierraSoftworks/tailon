// LogViewer component for displaying real-time logs
class LogViewer {
    constructor(appName) {
        this.appName = appName;
        this.eventSource = null;
        this.container = null;
        this.isActive = false;
        this.maxLines = 500;
    }

    // Create the log viewer DOM structure
    render() {
        const header = Utils.createElement('div', { className: 'logs-header' }, [
            Utils.createElement('h3', {}, ['Live Logs']),
            Utils.createElement('button', {
                className: 'btn btn-sm',
                onclick: () => this.clearLogs(),
                dataset: { tooltip: 'Clear Logs' }
            }, [Icons.clear()])
        ]);

        this.container = Utils.createElement('div', {
            className: 'logs-container',
            id: `logs-${this.appName}`
        }, [
            Utils.createElement('div', { className: 'log-placeholder' }, [
                'Click to expand and view live logs...'
            ])
        ]);

        return Utils.createElement('div', { className: 'logs-section' }, [
            header,
            this.container
        ]);
    }

    // Start log streaming
    startStreaming() {
        if (this.isActive || this.eventSource) {
            return;
        }

        this.isActive = true;
        this.container.classList.add('active');
        this.container.innerHTML = '';
        this.appendLog('Connecting to log stream...', new Date().toISOString());

        try {
            this.eventSource = API.createLogStream(this.appName);

            this.eventSource.onopen = () => {
                this.container.innerHTML = '';
                this.appendLog('Connected to log stream.', new Date().toISOString(), 'info');
            };

            this.eventSource.onmessage = (event) => {
                try {
                    const logData = JSON.parse(event.data);
                    this.appendLog(logData.message, logData.timestamp, logData.level, logData.source);
                } catch (e) {
                    // Handle plain text messages
                    this.appendLog(event.data, new Date().toISOString());
                }
            };

            this.eventSource.onerror = (event) => {
                console.error('SSE error for', this.appName, event);
                this.appendLog(
                    'Connection error - attempting to reconnect...',
                    new Date().toISOString(),
                    'error'
                );
            };

        } catch (error) {
            console.error('Failed to start log stream:', error);
            this.container.innerHTML = '';
            this.appendLog('Failed to connect to log stream.', new Date().toISOString(), 'error');
        }
    }

    // Stop log streaming
    stopStreaming() {
        if (this.eventSource) {
            this.eventSource.close();
            this.eventSource = null;
        }

        this.isActive = false;
        if (this.container) {
            this.container.classList.remove('active');
        }
    }

    // Append a log line
    appendLog(message, timestamp, level = null, source = 'stdout') {
        if (!this.container) return;

        // Determine source class for styling
        let sourceClass = 'log-stdout'; // default
        if (source === 'stderr') {
            sourceClass = 'log-stderr';
        } else if (source === 'audit') {
            sourceClass = 'log-audit';
        }

        const logLine = Utils.createElement('div', { className: `log-line ${sourceClass}` }, [
            Utils.createElement('span', { className: 'log-timestamp' }, [
                Utils.formatTimestamp(timestamp)
            ]),
            Utils.createElement('span', {
                className: `log-message${level ? ` log-${level}` : ''}`
            }, [Utils.escapeHtml(message)])
        ]);

        this.container.appendChild(logLine);

        // Auto-scroll to bottom
        this.container.scrollTop = this.container.scrollHeight;

        // Limit log lines to prevent memory issues
        const logLines = this.container.querySelectorAll('.log-line');
        if (logLines.length > this.maxLines) {
            logLines[0].remove();
        }
    }

    // Clear logs
    clearLogs() {
        if (this.container) {
            this.container.innerHTML = '';
            this.appendLog('Logs cleared.', new Date().toISOString(), 'info');
        }
    }

    // Cleanup when component is destroyed
    destroy() {
        this.stopStreaming();
        this.container = null;
    }
}
