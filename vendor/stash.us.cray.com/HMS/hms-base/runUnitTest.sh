#!/usr/bin/env bash

# Build the build base image
docker build -t cray/hms-base-build-base -f Dockerfile.build-base .

docker build -t cray/hms-base-testing -f Dockerfile.testing .
