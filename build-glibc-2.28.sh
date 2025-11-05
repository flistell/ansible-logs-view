#!/bin/bash

# Script to build ansible-logs-view-glibc-2.28 binary using podman and copy it to local host

set -e  # Exit immediately if a command exits with a non-zero status

# Define variables
IMAGE_NAME="ansible-logs-builder"
CONTAINER_NAME="ansible-logs-container"
BINARY_NAME="ansible-logs-view-glibc-2.28"
DOCKERFILE="Dockerfile-glibc-2.28"
OUTPUT_DIR="./build"
mkdir "$OUTPUT_DIR" 2>/dev/null

echo "Building the $BINARY_NAME binary using podman..."

# Build the podman image
echo "Building podman image: $IMAGE_NAME"
podman build -t "$IMAGE_NAME" -f "$DOCKERFILE" .

# Create a container from the image
echo "Creating container: $CONTAINER_NAME"
CONTAINER_ID=$(podman create --name "$CONTAINER_NAME" "$IMAGE_NAME")

# Start the container
echo "Starting container..."
podman start "$CONTAINER_ID" > /dev/null

# Copy the binary from the container to the current directory
echo "Copying binary from container to local host..."
podman cp "$CONTAINER_ID:/app/$BINARY_NAME" "$OUTPUT_DIR/"

# Stop and remove the container
echo "Cleaning up container..."
podman stop "$CONTAINER_ID" > /dev/null
podman rm "$CONTAINER_ID" > /dev/null

# Make the binary executable
chmod +x "$OUTPUT_DIR/$BINARY_NAME"

echo "Successfully built and copied $BINARY_NAME to local host"
echo "Binary location: $OUTPUT_DIR/$BINARY_NAME"
