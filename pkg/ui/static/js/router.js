// Simple router for SPA navigation
class Router {
    constructor() {
        this.routes = new Map();
        this.currentRoute = null;
        
        // Listen for browser navigation
        window.addEventListener('popstate', () => {
            this.handleRoute();
        });
    }

    // Register a route
    register(path, handler) {
        this.routes.set(path, handler);
    }

    // Navigate to a route
    navigate(path) {
        if (path !== this.currentRoute) {
            this.currentRoute = path;
            window.history.pushState({}, '', path === '/' ? '/' : `#${path}`);
            this.handleRoute();
        }
    }

    // Handle current route
    handleRoute() {
        const path = this.getCurrentPath();
        const handler = this.routes.get(path);
        
        if (handler) {
            handler();
        } else if (this.routes.has('/')) {
            // Fallback to home route
            this.routes.get('/')();
        } else {
            console.warn('No route handler found for:', path);
        }
        
        this.updateNavigation(path);
    }

    // Get current path from URL
    getCurrentPath() {
        const hash = window.location.hash;
        if (hash.startsWith('#/')) {
            return hash.substring(1);
        } else if (hash.startsWith('#')) {
            return hash;
        }
        return '/';
    }

    // Update navigation active states
    updateNavigation(currentPath) {
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
            
            // Check if this link matches current path
            const href = link.getAttribute('href');
            if ((currentPath === '/' && href === '#') || 
                (currentPath !== '/' && href === `#${currentPath}`)) {
                link.classList.add('active');
            }
        });
    }

    // Start the router
    start() {
        this.handleRoute();
    }
}
