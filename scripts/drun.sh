#!/bin/bash
sudo apparmor_parser -r -W topography-apparmor-profile.txt

# --security-opt apparmor=topography-apparmor-profile \
docker run \
    --rm \
    -p 8080:8080 \
    --security-opt seccomp=seccomp.json \
    --security-opt apparmor=topography-apparmor-profile \
    topography:slim
