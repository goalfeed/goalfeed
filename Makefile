.PHONY: frontend frontend-install backend build all clean dev docs

FRONTEND_DIR := web/frontend
BACKEND_BIN := goalfeed
# Pinned so local and CI generation produce byte-identical output
SWAG_VERSION := v1.16.4

# Build the React frontend (installs deps if needed)
frontend: frontend-install
	cd $(FRONTEND_DIR) && npm run build

# Install frontend dependencies using lockfile
frontend-install:
	cd $(FRONTEND_DIR) && npm ci

# Build the Go backend binary
backend:
	GO111MODULE=on go build -o $(BACKEND_BIN) .

# Regenerate the Swagger/OpenAPI docs in docs/ (docs.go is compiled into the binary)
# CI verifies these are in sync; run this after changing web/api/** or models/**.
docs:
	go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init \
		-g web/api/server.go -o docs --parseDependency --parseInternal

# Build both frontend and backend
build: frontend backend

# Run app with frontend and hot reloading (watches Go + React, rebuilds and restarts on change)
# Requires: fswatch (brew install fswatch)
dev:
	bash dev.sh dev

# Convenience target
all: build

# Clean generated artifacts
clean:
	rm -f $(BACKEND_BIN)
	rm -rf $(FRONTEND_DIR)/build


