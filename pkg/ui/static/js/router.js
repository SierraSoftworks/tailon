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
            // Use HTML5 pushState for clean URLs
            window.history.pushState({}, '', path);
            this.handleRoute();
        }
    }

    // Handle current route
    handleRoute() {
        const path = this.getCurrentPath();
        console.log('Router handling path:', path); // Debug log
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
        const pathname = window.location.pathname;
        
        // Use pathname directly for HTML5 pushState routing
        if (pathname && pathname !== '/') {
            return pathname.endsWith('/') ? pathname.slice(0, -1) : pathname;
        }
        
        return '/';
    }

    // Update navigation active states
    updateNavigation(currentPath) {
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
            
            // Check if this link matches current path
            const href = link.getAttribute('href');
            if (href === currentPath || 
                (currentPath === '/' && (href === '#' || href === '/'))) {
                link.classList.add('active');
            }
        });
    }

    // Start the router
    start() {
        // Handle initial route on page load
        this.handleRoute();
    }
}
