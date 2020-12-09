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
