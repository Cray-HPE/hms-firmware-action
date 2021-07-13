NAME ?= hms-firmware-action
VERSION ?= $(shell cat .version)

all: image unittest integration

image:
	docker build --pull ${DOCKER_ARGS} --tag '${NAME}:${VERSION}' .

unittest:
	./runUnitTest.sh

integration:
	./runIntegration.sh
