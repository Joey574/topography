#!/bin/bash
docker run \
    --rm \
    --read-only \
    --name topo \
    -p 8080:8080 \
    --cap-drop=ALL \
    --tmpfs /tmp:rw,noexec,nosuid,size=64m \
    --security-opt=no-new-privileges:true \
    topography:slim
