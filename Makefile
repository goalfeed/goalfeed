.PHONY: frontend frontend-install backend build all clean

FRONTEND_DIR := web/frontend
BACKEND_BIN := goalfeed

# Build the React frontend (installs deps if needed)
frontend: frontend-install
	cd $(FRONTEND_DIR) && npm run build

# Install frontend dependencies using lockfile
frontend-install:
	cd $(FRONTEND_DIR) && npm ci

# Build the Go backend binary
backend:
	GO111MODULE=on go build -o $(BACKEND_BIN) .

# Build both frontend and backend
build: frontend backend

# Convenience target
all: build

# Clean generated artifacts
clean:
	rm -f $(BACKEND_BIN)
	rm -rf $(FRONTEND_DIR)/build


