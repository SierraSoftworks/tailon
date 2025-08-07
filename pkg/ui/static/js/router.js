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
            // Always use hash-based navigation for consistency
            const newUrl = path === '/' ? '/' : `/#${path}`;
            window.history.pushState({}, '', newUrl);
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
        const hash = window.location.hash;
        const pathname = window.location.pathname;
        
        // Check hash first (for SPA navigation)
        if (hash.startsWith('#/')) {
            return hash.substring(1);
        } else if (hash.startsWith('#')) {
            // Convert #docs to /docs format
            return '/' + hash.substring(1);
        }
        
        // Check pathname for direct URL access
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
            if ((currentPath === '/' && href === '#') || 
                (currentPath !== '/' && (href === `#${currentPath}` || href === `#${currentPath.substring(1)}`))) {
                link.classList.add('active');
            }
        });
    }

    // Start the router
    start() {
        // Handle initial route on page load
        const initialPath = this.getCurrentPath();
        if (initialPath !== '/' && window.location.pathname !== '/') {
            // If we loaded with a specific pathname, convert to hash-based navigation
            this.navigate(initialPath);
        } else {
            this.handleRoute();
        }
    }
}
