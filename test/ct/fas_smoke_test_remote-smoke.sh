#!/bin/bash -l
#
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
#
###############################################################
#
#     CASM Test - Cray Inc.
#
#     TEST IDENTIFIER   : fas_smoke_test
#
#     DESCRIPTION       : Automated test for verifying basic Firmware Action
#                         Service (FAS) API infrastructure and installation
#                         on Cray Shasta systems.
#                         
#     AUTHOR            : Mitch Schooler
#
#     DATE STARTED      : 09/22/2020
#
#     LAST MODIFIED     : 09/22/2020
#
#     SYNOPSIS
#       This is a smoke test for the HMS FAS API that makes basic HTTP
#       requests using curl to verify that the service's API endpoints
#       respond and function as expected after an installation.
#
#     INPUT SPECIFICATIONS
#       Usage: fas_smoke_test
#       
#       Arguments: None
#
#     OUTPUT SPECIFICATIONS
#       Plaintext is printed to stdout and/or stderr. The script exits
#       with a status of '0' on success and '1' on failure.
#
#     DESIGN DESCRIPTION
#       This smoke test is based on the Shasta health check srv_check.sh
#       script in the CrayTest repository that verifies the basic health of
#       various microservices but instead focuses exclusively on the FAS
#       API. It was implemented to run from the ct-pipelines container off
#       of the NCN of the system under test within the DST group's Continuous
#       Testing (CT) framework as part of the remote-smoke test suite.
#
#     SPECIAL REQUIREMENTS
#       Must be executed from the ct-pipelines container on a remote host
#       (off of the NCNs of the test system) with the Continuous Test
#       infrastructure installed.
#
#     UPDATE HISTORY
#       user       date         description
#       -------------------------------------------------------
#       schooler   09/22/2020   initial implementation
#
#     DEPENDENCIES
#       - hms_smoke_test_lib_ncn-resources_remote-resources.sh which is
#         expected to be packaged in
#         /opt/cray/tests/remote-resources/hms/hms-test in the ct-pipelines
#         container.
#
#     BUGS/LIMITATIONS
#       None
#
###############################################################

# HMS test metrics test cases: 7
# 1. Check cray-fas pod statuses
# 2. GET /service/version API response code
# 3. GET /service/status API response code
# 4. GET /service/status/details API response code
# 5. GET /actions API response code
# 6. GET /images API response code
# 7. GET /snapshots API response code

# initialize test variables
TEST_RUN_TIMESTAMP=$(date +"%Y%m%dT%H%M%S")
TEST_RUN_SEED=${RANDOM}
OUTPUT_FILES_PATH="/tmp/fas_smoke_test_out-${TEST_RUN_TIMESTAMP}.${TEST_RUN_SEED}"
SMOKE_TEST_LIB="/opt/cray/tests/remote-resources/hms/hms-test/hms_smoke_test_lib_ncn-resources_remote-resources.sh"
CURL_ARGS="-k -i -s -S"
MAIN_ERRORS=0
CURL_COUNT=0

# cleanup
function cleanup()
{
    echo "cleaning up..."
    rm -f ${OUTPUT_FILES_PATH}*
}

# main
function main()
{
    AUTH_ARG="-H \"Authorization: Bearer $TOKEN\""

    # GET tests
    for URL_ARGS in \
        "apis/fas/v1/service/version" \
        "apis/fas/v1/service/status" \
        "apis/fas/v1/service/status/details" \
        "apis/fas/v1/actions" \
        "apis/fas/v1/images" \
        "apis/fas/v1/snapshots"
    do
        URL=$(url "${URL_ARGS}")
        URL_RET=$?
        if [[ ${URL_RET} -ne 0 ]] ; then
            cleanup
            exit 1
        fi
        run_curl "GET ${AUTH_ARG} ${URL}"
        if [[ $? -ne 0 ]] ; then
            ((MAIN_ERRORS++))
        fi
    done

    echo "MAIN_ERRORS=${MAIN_ERRORS}"
    return ${MAIN_ERRORS}
}

# check_pod_status
function check_pod_status()
{
    run_check_pod_status "cray-fas"
    return $?
}

# TARGET_SYSTEM is expected to be set in the ct-pipelines container
if [[ -z ${TARGET_SYSTEM} ]] ; then
    >&2 echo "ERROR: TARGET_SYSTEM environment variable is not set"
    cleanup
    exit 1
else
    echo "TARGET_SYSTEM=${TARGET_SYSTEM}"
    TARGET="auth.${TARGET_SYSTEM}"
    echo "TARGET=${TARGET}"
fi

# TOKEN is expected to be set in the ct-pipelines container
if [[ -z ${TOKEN} ]] ; then
    >&2 echo "ERROR: TOKEN environment variable is not set"
    cleanup
    exit 1
else
    echo "TOKEN=${TOKEN}"
fi

trap ">&2 echo \"recieved kill signal, exiting with status of '1'...\" ; \
    cleanup ; \
    exit 1" SIGHUP SIGINT SIGTERM

# source HMS smoke test library file
if [[ -r ${SMOKE_TEST_LIB} ]] ; then
    . ${SMOKE_TEST_LIB}
else
    >&2 echo "ERROR: failed to source HMS smoke test library: ${SMOKE_TEST_LIB}"
    exit 1
fi

# make sure filesystem is writable for output files
touch ${OUTPUT_FILES_PATH}
if [[ $? -ne 0 ]] ; then
    >&2 echo "ERROR: output file location not writable: ${OUTPUT_FILES_PATH}"
    cleanup
    exit 1
fi

echo "Running fas_smoke_test..."

# run initial pod status test
check_pod_status
if [[ $? -ne 0 ]] ; then
    echo "FAIL: fas_smoke_test ran with failures"
    cleanup
    exit 1
fi

# run main API tests
main
if [[ $? -ne 0 ]] ; then
    echo "FAIL: fas_smoke_test ran with failures"
    cleanup
    exit 1
else
    echo "PASS: fas_smoke_test passed!"
    cleanup
    exit 0
fi
