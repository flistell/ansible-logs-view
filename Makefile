# Variables
BINARY_NAME=ansible-logs-view
BINARY_NAME_228=ansible-logs-view-glibc-2.28

CMD_PATH=./cmd/ansible-logs-view
TARGET_DIR=./build

.PHONY: all build run clean test help build-glibc-2.28

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(TARGET_DIR)/$(BINARY_NAME) $(CMD_PATH)

run: build
	@echo "Running $(BINARY_NAME)..."
	@$(TARGET_DIR)/$(BINARY_NAME)

clean:
	@echo "Cleaning..."
	@go clean
	@rm -f $(TARGET_DIR)/$(BINARY_NAME)
	@rm -f $(TARGET_DIR)/$(BINARY_NAME_228)

test:
	@echo "Running tests..."
	@go test ./...

help:
	@echo "Available commands:"
	@echo "  build    Build the binary"
	@echo "  run      Build and run the application"
	@echo "  clean    Remove build artifacts"
	@echo "  test     Run tests"
	@echo "  help     Show this help message"
	@echo "  build-glibc-2.28  Build the binary with glibc 2.28"

build-glibc-2.28:
	@echo "Building the ansible-logs-view-glibc-2.28 binary using podman..."
	@mkdir -p ./build
	@podman build -t ansible-logs-builder -f Dockerfile-glibc-2.28 .
	@CONTAINER_ID=$$(podman create --name ansible-logs-container ansible-logs-builder); \
	podman start $$CONTAINER_ID > /dev/null; \
	podman cp $$CONTAINER_ID:/app/ansible-logs-view-glibc-2.28 ./build/$(BINARY_NAME_228); \
	podman stop $$CONTAINER_ID > /dev/null; \
	podman rm $$CONTAINER_ID > /dev/null; \
	chmod +x ./build/ansible-logs-view-glibc-2.28
	@echo "Successfully built and copied ansible-logs-view-glibc-2.28 to local host"
	@echo "Binary location: ./build/$(BINARY_NAME_228)"
