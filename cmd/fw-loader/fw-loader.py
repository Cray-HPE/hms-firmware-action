#!/usr/bin/env python3
#  MIT License
#
#  (C) Copyright [2020-2023] Hewlett Packard Enterprise Development LP
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
fw_loader - Load a fw image into S3, update fas images.

'''

import os
import sys
import logging
import requests
import time
import json
import argparse
import re
import tempfile
import s3
import semver
import uuid
import subprocess
import shutil
import re
from html.parser import HTMLParser
from urllib3.util import parse_url


sys.path.append('fas')  # HAVE TO HAVE THIS; i still dont understand python imports :(
from fas import FirmwareAction, models

fqdn = "api-gw-service-nmn.local"
proto = "http://"
alt_proto = "https://"

protos = [ proto, alt_proto, "mock:", "mem:", "s3:" ]

gw_uri = "/apis"

fas_gw_select = "/fas"
fas_app_uri = ""
fwloc_url = "file://fw/download/"
meta_data_sfx = ".json"

fasStatusPath = "/service/status"
#TODO update https://rgw-vip (hits the HA proxy )
s3_endpoint_def = "http://rgw.local:8080"
s3_endpoint_def = "http://s3:8080"

max_depth = 12

s3client = None

# Simple parser.  Need to isolate files from the
# directory listing that the web server returns.
# This implementation is assuming a directory listing
# with links to directories and files is the provided
# HTML.  Anything else will most likely fail miserably.
class MyHTMLParser(HTMLParser):
    def __init__(self):
        HTMLParser.__init__(self)
        self.directories = []
        self.files = []
        self.current = ""
    def handle_starttag(self, tag, attrs):
        if tag == "a":
            for attr in attrs:
                if attr[0] == "href":
                    self.current = attr[1]
    def handle_data(self, data):
        if self.current == data:
            if data.endswith('/'):
                self.directories.append(data)
            else:
                self.files.append(data)
    def handle_endtag(self, tag):
        self.current = ""

def wait_for_service(url):
    ready = False
    count = 0
    while not ready:
        try:
            rsp = requests.get(url)
            ready = rsp.status_code < 400
        except requests.exceptions.RequestException as e:
            pass
        if not ready:
            if count == 0:
                logging.warning("Service %s not ready, waiting...", url)
            count += 1
            time.sleep(5)
    if count > 0:
        logging.info("Service %s ready, waited %d seconds.", url, count * 5)

# Known device types.  This list could probably trimmed down further
# but leaving it for now
devtypes = [
    "ChassisBMC",
    "NodeBMC",
    "RouterBMC",
]
lc_devtypes = [x.lower() for x in devtypes]

def get_files_nexus():
    filelist = []
    try:
        result = subprocess.run(['/src/nexus_downloader.py'], stdout=subprocess.PIPE)
        logging.info(result.stdout)
        jfilelist = json.loads(result.stdout)
        logging.info(jfilelist)
        for file in jfilelist["files"]:
          filelist.append(file)
    except:
        logging.critical("NEXUS DOWNLOADER ERROR")
    return filelist

def has_proto(u):
    ret = False
    for p in protos:
        if u.startswith(p):
            ret = True;
            break
    return ret

def needs_uri(u):
    offset = u.find("://")
    if offset < 0: offset = 0
    else:          offset += 3
    return u[offset:].find("/") < 0

def build_url(opt, e, select, app_uri):
    if opt is None and e in os.environ:
        opt = os.environ[e]
    if opt is None:
        opt = proto + fqdn + gw_uri + select + app_uri
    else:
        if not has_proto(opt):
            opt = proto + opt
        if needs_uri(opt):
            opt += app_uri
        elif opt.endswith(gw_uri):
            opt += select + app_uri
    return opt

def dev_type_check(devtype):
    lc_dt = devtype.lower()
    if lc_dt in lc_devtypes:
        return devtypes[lc_devtypes.index(lc_dt)]
    return None

def resp_error(resp, url):
    if not resp.ok:
        logging.error("%s: %s", url, resp.text)
        os._exit(1)

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
                logging.error("Loading JSON from %s failed: %s" % (f, e))
                logging.error("Content: %s" % txt)

        return obj
    resp = requests.get(url + f)
    if resp.ok and resp.text.strip() != "":
        try:
            obj = json.loads(resp.text)
        except Exception as e:
            logging.error("Loading JSON from %s failed: %s" % (f, e))
            logging.error("Content: %s" % resp.text)
    return obj

def download_file(url):
    if url.startswith('file:'):
        return open(url[5:], "rb")
    f = tempfile.TemporaryFile()
    logging.debug("Downloading file from %s", url)
    nbytes = 0
    with requests.get(url, stream=True) as r:
        r.raise_for_status()
        for chunk in r.iter_content(chunk_size=8192):
            if chunk:
                nbytes += len(chunk)
                f.write(chunk)
        logging.debug("Downloaded %d byte file", nbytes)
    f.seek(0)
    return f

def check_attrs(obj, attrs, id_list):
    for a in attrs:
        if not a in obj:
            return False
    for id in id_list:
      found = True
      for c in id:
          if c not in obj:
              found = False
      if found == True:
          return found
    return found

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

#req_attrs = ["deviceType", "manufacturer", "models", "tags", "firmwareVersion", "semanticFirmwareVersion"]
req_attrs = ["tags", "firmwareVersion", "semanticFirmwareVersion", "fileName", "targets"]
id_attrs = [["manufacturer", "models", "deviceType"], ["softwareIds"]]
save_objs = []

def process_file(urls, f):
    logging.info("Processing File: %s %s", urls["fwloc"], f)
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
    if not obj is None and check_attrs(obj, req_attrs, id_attrs):
        if obj in save_objs:
            logging.info("Already saw this object: %s", obj)
            return 0
        save_objs.append(obj)
        # We've found a json file that contains at least the minimum
        # required attributes.
        imgs = []
        if "targets" in obj and isinstance(obj["targets"], str):
            obj["targets"] = [ obj["targets"] ]
        if "softwareIds" in obj and isinstance(obj["softwareIds"], str):
            obj["softwareIds"] = [ obj["softwareIds"] ]
        s3_path = ""
        if "targets" in obj:
            for targ in obj["targets"]:
                imgs.append(build_img(obj, s3_path, targ))
        else:
            imgs.append(build_img(obj, s3_path, ""))
        fas_imgs_url = urls["fas"] + "/images"
        logging.debug("FAS URL IMAGES path: %s", fas_imgs_url)
        #logging.info("Images: %s", json.dumps(imgs))
        ret = 0
        for d in imgs:
          if checkforexistingimage(fas_imgs_url, d) == False:
              if (s3_path == ""):
                download_path = urls["fwloc"] + obj["fileName"]
                try:
                  fp = download_file(download_path)
                except FileNotFoundError:
                  logging.error("ERROR: File Not Found: %s", download_path)
                else:
                  temp = parse_url(download_path)
                  s3_path = s3client.upload_image(fp, uuid.uuid1().hex + "/" + obj["fileName"], image_data)
                  fp.close()
              logging.debug("S3 path: %s", s3_path)
              d["s3URL"] = s3_path
              logging.info("IMAGE: %s", json.dumps(d))
              response = requests.post(fas_imgs_url, json=d)
              resp_error(response, fas_imgs_url)
              ret += 1
    else:
        logging.info("Missing required attributes: %s", obj)
    return ret

def checkforexistingimage(fas_imgs_url, img):
    images_resp = requests.get(fas_imgs_url)
    if images_resp.ok:
        logging.debug("Images: %s", images_resp.text)
        images_json = json.loads(images_resp.text)
        images = images_json["images"]
        logging.debug("IMAGES: %s", images)
        logging.debug("IMG: %s", img)
        for image in images:
            if not "deviceType" in image:
                image["deviceType"] = ""
            if not "manufacturer" in image:
                image["manufacturer"] = ""
            if not "target" in image:
                image["target"] = ""
            if not "models" in image:
                image["models"] = []
            if not "softwareIds" in image:
                image["softwareIds"] = []
            logging.debug("DeviceType: %s -- %s", img["deviceType"], image["deviceType"])
            #
            # Samantic Version is in the form x.y.z-b where x,y,z,b are all numbers
            #
            saveSemanticFirmwareVersion = img["semanticFirmwareVersion"]
            try:
                buildSemanticVersion = img["semanticFirmwareVersion"].split("-")
                # Remove leading zeros from version string, because it doesn't work
                img["semanticFirmwareVersion"] = (".".join(str(int(i)) for i in buildSemanticVersion[0].split(".")))
                # make sure version is x.y.z otherwise add .0 to missing values
                for i in range(3-len(buildSemanticVersion[0].split("."))):
                    img["semanticFirmwareVersion"] = img["semanticFirmwareVersion"] + ".0"
                if len(buildSemanticVersion) > 1 and len(buildSemanticVersion[1]) > 0:
                    try:
                        # Remove leading zeros because semver does not like them
                        img["semanticFirmwareVersion"] = img["semanticFirmwareVersion"] + "-" + str(int(buildSemanticVersion[1]))
                    except:
                        # Maybe not an int, so just add string back
                        img["semanticFirmwareVersion"] = img["semanticFirmwareVersion"] + "-" + buildSemanticVersion[1]
            except:
                img["semanticFirmwareVersion"] = saveSemanticFirmwareVersion
                logging.error("ERRORS found in senamticFirmwareVersion - May be invalid: %s", img["semanticFirmwareVersion"])
            logging.debug("SemanticFirmwareVersion: %s -- %s", img["semanticFirmwareVersion"], image["semanticFirmwareVersion"])
            basicFound = False
            tagFound = False
            modelFound = False
            versionFound = False
            if hasattr(semver, 'Version'): # semver version 3.x
                if semver.Version.is_valid(img["semanticFirmwareVersion"]) == True:
                    logging.debug("T SemanticFirmwareVersion: %d", semver.compare(img["semanticFirmwareVersion"], image["semanticFirmwareVersion"]))
                    if semver.compare(img["semanticFirmwareVersion"], image["semanticFirmwareVersion"]) == 0:
                        versionFound = True
            else: # semver version 2.x
                if semver.VersionInfo.isvalid(img["semanticFirmwareVersion"]) == True:
                    logging.debug("T SemanticFirmwareVersion: %d", semver.compare(img["semanticFirmwareVersion"], image["semanticFirmwareVersion"]))
                    if semver.compare(img["semanticFirmwareVersion"], image["semanticFirmwareVersion"]) == 0:
                        versionFound = True
            logging.debug("Target: %s -- %s", img["target"], image["target"])
            if (img["deviceType"].lower() == image["deviceType"].lower()) and (img["target"] == image["target"]) and (img["manufacturer"].lower() == image["manufacturer"].lower()):
                basicFound = True
            for tag in image["tags"]:
                for tag2 in img["tags"]:
                    if tag.lower() == tag2.lower():
                        tagFound = True
            for model in image["models"]:
                for model2 in img["models"]:
                    if model == model2:
                        modelFound = True
            for model in image["softwareIds"]:
                for model2 in img["softwareIds"]:
                    if model == model2:
                        modelFound = True
            if basicFound and versionFound and tagFound and modelFound:
                    logging.info("Image Already found in FAS")
                    return True
    return False

def process_fw(urls, depth=0):
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
            ret += process_file(urls, f)
    # If at the base we find a directory named "FW", then we will assume it has
    # all the firmware and only search it.
    if depth == 0 and "FW/" in dirs:
        dirs = ["FW/"]
    for d in dirs:
        urls["fwloc"] = fwloc_url + d
        ret += process_fw(urls, depth=depth+1)
    return ret

def check_log_level(log_level):
    llevel = logging.INFO
    log_level = log_level.upper()
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
    return log_level, llevel

def check_arg(arg, evar, default_val):
    val = None
    if arg is not None:
        val = arg
    elif evar in os.environ:
        val = os.environ[evar]
    else:
        val = default_val
    return val

def main():
    #TODO https://stackoverflow.com/questions/10551117/setting-options-from-environment-variables-when-using-argparse
    global s3client
    parser = argparse.ArgumentParser()
    # parser.add_argument("--fqdn", default=os.environ.get('FQDN'), help="Fully qualified domain name to use for services")
    # parser.add_argument("--fwloc", default=os.environ.get('FWLOC'), help="URI/URL to FWLOC service")
    # parser.add_argument("--fas", default=os.environ.get('FAS'), help="URI/URL to FAS service")
    # parser.add_argument("--s3-endpoint", default=os.environ.get('S3_ENDPOINT'), help="URL to S3 service")
    # parser.add_argument("--s3-bucket", default=os.environ.get('S3_BUCKET'), help="S3 bucket")
    # parser.add_argument("--s3-access-key", default=os.environ.get('S3_ACCESS_KEY'), help="S3 access key")
    # parser.add_argument("--s3-secret-key", default=os.environ.get('S3_SECRET_KEY'), help="S3 secret key")
    #
    # parser.add_argument("--log-level", default=os.environ.get('LOG_LEVEL'), help="set log level", choices=["debug", "info", "warning", "error"])
    #
    parser.add_argument("--fqdn", help="Fully qualified domain name to use for services")
    parser.add_argument("--fwloc", help="URI/URL to FW Location")
    parser.add_argument("--fas", help="URI/URL to FAS service")
    parser.add_argument("--s3-endpoint", help="URL to S3 service")
    parser.add_argument("--s3-bucket", help="S3 bucket")
    parser.add_argument("--s3-access-key", help="S3 access key")
    parser.add_argument("--s3-secret-key", help="S3 secret key")

    parser.add_argument("--log-level", help="set log level", choices=["debug", "info", "warning", "error"])
    parser.add_argument("--test-run", help="set a test run", type=bool, nargs='?', default=False, const=True)
    parser.add_argument("--local-file", help="set to a local file")
    args = parser.parse_args()

    # CONFIGURE LOGGING
    log_level = "INFO"
    llevel = logging.INFO

    if args.log_level is not None:
        log_level, llevel = check_log_level(args.log_level)
    elif "LOG_LEVEL" in os.environ:
        log_level, llevel = check_log_level(os.environ['LOG_LEVEL'])

    logging.Formatter.converter = time.gmtime
    FORMAT = '%(asctime)-15s-%(name)s-%(levelname)s-%(message)s'
    logging.basicConfig(format=FORMAT, level=llevel,datefmt='%Y-%m-%dT%H:%M:%SZ')
    lgr = logging.getLogger()
    lgr.name = 'FWLoader'

    logging.info('Starting FW Loader, LOG_LEVEL: %s; value: %s', log_level, llevel)

    if args.fqdn is not None: fqdn = args.fqdn
    logging.debug("Command line args: %s", args)

    urls = {
        "fas":build_url(args.fas, "FAS_URL",fas_gw_select, fas_app_uri)
    }
    urls["fwloc"] = check_arg(args.fwloc, "FWLOC_URL", fwloc_url)
    if not (urls["fwloc"].startswith(proto) or urls["fwloc"].startswith(alt_proto) or urls["fwloc"].startswith("file:/")):
        urls["fwloc"] = proto + urls["fwloc"]
    savefwloc = urls["fwloc"]

    s3_endpoint = check_arg(args.s3_endpoint, "S3_ENDPOINT", s3_endpoint_def)
    if not (s3_endpoint.startswith(proto) or s3_endpoint.startswith(alt_proto)):
        s3_endpoint = proto + urls["s3"]
    s3_bucket = check_arg(args.s3_bucket, "S3_BUCKET", "fw-update")
    s3_access_key = check_arg(args.s3_access_key, "S3_ACCESS_KEY", None)
    s3_secret_key = check_arg(args.s3_secret_key, "S3_SECRET_KEY", None)

    try:
        s3client = s3.client(s3_endpoint, s3_access_key, s3_secret_key, s3_bucket)
    except:
        logging.critical("LOADER ERROR: CAN NOT CONNECT TO S3 - EXITING")
        os._exit(5)

    logging.info("urls: %s", urls)

    wait_for_service(urls["fas"] + fasStatusPath)

    numup = 0
    if (args.local_file != None):
        logging.info("Using local file: " + args.local_file)
        nexus_filelist = [args.local_file]
    else:
        if (not args.test_run):
           nexus_filelist = get_files_nexus()
           if nexus_filelist == []:
             logging.critical("NEXUS ERROR")
             os._exit(2)
           logging.info(nexus_filelist)
        else:
           nexus_filelist = ["DOWNLOAD"]
    for file in nexus_filelist:
       urls["fwloc"] = savefwloc
       if (not args.test_run):
           download_path = '/fw/download'
           #Cleanse the path
           shutil.rmtree(download_path, ignore_errors=True)
           while os.path.exists(download_path):
             logging.info("Path Exists after remove - try again")
             time.sleep(5)
             shutil.rmtree(download_path, ignore_errors=True)
           os.mkdir(download_path)
           _, file_ext = os.path.splitext(file)
           if file_ext.lower() == ".rpm":
             logging.info("extracting rpm: " + file)
             rpm2cpio_digester = subprocess.Popen(['rpm2cpio', file], stdout=subprocess.PIPE, cwd=download_path)
             cpio_digester = subprocess.Popen(['cpio', '-idmv'], stdin=rpm2cpio_digester.stdout, cwd=download_path)
             cpio_digester.wait()
           elif file_ext.lower() == ".zip":
             logging.info("unzip: "+ file)
             unzip_digester = subprocess.run(['unzip', file], stdout=subprocess.PIPE, cwd=download_path)
             unzip_digester.wait()
           else:
             logging.error("unsupported file extension: " + file_ext)
       numup += process_fw(urls)
    logging.info("Number of Updates: %d", numup)

    #TODO get a list of all images; then update the urls to be public read

    API_URL = os.getenv('API_URL')
    API_SERVER_PORT = os.getenv('API_SERVER_PORT')
    API_BASE_PATH = os.getenv('API_BASE_PATH')

    fasy = FirmwareAction.FirmwareAction(API_URL, API_SERVER_PORT, API_BASE_PATH, False, "DEBUG")
    respImg = fasy.get_images()
    #logging.info(json.dumps(json.loads(respImg.text)))
    #logging.info(respImg.status_code)
    images = json.loads(respImg.text)
    targeted_images = []
    logging.info("Iterate images")
    for k, v in images.items():
        for v2 in v:
            logging.debug(v2["s3URL"])

            #CASMHMS-4222 ; purge the persist images from FAS
            if (
                "manufacturer" in v2 and v2["manufacturer"].lower() == "cray"
                and "deviceType" in v2 and v2["deviceType"].lower() == "routerbmc"
                and "target" in v2 and v2["target"].lower() == "bmc"
                and "persist" in v2["tags"]
            ):
                logging.info("Deleting persist routerBMC image ")
                logging.info(v2)
                fasy.delete_image(v2["imageID"])
                numup -= 1
            else:
                targeted_images.append(v2["s3URL"])
    for i in targeted_images:
        updated_key = i.replace("s3://", "")
        # TODO fix the URLs showing up as `s3:/` ; that has to be on the loader
        updated_key = updated_key.replace("s3:/", "")
        updated_key = updated_key.replace("fw-update/", "")
        stat = s3client.update_image_acl(updated_key)
        logging.debug(stat)
    logging.info("finished updating images ACL")

    if (args.local_file != None):
        logging.info("removing local file: " + args.local_file)
        os.remove(args.local_file)
    logging.info("*** Number of Updates: %d ***", numup)

    os._exit(0)


if __name__ == '__main__':
    main()
# vim: set expandtab tabstop=4 shiftwidth=4
