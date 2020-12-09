#!/usr/bin/env bash

# Build the build base image
docker build -t cray/hms-hmetcd-build-base -f Dockerfile.build-base .

docker build -t cray/hms-hmetcd-coverage -f Dockerfile.coverage .
