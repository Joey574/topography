#!/bin/sh
slim build \
  --target topography:latest \
  --tag topography:slim \
  --expose 8080 \
  --publish-port 8080:8080 \
  --http-probe-ports 8080 \
  --show-clogs \
  --copy-meta-artifacts .