// API client for communicating with the backend
const API = {
    baseURL: '',

    // Initialize API client
    init() {
        this.baseURL = window.location.origin;
    },

    // Generic request handler
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
            },
        };

        const finalOptions = { ...defaultOptions, ...options };

        try {
            const response = await fetch(url, finalOptions);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                return await response.json();
            }
            
            return await response.text();
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    },

    // Get all applications
    async getApplications() {
        return await this.request('/api/v1/apps');
    },

    // Get specific application
    async getApplication(name) {
        return await this.request(`/api/v1/apps/${name}`);
    },

    // Start application
    async startApplication(name) {
        return await this.request(`/api/v1/apps/${name}/start`, { method: 'POST' });
    },

    // Stop application
    async stopApplication(name) {
        return await this.request(`/api/v1/apps/${name}/stop`, { method: 'POST' });
    },

    // Restart application
    async restartApplication(name) {
        return await this.request(`/api/v1/apps/${name}/restart`, { method: 'POST' });
    },

    // Get application logs
    async getLogs(name) {
        return await this.request(`/api/v1/apps/${name}/logs`);
    },

    // Create EventSource for log streaming
    createLogStream(name) {
        return new EventSource(`${this.baseURL}/api/v1/apps/${name}/logs?stream=true`);
    }
};
