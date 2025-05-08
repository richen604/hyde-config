# Makefile for hyde-config

# Variables
GO := go
BINARY_NAME := hyde-config
LOCAL_LIB_DIR := $(HOME)/.local/lib/hyde
CONFIG_DIR := $(HOME)/.config/hyde
STATE_DIR := $(HOME)/.local/state/hyde
SERVICE_DIR := $(HOME)/.config/systemd/user
VERSION := 0.1.0
BUILD_DIR := build
RELEASE_BIN_DIR := bin
GOFLAGS := -v

.PHONY: all build clean install uninstall run setup-dirs deps service-install service-enable service-disable service-start service-stop service-status release

all: build

# Download dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Build the binary
build: deps
	mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)

# Build release version
release: deps
	mkdir -p $(RELEASE_BIN_DIR)
	$(GO) build -ldflags="-s -w" -o $(RELEASE_BIN_DIR)/$(BINARY_NAME)
	@echo "Release binary built in $(RELEASE_BIN_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -rf $(RELEASE_BIN_DIR)
	$(GO) clean

# Setup required directories
setup-dirs:
	mkdir -p $(CONFIG_DIR)
	mkdir -p $(STATE_DIR)
	mkdir -p $(SERVICE_DIR)
	mkdir -p $(LOCAL_LIB_DIR)

# Run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run as daemon
run-daemon: build
	./$(BUILD_DIR)/$(BINARY_NAME) --daemon

# Install the binary to ~/.local/lib/hyde directory
install: build setup-dirs
	install -Dm755 $(BUILD_DIR)/$(BINARY_NAME) $(LOCAL_LIB_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(LOCAL_LIB_DIR)"

# Install the systemd service file
service-install: install
	install -Dm644 hyde-config.service $(SERVICE_DIR)/hyde-config.service
	@echo "Installed systemd service file to $(SERVICE_DIR)"
	@echo "To enable and start the service, run: 'make service-enable && make service-start'"

# Enable the systemd service (start on boot)
service-enable:
	systemctl --user enable hyde-config.service
	@echo "Service enabled. It will start automatically on login."

# Start the systemd service
service-start:
	systemctl --user start hyde-config.service
	@echo "Service started."

# Stop the systemd service
service-stop:
	systemctl --user stop hyde-config.service
	@echo "Service stopped."

# Disable the systemd service
service-disable:
	systemctl --user disable hyde-config.service
	@echo "Service disabled. It will no longer start automatically on login."

# Check service status
service-status:
	systemctl --user status hyde-config.service

# Uninstall the binary from system
uninstall: service-stop service-disable
	rm -f $(LOCAL_LIB_DIR)/$(BINARY_NAME)
	rm -f $(SERVICE_DIR)/hyde-config.service
	@echo "Uninstalled $(BINARY_NAME) from $(LOCAL_LIB_DIR)"
	@echo "Uninstalled systemd service file from $(SERVICE_DIR)"

# Create a sample config file if it doesn't exist
config-sample: setup-dirs
	@if [ ! -f $(CONFIG_DIR)/config.toml ]; then \
		echo "Creating sample config file at $(CONFIG_DIR)/config.toml"; \
		echo '# Hyde Configuration File\n\n[theme]\naccent = "blue"\ndark_mode = true\n\n[hyprland]\nbackground = "#1E1E2E"\nforeground = "#CDD6F4"\n\n[hyprland.decoration]\nrounding = 10\nblur = true\nblur_size = 3\nblur_passes = 1' > $(CONFIG_DIR)/config.toml; \
	else \
		echo "Config file already exists at $(CONFIG_DIR)/config.toml"; \
	fi

# Build for different architectures
build-all: deps
	@echo "Building for multiple platforms..."
	mkdir -p $(RELEASE_BIN_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(RELEASE_BIN_DIR)/$(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 $(GO) build -ldflags="-s -w" -o $(RELEASE_BIN_DIR)/$(BINARY_NAME)-linux-arm64
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(RELEASE_BIN_DIR)/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="-s -w" -o $(RELEASE_BIN_DIR)/$(BINARY_NAME)-darwin-arm64

# Help command
help:
	@echo "Available commands:"
	@echo "  make deps               - Download Go dependencies"
	@echo "  make build              - Build the hyde-config binary"
	@echo "  make release            - Build optimized release binary in $(RELEASE_BIN_DIR) directory"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make install            - Install hyde-config to $(LOCAL_LIB_DIR)"
	@echo "  make uninstall          - Remove hyde-config from $(LOCAL_LIB_DIR)"
	@echo "  make run                - Build and run hyde-config"
	@echo "  make run-daemon         - Run in daemon mode"
	@echo "  make service-install    - Install the systemd service file"
	@echo "  make service-enable     - Enable the service to start on boot"
	@echo "  make service-start      - Start the service"
	@echo "  make service-stop       - Stop the service"
	@echo "  make service-disable    - Disable the service from starting on boot"
	@echo "  make service-status     - Check the service status"
	@echo "  make setup-dirs         - Create required directories"
	@echo "  make config-sample      - Create a sample config file"
	@echo "  make build-all          - Build for multiple platforms"