# MIT License
#
# (C) Copyright 2021 Hewlett Packard Enterprise Development LP
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
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

# Service
NAME ?= cray-firmware-action
VERSION ?= $(shell cat .version)

# Helm Chart
CHART_PATH ?= kubernetes
CHART_NAME ?= cray-hms-firmware-action
CHART_VERSION ?= $(shell cat .version)

# Common RPM variable
BUILD_METADATA ?= "1~development~$(shell git rev-parse --short HEAD)"

# CT Test RPM
TEST_SPEC_NAME ?= hms-fas-ct-test
TEST_RPM_VERSION ?= $(shell cat .version)
TEST_SPEC_FILE ?= ${TEST_SPEC_NAME}.spec
TEST_SOURCE_NAME ?= ${TEST_SPEC_NAME}-${TEST_RPM_VERSION}
TEST_BUILD_DIR ?= $(PWD)/dist/fas-ct-test-rpmbuild
TEST_SOURCE_PATH := ${TEST_BUILD_DIR}/SOURCES/${TEST_SOURCE_NAME}.tar.bz2

all: image chart unittest integration test_rpm

image:
	docker build --pull ${DOCKER_ARGS} --tag '${NAME}:${VERSION}' .

chart:
	helm repo add cray-algol60 https://artifactory.algol60.net/artifactory/csm-helm-charts
	helm dep up ${CHART_PATH}/${CHART_NAME}
	helm package ${CHART_PATH}/${CHART_NAME} -d ${CHART_PATH}/.packaged --version ${CHART_VERSION}

unittest:
	./runUnitTest.sh

integration:
	./runIntegration.sh

test_rpm: test_rpm_prepare test_rpm_package_source test_rpm_build_source test_rpm_build

test_rpm_prepare:
	rm -rf $(TEST_BUILD_DIR)
	mkdir -p $(TEST_BUILD_DIR)/SPECS $(TEST_BUILD_DIR)/SOURCES
	cp $(TEST_SPEC_FILE) $(TEST_BUILD_DIR)/SPECS/

test_rpm_package_source:
	tar --transform 'flags=r;s,^,/$(TEST_SOURCE_NAME)/,' --exclude .git --exclude dist -cvjf $(TEST_SOURCE_PATH) ./${TEST_SPEC_FILE} ./test/ct ./LICENSE

test_rpm_build_source:
	BUILD_METADATA=$(BUILD_METADATA) rpmbuild -ts $(TEST_SOURCE_PATH) --define "_topdir $(TEST_BUILD_DIR)"

test_rpm_build:
	BUILD_METADATA=$(BUILD_METADATA) rpmbuild -ba $(TEST_SPEC_FILE) --define "_topdir $(TEST_BUILD_DIR)" --nodeps
