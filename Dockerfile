# MIT License
#
# (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

# Dockerfile for building hms-firmware-action.

# Build base just has the packages installed we need.
FROM arti.dev.cray.com/baseos-docker-master-local/golang:1.14-alpine3.12 AS build-base

RUN set -ex \
    && apk update \
    && apk add build-base

# Base copies in the files we need to test/build.
FROM build-base AS base

# Copy all the necessary files to the image.
COPY cmd $GOPATH/src/stash.us.cray.com/HMS/hms-firmware-action/cmd
COPY vendor $GOPATH/src/stash.us.cray.com/HMS/hms-firmware-action/vendor
COPY internal $GOPATH/src/stash.us.cray.com/HMS/hms-firmware-action/internal
COPY .version $GOPATH/src/stash.us.cray.com/HMS/hms-firmware-action/.version

### Build Stage ###
FROM base AS builder

RUN set -ex && go build -v -i -o /usr/local/bin/hms-firmware-action stash.us.cray.com/HMS/hms-firmware-action/cmd/hms-firmware-action

### Build python base ###

FROM arti.dev.cray.com/baseos-docker-master-local/alpine:3.12 AS deploy-base

COPY cmd/fw-loader/Pipfile /

RUN set -x \
    && apk update \
    && apk add --no-cache \
        bash \
        curl \
        python3 \
        py3-pip \
        rpm \
    && pip3 install --upgrade pip \
    && pip3 install pipenv --ignore-installed \
    && pipenv install --deploy --ignore-pipfile

### Final Stage ###

FROM deploy-base
LABEL maintainer="Cray, Inc."
EXPOSE 28800
STOPSIGNAL SIGTERM

# Get the hms-firmware-action from the builder stage.
COPY --from=builder /usr/local/bin/hms-firmware-action /usr/local/bin/.
COPY configs configs

COPY .version /
COPY fw-loader /
COPY cmd/fw-loader /src

# Setup environment variables.
ENV SMS_SERVER=https://api-gateway.default.svc.cluster.local/apis/smd
ENV VAULT_ENABLED="true"
ENV VAULT_ADDR="http://cray-vault.vault:8200"
ENV VAULT_KEYPATH="secret/hms-creds"

ENV LOG_LEVEL "INFO"
ENV SERVICE_RESERVATION_VERBOSITY "INFO"
ENV TRS_IMPLEMENTATION "LOCAL"
ENV STORAGE "ETCD"
ENV ETCD_HOST "etcd"
ENV ETCD_PORT "2379"
ENV HSMLOCK_ENABELD "true"
ENV NODE_BLACKLIST "ignore_ignore_ignore"

### IGNORE THIS VVVV
# Can also set KV_URL to point to the Key Value service access.
# This can be a reference to an etcd service, or "mem:", in which
# case, this will store everything directly in memory.
# If this is not set, then the service will look for the env vars
# ETCD_HOST and ETCD_PORT, and if those are set, it will construct
# the service name as http://$ETCD_HOST:$ETCD_PORT.
# ENV KV_URL=mem:
### ENDOF IGNORE ^^^

# Get the hms-firmware-action from the builder stage.
COPY --from=builder /usr/local/bin/hms-firmware-action /usr/local/bin/.

# Set up for fw-loader
COPY entrypoint.sh /
COPY wait-for.sh /
COPY .version /
COPY fw-loader /
COPY cmd/fw-loader /src

ENV API_URL "http://cray-fas"
ENV API_SERVER_PORT ":28800"
ENV API_BASE_PATH "/v1"
ENV NUKE_FAS "false"
ENV IGNORE_POPULATED_FAS "true"

#CMD ["sh", "-c", "hms-firmware-action --smsserver=$SMS_SERVER ${KV_URL:+-kv-url=}$KV_URL ${DEPFILE:+-depfile=}$DEPFILE"]
CMD ["sh", "-c", "/entrypoint.sh"]
