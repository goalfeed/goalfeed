#!/bin/bash

# Goalfeed Web Interface Development Script

echo "🏒 Goalfeed Web Interface Setup"
echo "================================"

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the goalfeed root directory"
    exit 1
fi

# Function to start the web server (single command)
start_web_server() {
    echo "🚀 Starting Goalfeed web server..."
    echo "📝 This will build the React frontend and start the server"
    echo "🌐 Web interface will be available at: http://localhost:8080"
    echo ""
    ./goalfeed --web --web-port 8080 --cfl "*" --nhl "*" --mlb "*"
}

# Function to start development mode (separate processes)
start_dev_mode() {
    echo "🔄 Starting development environment..."
    echo "📝 This will start the backend. Open another terminal and run:"
    echo "   $0 frontend"
    echo ""
    ./goalfeed --web --web-port 8080 --cfl "*" --nhl "*" --mlb "*"
}

# Function to start the frontend (development only)
start_frontend() {
    echo "🎨 Starting React frontend in development mode..."
    cd web/frontend
    npm install
    npm start
}

# Function to build the frontend
build_frontend() {
    echo "🔨 Building React frontend..."
    cd web/frontend
    npm install
    npm run build
    cd ../..
    echo "✅ Frontend built successfully!"
}

# Function to show help
show_help() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  web        - Start web server (builds frontend and serves everything)"
    echo "  frontend   - Start React development server (for development only)"
    echo "  build      - Build the React frontend for production"
    echo "  dev        - Start backend only (for development with separate frontend)"
    echo "  help       - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 web         # Start complete web server on port 8080"
    echo "  $0 frontend    # Start React dev server on port 3000 (development)"
    echo "  $0 build       # Build frontend for production"
    echo ""
    echo "Recommended: Use 'web' command for single-command operation"
}

# Main script logic
case "$1" in
    "web")
        start_web_server
        ;;
    "frontend")
        start_frontend
        ;;
    "build")
        build_frontend
        ;;
    "dev")
        start_dev_mode
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        echo "❌ Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
