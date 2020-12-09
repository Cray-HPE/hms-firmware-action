#!/usr/bin/env python3
# MIT License
#
# (C) Copyright [2021] Hewlett Packard Enterprise Development LP
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

import sys
import os
import json
sys.path.append('fas')  # HAVE TO HAVE THIS; i still dont understand python imports :(
import logging
import time
from fas import FirmwareAction, models

print(sys.path)

# FULL_PATH_TO_DIR = os.path.abspath(os.path.dirname(os.path.realpath(__file__)))
# FULL_PATH_TO_PARENT_DIR = os.path.dirname(FULL_PATH_TO_DIR)
# FULL_PATH_TO_UTIL_DIR = os.path.join(FULL_PATH_TO_PARENT_DIR,"./fas")
#
# try:
#     from fas import action
#     from fas import config
#     from fas import image
#     from fas import snapshot
#     from fas import models
# except:
#     sys.path.append( os.path.abspath(FULL_PATH_TO_UTIL_DIR))
#     from fas import action
#     from fas import config
#     from fas import image
#     from fas import snapshot
#     from fas import models



def main():
    logging.Formatter.converter = time.gmtime
    FORMAT = '%(asctime)-15s-%(levelname)s-%(message)s'
    logging.basicConfig(format=FORMAT, level=logging.DEBUG, datefmt='%Y-%m-%dT%H:%M:%SZ')
    logging.info('STARTING BADGER LOADER')

    # xnames = ["x0c0s1b0","x0c0s2b0"]
    # scp = fas.action.StateComponentFilter(xnames, [], [], [])
    # ihf = fas.action.InventoryHardwareFilter("intel", "x-3")
    # com = fas.action.Command("latest", False, False, 0, "test")
    #
    # actionParameter = fas.action.ActionParameter(scp, ihf, None, None, com)
    #
    #
    # actionResp = fas.action.create_action(actionParameter)
    # logging.debug(actionResp.status_code)
    # actions = fas.action.get_actions().text
    # print(actions)
    # st = fas.action.delete_actions(actions)
    #
    # for s in st:
    #     s
    # print(st)
    # a = fas.action.get_actions().text
    # print(a)

    fasy = FirmwareAction.FirmwareAction("http://localhost", ":28800", "", "DEBUG", False)
    # ### SNAPSHOTS
    # xnames = ["x0c0s1b0", "x0c0s2b0"]
    # scp = models.StateComponentFilter(xnames, [], [], [])
    #
    # snapshotParameter = models.SnapshotParameter("today", scp,  None, None, None)
    #
    #
    #
    # snapshotResp = fasy.create_snapshot(snapshotParameter)
    # logging.debug(snapshotResp.status_code)
    # logging.debug(snapshotResp.text)
    #
    #
    # logging.debug(fasy.get_snapshots().text)

    ### IMAGES
    image = models.Image("NodeBMC", "Cray", "X", "BIOS", "www.s3.com", "vv1.0.0", "1.0.0")

    respImg = fasy.get_images()
    logging.info(json.dumps(json.loads(respImg.text)))
    logging.info(respImg.status_code)

if __name__ == '__main__':
    main()
