# Copyright 2020 Hewlett Packard Enterprise Development LP
# This file only exists as a means to run tests in an automated fashion.

FROM dtr.dev.cray.com/baseos/alpine:3.12 AS build-base

# Configure pip to use the DST PIP Mirror
# PIP Looks for these enviroment variables to configure the PIP mirror
ENV PIP_TRUSTED_HOST dst.us.cray.com
ENV PIP_INDEX_URL http://$PIP_TRUSTED_HOST/dstpiprepo/simple/

# Only copy the necessary part of Liquibase
COPY Pipfile* /
COPY src src
COPY requirements.txt .

RUN set -ex \
    && apk update \
    && apk add --no-cache \
        bash \
        curl \
        python3 \
        py3-pip \
    && pip3 install --upgrade pip \
    && pip3 install \
        pipenv \
        requests \
        pytest

##if you use CMD, then it will run like a service; we want this to execute the tests and quit
##RUN go test -v ./...
#RUN set -ex \
#    && go test -v -o firmware-action stash.us.cray.com/HMS/hms-firmware-action/internal/domain \
#    && go test -v -o firmware-action stash.us.cray.com/HMS/hms-firmware-action/internal/api \
#    && go test -v -o firmware-action stash.us.cray.com/HMS/hms-firmware-action/internal/model \
#    && go test -v -o firmware-action stash.us.cray.com/HMS/hms-firmware-action/internal/storage \
#    && go test -v -o firmware-action stash.us.cray.com/HMS/hms-firmware-action/internal/hsm

#RUN set -ex \
#    && apk add --no-cache \
#        g++ \
#        python3 \
#        python3-dev \
#    && pip3 install -r requirements.txt

CMD ["sh", "-c", "sleep 1000" ]