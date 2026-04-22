#!/bin/bash
set -e

IMAGE_NAME="topography"

echo "Building the Docker image..."
docker build -t $IMAGE_NAME .

echo -e "\nRunning the container on http://localhost:8080 ..."
# Map port 8080 on your host to port 8080 inside the container
docker run --rm -p 8080:8080 $IMAGE_NAME