#!/bin/sh

# Usage: GIT_TOKEN=aabbccdd ./build.sh

# Builds the app via docker and places result in ./deploy/bin

docker build --build-arg GIT_TOKEN=$GIT_TOKEN -t rotcore:latest .

# Automatically remove the container after it's built
# Mount bin dir to copy built executables in
docker run --rm=true --mount type=bind,source="$(pwd)/deploy/bin",target=/data rotcore:latest