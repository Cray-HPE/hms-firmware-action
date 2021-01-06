# Copyright 2020 Hewlett Packard Enterprise Development LP
# This file only exists as a means to run tests in an automated fashion.

FROM dtr.dev.cray.com/baseos/alpine:3.12 AS build-base

# Configure pip to use the DST PIP Mirror
# PIP Looks for these enviroment variables to configure the PIP mirror
ENV PIP_TRUSTED_HOST dst.us.cray.com
ENV PIP_INDEX_URL http://$PIP_TRUSTED_HOST/dstpiprepo/simple/

ENV LOG_LEVEL TRACE
ENV API_URL "http://firmware-action"
ENV API_SERVER_PORT ":28800"
ENV API_BASE_PATH ""
ENV VERIFY_SSL False

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

WORKDIR src

# PROTIP: python -m pytest test/ is different than pytest test/
# the first one appends some path stuff and python paths are a PITA; so DONT change this!
RUN set -ex \
    && pwd \
    && python3 -m pytest test/

#in case you want to sleep instead of RUN
#CMD ["sh", "-c", "sleep 1000" ]

#build and run
#docker build --rm --no-cache --network hms-firmware-action_rts -f test.Dockerfile .

#build then run-
#docker build -t fas_test -f test.Dockerfile .
#docker run -d --name fas_test --network hms-firmware-action_rts fas_test