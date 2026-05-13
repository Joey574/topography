#!/bin/bash
sudo apparmor_parser -r -W topography-apparmor-profile.txt

docker run \
    --rm \
    --read-only \
    --name topo \
    -p 8080:8080 \
    --cap-drop=ALL \
    --tmpfs /tmp:rw,noexec,nosuid,size=64m \
    --security-opt=no-new-privileges:true \
    --security-opt seccomp=seccomp.json \
    --security-opt apparmor=topography-apparmor-profile \
    topography:slim
