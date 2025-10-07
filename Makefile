.PHONY: run build clean help

# Variables
APP_NAME=godbadmin
VERSION?=1.0.0
DIST_DIR=dist
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

# Default target
help:
	@echo "Available targets:"
	@echo "  make run     - Run the application"
	@echo "  make build   - Build for multiple platforms"
	@echo "  make clean   - Clean build artifacts"

# Run the application
run:
	@echo "Starting $(APP_NAME)..."
	@go run main.go

# Build for multiple platforms
build: clean
	@echo "Building $(APP_NAME) v$(VERSION) for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		OUTPUT_NAME=$(DIST_DIR)/$(APP_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then \
			OUTPUT_NAME=$$OUTPUT_NAME.exe; \
		fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build -o $$OUTPUT_NAME -ldflags="-s -w" .; \
		if [ $$? -eq 0 ]; then \
			echo "  ✓ Built: $$OUTPUT_NAME"; \
		else \
			echo "  ✗ Failed: $$OUTPUT_NAME"; \
		fi; \
	done
	@echo "Build complete! Binaries are in $(DIST_DIR)/"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(DIST_DIR)
	@echo "Clean complete!"

# Build for current platform only
build-local:
	@echo "Building $(APP_NAME) for current platform..."
	@go build -o $(APP_NAME) -ldflags="-s -w" .
	@echo "Build complete: $(APP_NAME)"
