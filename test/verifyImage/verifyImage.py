#!/usr/bin/env python3
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

'''
verifyImage - checks firmware rpm or zip file for completness.

'''

import os
import sys
import logging
import requests
import json
import argparse
import re

download_loc = "./fas_verify_download"
max_depth = 12
meta_data_sfx = ".json"

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
WHITE='\033[1;37m'
BROWN='\033[1;33m'

#sys.path.append('fas')  # HAVE TO HAVE THIS; i still dont understand python imports :(
#from fas import FirmwareAction, models

def get_json(url, f):
    obj = None
    if url.startswith("file:"):
        fp = open(url[5:] + f)
        txt = fp.read()
        if txt.strip() != "":
            try:
                #remove trailing , if present in json
                txt = re.sub(r",\s}", "}", txt)
                obj = json.loads(txt)
            except Exception as e:
                print("Loading JSON from %s failed: %s" % (f, e))
                print("Content: %s" % txt)

        return obj
    resp = requests.get(url + f)
    if resp.ok and resp.text.strip() != "":
        try:
            obj = json.loads(resp.text)
        except Exception as e:
            print("Loading JSON from %s failed: %s" % (f, e))
            print("Content: %s" % resp.text)
    return obj

def check_attrs(file, obj, attrs, id_list, id_values, opt_list, opt_values):
    good = True
    print()
    print(BLUE + "CHECKING FILE: " + file + WHITE)
    print("------------- REQUIRED ITEMS -------------")
    for a in attrs:
        if not a in obj:
            print(RED + "*** ERROR: REQUIRED ITEM: " + a + " ***" + WHITE)
            good = False
    for a in attrs:
        if a in obj:
            if isinstance(obj[a], list):
                print(GREEN + "Required item " + a + " found: [ " + ','.join(map(str,obj[a])) + " ]" + WHITE)
            elif isinstance(obj[a], int):
                print(GREEN + "Required item " + a + " found: " + str(obj[a]) + WHITE)
            else:
                print(GREEN + "Required item " + a + " found: " + obj[a] + WHITE)
    print("------------- OPTIONAL ITEMS -------------")
    for a in opt_list:
        if a in obj:
            if isinstance(obj[a], list):
                print(GREEN + "Optional item " + a + " found: [ " + ','.join(map(str,obj[a])) + " ]" + WHITE)
            elif isinstance(obj[a], int):
                print(GREEN + "Optional item " + a + " found: " + str(obj[a]) + WHITE)
            else:
                print(GREEN + "Optional item " + a + " found: " + obj[a] + WHITE)
    i = 0
    for a in opt_list:
        if not a in obj:
            print(BROWN + "Optional item " + a + " NOT found, defaulting to: " + opt_values[i] + WHITE)
        i = i + 1
    print("------------- MANUFACTUER/MODELS/DEVICETYPE/SOFTWAREIDS -------------")
    i = 0
    for id in id_list:
      found = True
      for a in id:
          if a not in obj:
              print(BROWN + "WARNING: MISSING ITEM: " + a + " : ", id_values[i] + WHITE)
              found = False
          else:
              if isinstance(obj[a], list):
                  print(GREEN + "Item " + a + " found: [ " + ','.join(map(str,obj[a])) + " ]" + WHITE)
              elif isinstance(obj[a], int):
                print(GREEN + "Item " + a + " found: " + str(obj[a]) + WHITE)
              else:
                  print(GREEN + "Item " + a + " found: " + obj[a] + WHITE)
          i = i + 1
    # Checking softwareId for cray
    if "manufacturer" in obj:
        if obj["manufacturer"].lower() == "cray":
            if "softwareIds" not in obj:
                print (RED + "*** ERROR: MISSING softwareIds and manufacturer is cray ***" + WHITE)
                good = False
    else:
        if "softwareIds" not in obj:
            print (RED + "*** ERROR: MISSING softwareIds and manufacturer is missing ***" + WHITE)
            good = False

    print("------------- EXTRA ITEMS -------------")
    for c in obj:
        if c not in attrs and c not in opt_list and c not in id_list[0] and c not in id_list[1] and c not in req_attrs_eq:
            print (BROWN + "*** WARNING: " + c + " extra item found" + WHITE)

    print()
    print("----------------------")
    if good:
        print(GREEN + "******** FILE GOOD ********" + WHITE)
    else:
        print(RED + "****** FILE HAS ERRORS *******" + WHITE)
    print("----------------------")
    print()
    return good

def build_img(obj, s3_path, targ):
    fw = { "s3URL":s3_path, "target":targ }
    fw["target"] = targ
    if "deviceType" in obj: fw["deviceType"] = obj["deviceType"]
    else: fw["deviceType"] = ""
    if "manufacturer" in obj: fw["manufacturer"] = obj["manufacturer"]
    else: fw["manufacturer"] = ""
    if "models" in obj: fw["models"] = obj["models"]
    else: fw["models"] = []
    if "softwareIds" in obj: fw["softwareIds"] = obj["softwareIds"]
    else: fw["softwareIds"] = []
    if "tags" in obj: fw["tags"] = obj["tags"]
    if "firmwareVersion" in obj: fw["firmwareVersion"] = obj["firmwareVersion"]
    if "semanticFirmwareVersion" in obj: fw["semanticFirmwareVersion"] = obj["semanticFirmwareVersion"]
    if "allowableDeviceStates" in obj: fw["allowableDeviceStates"] = obj["allowableDeviceStates"]
    else: fw["allowableDeviceStates"] = []
    if "needManualReboot" in obj: fw["needManualReboot"] = obj["needManualReboot"]
    else: fw["needManualReboot"] = False
    if "pollingSpeedSeconds" in obj: fw["pollingSpeedSeconds"] = obj["pollingSpeedSeconds"]
    else: fw["pollingSpeedSeconds"] = 30
    if "updateURI" in obj: fw["updateURI"] = obj["updateURI"]
    if "versionURI" in obj: fw["versionURI"] = obj["versionURI"]
    if "waitTimeBeforeManualRebootSeconds" in obj: fw["waitTimeBeforeManualRebootSeconds"] = obj["waitTimeBeforeManualRebootSeconds"]
    if "waitTimeAfterRebootSeconds" in obj: fw["waitTimeAfterRebootSeconds"] = obj["waitTimeAfterRebootSeconds"]
    if "forceResetType" in obj: fw["forceResetType"] = obj["forceResetType"]
    return fw

req_attrs = ["tags", "firmwareVersion", "semanticFirmwareVersion", "fileName", "targets"]
req_attrs_eq = ["filename", "target", "softwareIDs", "softwareId", "softwareID"]
id_attrs = [["manufacturer", "models", "deviceType"], ["softwareIds"]]
id_attrs_values = ["not required if softwareIds present", "not required if softwareIds present", "not required if softwareIds present", "only required if model is cray"]
opt_attrs = ["allowableDeviceStates", "needManualReboot", "pollingSpeedSeconds", "updateURI", "versionURI", "waitTimeBeforeManualRebootSeconds", "waitTimeAfterRebootSeconds", "forceResetType"]
opt_attrs_values = ["[]", "False", "30", "-", "-", "-", "-","-"]
save_objs = []

def process_file(urls, f, s3):
    logging.info("Processing File: %s", f)
    obj = get_json(urls["fwloc"], f)
    if obj == None:
      logging.error("CANNOT PROCESS FILE %s", f)
      return 0
    image_data = obj
    logging.debug('Object: %s', obj)
    ret = 0
    if "filename" in obj: obj["fileName"] = obj["filename"]
    if "target" in obj: obj["targets"] = obj["target"]
    if "softwareIDs" in obj: obj["softwareIds"] = obj["softwareIDs"]
    if "softwareId" in obj: obj["softwareIds"] = obj["softwareId"]
    if "softwareID" in obj: obj["softwareIds"] = obj["softwareID"]
    if not obj is None and check_attrs(f, obj, req_attrs, id_attrs, id_attrs_values, opt_attrs, opt_attrs_values):
        save_objs.append(obj)
        # We've found a json file that contains at least the minimum
        # required attributes.
        imgs = []
        if "targets" in obj and isinstance(obj["targets"], str):
            obj["targets"] = [ obj["targets"] ]
        if "softwareIds" in obj and isinstance(obj["softwareIds"], str):
            obj["softwareIds"] = [ obj["softwareIds"] ]
        s3_path = s3
        if "targets" in obj:
            for targ in obj["targets"]:
                imgs.append(build_img(obj, s3_path, targ))
        else:
            imgs.append(build_img(obj, s3_path, ""))

        print("------------- FILE CHECK -------------")
        if "fileName" in obj:
            filepath = urls["fwloc"][5:] + obj["fileName"]
            if not os.path.exists(filepath):
                print(RED + "*** ERROR: FILE NOT FOUND: " + filepath + WHITE)
                good = False
            else:
                print(GREEN + "*** FILE FOUND :" + filepath + WHITE)
        else:
            print(RED + "*** ERROR: FILENAME NOT FOUND IN FILE" + WHITE)
        for img in imgs:
            print("------------- FAS JSON IMAGE -------------")
            print(json.dumps(img))
            print("----------------------")
        sys.stdout.flush()
        ret = 1
    else:
        print("----------------------")
        print(RED + "*** ERROR: Missing required attributes: " + json.dumps(obj) + WHITE)
        print("----------------------")
    return ret

def get_file_list(fwloc_url):
    logging.info("get_file_list(%s)", fwloc_url)
    if fwloc_url.startswith("file:"):
        files = []
        dirs = []
        for d in os.scandir(fwloc_url[5:]):
            if os.path.isdir(d.path):
                dirs.append(d.name)
            elif os.path.isfile(d.path):
                files.append(d.name)
        return (files, dirs)
    try:
        resp = requests.get(fwloc_url)
    except requests.exceptions.RequestException as e:
        raise SystemExit(e)
    if resp.ok:
        p = MyHTMLParser()
        p.feed(resp.text)
        return (p.files, p.directories)


def process_fw(urls, s3, depth=0):
    ret = 0
    fwloc_url = urls["fwloc"]
    logging.info("Processing files from %s", fwloc_url)
    if depth > max_depth:
        logging.info("Hit max directory depth: %d", depth)
        return 0
    if not fwloc_url.endswith("/"): fwloc_url += "/"
    files, dirs = get_file_list(fwloc_url)
    logging.debug("Files: %s", files)
    logging.debug("Directories: %s", dirs)
    urls["fwloc"] = fwloc_url
    for f in files:
        if f.endswith(meta_data_sfx):
            ret += process_file(urls, f, s3)
    # If at the base we find a directory named "FW", then we will assume it has
    # all the firmware and only search it.
    if depth == 0 and "FW/" in dirs:
        dirs = ["FW/"]
    for d in dirs:
        urls["fwloc"] = fwloc_url + d
        ret += process_fw(urls, s3, depth=depth+1)
    return ret

def main():
    global GREEN
    global RED
    global BLUE
    global WHITE
    global BROWN
    parser = argparse.ArgumentParser()
    parser.add_argument("-f", help="File to process", default="")
    parser.add_argument("-s3", help="s3URL", default="")
    parser.add_argument("--nocolors", help="no colors", action="store_true")
    args = parser.parse_args()
    if args.nocolors:
        print("NOCOLORS")
        GREEN=""
        RED=""
        BLUE=""
        WHITE=""
        BROWN=""
    if args.f == "":
        print("ERROR: Missing file name")
        os._exit(1)
    if not os.path.exists(args.f):
        print("ERROR: File not found " + args.f)
        os._exit(1)
    file_name = os.path.abspath(args.f)
    print(file_name)
    _, file_ext = os.path.splitext(file_name)
    urls = {}
    urls["fwloc"] = "file:" + download_loc
    numup = 0
    if file_ext.lower() == ".rpm":
        print('rm -rf '+ download_loc + '; mkdir '+ download_loc +'; cd '+ download_loc +'; rpm2cpio '+ file_name +' | cpio -idmv')
        os.system('rm -rf ' + download_loc + '; mkdir '+ download_loc +'; cd '+ download_loc +'; rpm2cpio '+ file_name +' | cpio -idmv')
        numup += process_fw(urls, args.s3)
    elif file_ext.lower() == ".zip":
        print('rm -rf ' + download_loc + '; mkdir '+ download_loc +'; cd '+ download_loc +'; unzip '+ file_name)
        os.system('rm -rf ' + download_loc + '; mkdir '+ download_loc +'; cd '+ download_loc +'; unzip '+ file_name)
        numup += process_fw(urls, args.s3)
    elif file_ext.lower() == ".json":
       urls["fwloc"] = "file:"
       process_file(urls, args.f, args.s3)
    else:
       print("ERROR: File extions not supported "+ file_ext)
    os._exit(0)

if __name__ == '__main__':
    main()
