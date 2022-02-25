# MIT License
#
# (C) Copyright [2021-2022] Hewlett Packard Enterprise Development LP
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

all: image unittest integration  snyk ct_image ct

image:
	docker build --pull ${DOCKER_ARGS} --tag '${NAME}:${VERSION}' .
unittest:
	./runUnitTest.sh

integration:
	./runIntegration.sh

snyk:
	./runSnyk.sh

ct:
	./runCT.sh

ct_image:
	docker build --no-cache -f test/ct/Dockerfile test/ct/ --tag hms-firmware-action-ct:${VERSION}

	###########
	#1st way to do this.  works!
	# docker-compose -f docker-compose.test.ct.yaml build --no-cache
    #docker-compose -p tavern -f docker-compose.test.ct.yaml up -d cray-fas
    # make ct_image
	#docker run --rm -i -t --network tavern_fas --user root --entrypoint /bin/bash  hms-firmware-action-ct:1.18.0
		#pytest -vvvv --tavern-global-cfg=/src/utils/tavern_global_config.yaml /src/app
	#docker run --rm -i -t --network tavern_fas hms-firmware-action-ct:1.18.0 functional


	#docker-compose -f docker-compose.test.ct.yaml -p tavern  down


	######### second way
	#docker-compose -f docker-compose.test.ct.yaml up --exit-code-from ct-tests-functional ct-tests-functional #this fails
	#docker-compose -f docker-compose.test.ct.yaml -p hms-firmware-action  down
	# This has no delay, and fails like: file "/usr/lib/python3.9/site-packages/requests/adapters.py", line 516, in send
                                         	#ct-tests-functional_1   |     raise ConnectionError(e, request=request)
                                         	#ct-tests-functional_1   | requests.exceptions.ConnectionError: HTTPConnectionPool(host='cray-smd', port=27779): Max retries exceeded with url: /hsm/v2doc/State/Components?type=NodeBMC (Caused by NewConnectionError('<urllib3.connection.HTTPConnection object at 0x7f215147fe20>: Failed to establish a new connection: [Errno 111] Connection refused'))
                                         	#ct-tests-functional_1   | =========================== short test summary info ============================
                                         	#ct-tests-functional_1   | FAILED test_actions.tavern.yaml::Verify the actions resource
                                         	#ct-tests-functional_1   | ========================= 1 failed, 5 passed in 14.76s =========================
                                         	#ct-tests-functional_1   | ERROR:root:FAIL


	######## third way
	#./runCT.sh #this fails b/c of xname?! FAILS  requests.exceptions.ConnectionError: HTTPConnectionPool(host='cray-smd', port=27779): Max retries exceeded with url: /hsm/v2/State/Components?type=NodeBMC (Caused by NewConnectionError('<urllib3.connection.HTTPConnection object at 0x7f396b4044c0>: Failed to establish a new connection: [Errno 111] Connection refused'))
	# when I add a sleep 10 + reorder the tests, then I get:
	# ct-tests-functional_1   | Errors:
      	#ct-tests-functional_1   | E   tavern.util.exceptions.TestFailError: Test 'Ensure that the BMC firmware can be updated with a FAS action' failed:
      	#ct-tests-functional_1   |     - Status code was 400, expected 202:
      	#ct-tests-functional_1   |         {"type": "about:blank", "detail": "invalid/duplicate xnames: [None]", "status": 400, "title": "Bad Request"}
      	#ct-tests-functional_1   | ------------------------------ Captured log call -------------------------------
      	#ct-tests-functional_1   | WARNING  tavern.util.dict_util:dict_util.py:54 Formatting 'xname' will result in it being coerced to a string (it is a <class 'NoneType'>)
      	#ct-tests-functional_1   | ERROR    tavern.response.base:base.py:43 Status code was 400, expected 202:
      	#ct-tests-functional_1   |     {"type": "about:blank", "detail": "invalid/duplicate xnames: [None]", "status": 400, "title": "Bad Request"}

	#does:

		#docker-compose build --no-cache
		#docker-compose up  -d cray-fas #this will stand up everything except for the integration test container
		#docker-compose up --exit-code-from ct-tests-functional ct-tests-functional

