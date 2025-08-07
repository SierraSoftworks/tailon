// Utility functions for the SPA
const Utils = {
    // Create DOM elements with attributes and children
    createElement(tag, attributes = {}, children = []) {
        const element = document.createElement(tag);
        
        // Set attributes
        Object.entries(attributes).forEach(([key, value]) => {
            if (key === 'className') {
                element.className = value;
            } else if (key === 'onclick') {
                element.onclick = value;
            } else if (key === 'dataset') {
                Object.entries(value).forEach(([dataKey, dataValue]) => {
                    element.dataset[dataKey] = dataValue;
                });
            } else {
                element.setAttribute(key, value);
            }
        });
        
        // Add children
        children.forEach(child => {
            if (typeof child === 'string') {
                element.appendChild(document.createTextNode(child));
            } else if (child instanceof Node) {
                element.appendChild(child);
            }
        });
        
        return element;
    },

    // Create SVG icons
    createSVG(viewBox, paths) {
        const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        svg.setAttribute('class', 'btn-icon');
        svg.setAttribute('viewBox', viewBox);
        svg.setAttribute('fill', 'none');
        svg.setAttribute('stroke', 'currentColor');
        svg.setAttribute('stroke-width', '2');
        svg.setAttribute('stroke-linecap', 'round');
        svg.setAttribute('stroke-linejoin', 'round');
        
        paths.forEach(pathData => {
            const element = document.createElementNS('http://www.w3.org/2000/svg', pathData.type);
            Object.entries(pathData.attrs).forEach(([key, value]) => {
                element.setAttribute(key, value);
            });
            svg.appendChild(element);
        });
        
        return svg;
    },

    // Escape HTML to prevent XSS
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    },

    // Format timestamp
    formatTimestamp(timestamp) {
        return new Date(timestamp).toLocaleTimeString();
    },

    // Show toast notification
    showToast(message, type = 'info') {
        const container = document.getElementById('toast-container');
        
        // Remove existing toasts
        container.innerHTML = '';
        
        const toast = Utils.createElement('div', {
            className: `toast alert alert-${type === 'error' ? 'error' : 'success'}`,
            style: `
                padding: 1rem 1.5rem;
                border-radius: 8px;
                font-weight: 500;
                margin-bottom: 0.5rem;
                animation: slideIn 0.3s ease-out;
            `
        }, [message]);
        
        container.appendChild(toast);
        
        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (toast.parentNode) {
                toast.remove();
            }
        }, 5000);
    },

    // Debounce function calls
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }
};

// Icon definitions
const Icons = {
    play: () => Utils.createSVG('0 0 24 24', [
        { type: 'polygon', attrs: { points: '5,3 19,12 5,21' } }
    ]),
    
    stop: () => Utils.createSVG('0 0 24 24', [
        { type: 'rect', attrs: { x: '6', y: '6', width: '12', height: '12' } }
    ]),
    
    restart: () => Utils.createSVG('0 0 24 24', [
        { type: 'polyline', attrs: { points: '23 4 23 10 17 10' } },
        { type: 'polyline', attrs: { points: '1 20 1 14 7 14' } },
        { type: 'path', attrs: { d: 'm3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15' } }
    ]),
    
    logs: () => Utils.createSVG('0 0 24 24', [
        { type: 'line', attrs: { x1: '3', y1: '6', x2: '21', y2: '6' } },
        { type: 'line', attrs: { x1: '3', y1: '12', x2: '21', y2: '12' } },
        { type: 'line', attrs: { x1: '3', y1: '18', x2: '21', y2: '18' } }
    ]),
    
    chevronDown: () => Utils.createSVG('0 0 24 24', [
        { type: 'polyline', attrs: { points: '6,9 12,15 18,9' } }
    ]),
    
    clear: () => Utils.createSVG('0 0 24 24', [
        { type: 'polyline', attrs: { points: '3,6 5,6 21,6' } },
        { type: 'path', attrs: { d: 'm19,6 v14 a2,2 0 0,1 -2,2 H7 a2,2 0 0,1 -2,-2 V6 m3,0 V4 a2,2 0 0,1 2,-2 h4 a2,2 0 0,1 2,2 v2' } }
    ])
};
