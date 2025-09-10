#!/bin/bash

# Goalfeed Development Helper Script
# This script monitors for changes and automatically rebuilds components

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[$(date +'%H:%M:%S')]${NC} ✅ $1"
}

print_warning() {
    echo -e "${YELLOW}[$(date +'%H:%M:%S')]${NC} ⚠️  $1"
}

print_error() {
    echo -e "${RED}[$(date +'%H:%M:%S')]${NC} ❌ $1"
}

# Function to rebuild frontend
rebuild_frontend() {
    print_status "Rebuilding React frontend..."
    cd web/frontend
    if npm run build; then
        print_success "Frontend build completed successfully"
    else
        print_error "Frontend build failed"
        return 1
    fi
    cd ../..
}

# Function to rebuild backend
rebuild_backend() {
    print_status "Rebuilding Go backend..."
    if go build -o goalfeed .; then
        print_success "Backend build completed successfully"
    else
        print_error "Backend build failed"
        return 1
    fi
}

# Function to restart server
restart_server() {
    print_status "Restarting Goalfeed server..."
    pkill -f goalfeed || true
    sleep 1
    ./goalfeed --web --web-port 8080 --cfl "*" --nhl "*" --mlb "*" &
    print_success "Server restarted on port 8080"
}

# Function to watch for changes
watch_changes() {
    print_status "Starting file watcher..."
    print_status "Watching for changes in:"
    print_status "  - web/frontend/src/ (will rebuild frontend)"
    print_status "  - *.go files (will rebuild backend)"
    print_status "  - web/api/ (will rebuild backend)"
    print_status ""
    print_status "Press Ctrl+C to stop watching"
    
    # Use fswatch if available, otherwise fall back to basic monitoring
    if command -v fswatch >/dev/null 2>&1; then
        fswatch -o web/frontend/src/ web/api/ *.go | while read; do
            if [[ "$REPLY" == *"web/frontend/src"* ]] || [[ "$REPLY" == *"web/frontend"* ]]; then
                print_status "Frontend files changed, rebuilding..."
                if rebuild_frontend; then
                    restart_server
                fi
            elif [[ "$REPLY" == *".go"* ]] || [[ "$REPLY" == *"web/api"* ]]; then
                print_status "Backend files changed, rebuilding..."
                if rebuild_backend; then
                    restart_server
                fi
            fi
        done
    else
        print_warning "fswatch not found. Install with: brew install fswatch"
        print_status "Falling back to manual rebuild mode..."
        print_status "Run './dev.sh rebuild' to rebuild manually"
    fi
}

# Function to run development server
run_dev() {
    print_status "Starting Goalfeed development environment..."
    
    # Initial build
    rebuild_backend
    rebuild_frontend
    
    # Start server
    restart_server
    
    print_success "Development environment ready!"
    print_status "Web interface: http://localhost:8080"
    print_status "API endpoint: http://localhost:8080/api/games"
    print_status "WebSocket: ws://localhost:8080/ws"
    print_status ""
    print_status "Starting file watcher..."
    watch_changes
}

# Function to show help
show_help() {
    echo "Goalfeed Development Helper"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  dev, start    Start development environment with file watching"
    echo "  rebuild       Rebuild both frontend and backend"
    echo "  frontend      Rebuild only frontend"
    echo "  backend       Rebuild only backend"
    echo "  restart       Restart the server"
    echo "  stop          Stop the server"
    echo "  help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 dev        # Start full development environment"
    echo "  $0 rebuild    # Rebuild everything"
    echo "  $0 frontend   # Rebuild just the React frontend"
}

# Main script logic
case "${1:-dev}" in
    "dev"|"start")
        run_dev
        ;;
    "rebuild")
        rebuild_backend
        rebuild_frontend
        restart_server
        ;;
    "frontend")
        rebuild_frontend
        restart_server
        ;;
    "backend")
        rebuild_backend
        restart_server
        ;;
    "restart")
        restart_server
        ;;
    "stop")
        print_status "Stopping Goalfeed server..."
        pkill -f goalfeed || true
        print_success "Server stopped"
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac

