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

# This file only exists as a means to run tests in an automated fashion.

FROM arti.dev.cray.com/baseos-docker-master-local/alpine:3.12 AS build-base

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