#!/bin/sh
set -e
IMAGE_NAME="topography"
docker build -t $IMAGE_NAME .
slim --log-level=error --report=off build --tag "${IMAGE_NAME}:slim" --http-probe=false "${IMAGE_NAME}:latest"
