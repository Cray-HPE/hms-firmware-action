#!/bin/bash -l

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

###############################################################
#
#     CASM Test - Cray Inc.
#
#     TEST IDENTIFIER   : fas_tavern_api_test
#
#     DESCRIPTION       : Automated test for verifying the HMS Firmware
#                         Action Service (FAS) API on Cray Shasta systems.
#                         
#     AUTHOR            : Mitch Schooler
#
#     DATE STARTED      : 07/06/2020
#
#     LAST MODIFIED     : 01/29/2021
#
#     SYNOPSIS
#       This is a test wrapper for HMS Firmware Action Service (FAS)
#       API tests implemented in Tavern that launch via pytest within the
#       Continuous Test (CT) framework. All Tavern tests packaged in
#       the target CT test directory for FAS are executed.
#
#     INPUT SPECIFICATIONS
#       Usage: fas_tavern_api_test
#       
#       Arguments: None
#
#     OUTPUT SPECIFICATIONS
#       Plaintext is printed to stdout and/or stderr. The script exits
#       with a status of '0' on success and '1' on failure.
#
#     DESIGN DESCRIPTION
#       This test wrapper generates a Tavern configuration file based
#       on the target test system it is running on and then executes all
#       FAS Tavern API CT tests using HMS's hms-pytest container which
#       includes pytest and other dependencies required to run Tavern.
#
#     SPECIAL REQUIREMENTS
#       Must be executed on the target test system on a fully-installed
#       NCN with the Continuous Test infrastructure in place.
#
#     UPDATE HISTORY
#       user       date         description
#       -------------------------------------------------------
#       schooler   07/06/2020   initial implementation
#       schooler   09/21/2020   use latest hms_common_file_generator
#       schooler   01/29/2021   set VERIFY=False for running on PIT nodes
#
#     DEPENDENCIES
#       - hms-pytest wrapper script which is expected to be packaged
#         in /usr/bin on the NCN.
#       - hms_smoke_test_lib_ncn-resources_remote-resources.sh which
#         is expected to be packaged in
#         /opt/cray/tests/ncn-resources/hms/hms-test on the NCN.
#       - hms_pytest_ini_file_generator_ncn-resources_remote-resources.py
#         which is expected to be packaged in
#         /opt/cray/tests/ncn-resources/hms/hms-test on the NCN.
#       - hms_common_file_generator_ncn-resources_remote-resources.py
#         which is expected to be packaged in
#         /opt/cray/tests/ncn-resources/hms/hms-test on the NCN.
#       - FAS Tavern API tests with names of the form test_*.tavern.yaml
#         which are expected to be packaged in
#         /opt/cray/tests/functional/hms/hms-firmware-action on the NCN.
#
#     BUGS/LIMITATIONS
#       None
#
###############################################################

# cleanup
function cleanup()
{
    echo "cleaning up..."
    if [[ -d ${FAS_TEST_DIR}/.pytest_cache ]] ; then
        rm -rf ${FAS_TEST_DIR}/.pytest_cache
    fi
    rm -f ${PYTEST_INI_PATH}
    rm -f ${COMMON_FILE_PATH}
}

# HMS path declarations
HMS_TEST_LIB="/opt/cray/tests/ncn-resources/hms/hms-test/hms_smoke_test_lib_ncn-resources_remote-resources.sh"
PYTEST_INI_GENERATOR="/opt/cray/tests/ncn-resources/hms/hms-test/hms_pytest_ini_file_generator_ncn-resources_remote-resources.py"
PYTEST_INI_PATH="/opt/cray/tests/ncn-functional/hms/hms-firmware-action/pytest.ini"
COMMON_FILE_GENERATOR="/opt/cray/tests/ncn-resources/hms/hms-test/hms_common_file_generator_ncn-resources_remote-resources.py"
COMMON_FILE_PATH="/opt/cray/tests/ncn-functional/hms/hms-firmware-action/common.yaml"
FAS_TEST_DIR="/opt/cray/tests/ncn-functional/hms/hms-firmware-action"
API_TARGET="https://api-gw-service-nmn.local/apis"

# set SSL certificate checking to False for test execution from PIT nodes
VERIFY="False"
echo "VERIFY=${VERIFY}"

# set up signal handling
trap ">&2 echo \"recieved kill signal, exiting with status of '1'...\" ; \
    cleanup ; \
    exit 1" SIGHUP SIGINT SIGTERM

# verify that the hms-pytest wrapper script exists
PYTEST_PATH=$(which hms-pytest)
if [[ -z ${PYTEST_PATH} ]] ; then
    >&2 echo "ERROR: failed to locate command: hms-pytest"
    cleanup
    exit 1
fi

# source the HMS smoke test library file
if [[ -r ${HMS_TEST_LIB} ]] ; then
    . ${HMS_TEST_LIB}
else
    >&2 echo "ERROR: failed to source HMS smoke test library: ${HMS_TEST_LIB}"
    cleanup
    exit 1
fi

# verify that the Tavern test directory exists
if [[ ! -d ${FAS_TEST_DIR} ]] ; then
    >&2 echo "ERROR: failed to locate Tavern test directory: ${FAS_TEST_DIR}"
    cleanup
    exit 1
else
    TEST_DIR_FILES=$(ls ${FAS_TEST_DIR})
    TEST_DIR_TAVERN_FILES=$(echo "${TEST_DIR_FILES}" | grep -E "^test_.*\.tavern\.yaml")
    if [[ -z "${TEST_DIR_TAVERN_FILES}" ]] ; then
        >&2 echo "ERROR: no Tavern tests in CT test directory: ${FAS_TEST_DIR}"
        >&2 echo "${TEST_DIR_FILES}"
        cleanup
        exit 1
    fi
fi

# verify that the pytest.ini generator tool exists
if [[ ! -x ${PYTEST_INI_GENERATOR} ]] ; then
    >&2 echo "ERROR: failed to locate executable pytest.ini file generator: ${PYTEST_INI_GENERATOR}"
    cleanup
    exit 1
fi

# verify that the common file generator tool exists
if [[ ! -x ${COMMON_FILE_GENERATOR} ]] ; then
    >&2 echo "ERROR: failed to locate executable common file generator: ${COMMON_FILE_GENERATOR}"
    cleanup
    exit 1
fi

echo "Running fas_tavern_api_test..."

# retrieve Keycloak authentication token for session
TOKEN=$(get_auth_access_token)
TOKEN_RET=$?
if [[ ${TOKEN_RET} -ne 0 ]] ; then
    cleanup
    exit 1
fi

# generate pytest.ini configuration file
GENERATE_PYTEST_INI_CMD="${PYTEST_INI_GENERATOR} --file ${PYTEST_INI_PATH}"
timestamp_print "Running '${GENERATE_PYTEST_INI_CMD}'..."
eval "${GENERATE_PYTEST_INI_CMD}"
GENERATE_PYTEST_INI_RET=$?
if [[ ${GENERATE_PYTEST_INI_RET} -ne 0 ]] ; then
    >&2 echo "ERROR: pytest.ini file generator failed with error code: ${GENERATE_PYTEST_INI_RET}"
    cleanup
    exit 1
else
    if [[ ! -r ${PYTEST_INI_PATH} ]] ; then
        >&2 echo "ERROR: failed to generate readable pytest.ini file"
        cleanup
        exit 1
    fi
fi

# generate Tavern common.yaml configuration file
GENERATE_COMMON_FILE_CMD="${COMMON_FILE_GENERATOR} --base_url ${API_TARGET} --file ${COMMON_FILE_PATH} --access_token ${TOKEN} --verify ${VERIFY}"
timestamp_print "Running '${GENERATE_COMMON_FILE_CMD}'..."
eval "${GENERATE_COMMON_FILE_CMD}"
GENERATE_COMMON_FILE_RET=$?
if [[ ${GENERATE_COMMON_FILE_RET} -ne 0 ]] ; then
    >&2 echo "ERROR: common file generator failed with error code: ${GENERATE_COMMON_FILE_RET}"
    cleanup
    exit 1
else
    if [[ ! -r ${COMMON_FILE_PATH} ]] ; then
        >&2 echo "ERROR: failed to generate readable Tavern common.yaml file"
        cleanup
        exit 1
    fi
fi

# execute Tavern tests in the hms-pytest container with pytest
PYTEST_CMD="${PYTEST_PATH} --tavern-global-cfg=${COMMON_FILE_PATH} ${FAS_TEST_DIR}"
timestamp_print "Running '${PYTEST_CMD}'..."
eval "${PYTEST_CMD}"
TAVERN_RET=$?
if [[ ${TAVERN_RET} -ne 0 ]] ; then
    echo "FAIL: fas_tavern_api_test ran with failures"
    cleanup
    exit 1
else
    echo "PASS: fas_tavern_api_test passed!"
    cleanup
    exit 0
fi
