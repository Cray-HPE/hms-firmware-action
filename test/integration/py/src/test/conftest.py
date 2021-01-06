#  MIT License
#
#  (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
#
#  Permission is hereby granted, free of charge, to any person obtaining a
#  copy of this software and associated documentation files (the "Software"),
#  to deal in the Software without restriction, including without limitation
#  the rights to use, copy, modify, merge, publish, distribute, sublicense,
#  and/or sell copies of the Software, and to permit persons to whom the
#  Software is furnished to do so, subject to the following conditions:
#
#  The above copyright notice and this permission notice shall be included
#  in all copies or substantial portions of the Software.
#
#  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
#  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
#  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
#  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
#  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
#  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
#  OTHER DEALINGS IN THE SOFTWARE.

import pytest
import sys
import os

sys.path.append('../fas')
sys.path.append('fas')  # HAVE TO HAVE THIS; i still dont understand python imports :(
import logging
import time
from fas import FirmwareAction, models

print(sys.path)

import logging
import time
from fas import FirmwareAction, models


pytest.FAS = None
def pytest_configure(config):
    llevel = logging.DEBUG
    log_level = "DEBUG"
    if "LOG_LEVEL" in os.environ:
        log_level = os.environ['LOG_LEVEL'].upper()
        if log_level == "DEBUG":
            llevel = logging.DEBUG
        elif log_level == "INFO":
            llevel = logging.INFO
        elif log_level == "WARNING":
            llevel = logging.WARNING
        elif log_level == "ERROR":
            llevel = logging.ERROR
        elif log_level == "NOTSET":
            llevel = logging.NOTSET

    logging.Formatter.converter = time.gmtime
    FORMAT = '%(asctime)-15s-%(levelname)s-%(message)s'
    logging.basicConfig(format=FORMAT, level=llevel,datefmt='%Y-%m-%dT%H:%M:%SZ')
    logging.info('STARTING TESTING CONFIG')
    logging.info("LOG_LEVEL: %s; value: %s", log_level, llevel)

    # CONFIGURE CONNECTION
    if "API_URL" in os.environ:
        api_url = os.environ['API_URL']
    else:
        api_url = "http://localhost"

    if "API_SERVER_PORT" in os.environ:
        api_server_port = os.environ['API_SERVER_PORT']
    else:
        api_server_port = ":28800"

    if "API_BASE_PATH" in os.environ:
        api_base_path = os.environ['API_BASE_PATH']
    else:
        api_base_path = ""

    #have to setup ssl policy before trying to use the api
    verify_ssl = False
    if "VERIFY_SSL" in os.environ:
        if os.environ['VERIFY_SSL'].upper() == 'FALSE':
            verify_ssl = False

    fasy = FirmwareAction.FirmwareAction(api_url, api_server_port, api_base_path, verify_ssl, log_level )

    res = fasy.test_connection()
    if not res:
        logging.error("failed to connect to api")
        assert 0
    else:
        logging.info("connection established")
        pytest.FAS = fasy

