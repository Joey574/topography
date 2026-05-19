#!/bin/sh
set -e
docker build -t topography .
slim --report=off build \
    --http-probe=false \
    --remove-file-artifacts \
    --tag "topography:slim" \
    --obfuscate-metadata \
    --obfuscate-app-package-names empty \
    "topography:latest"
