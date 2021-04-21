#! /usr/bin/env python3

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

import requests
import sys
import os
import semver
import pathlib
import hashlib
import logging
import time
import xmltodict
import gzip
import json

# Constants
DOWNLOAD_CHUNK_SIZE=4094

# Environment variables (with example value)
# NEXUS_ENDPOINT  - http://host.docker.internal:8081
# NEXUS_REPO - shasta-firmware
# ASSETS_DIR - /firmware

NEXUS_ENDPOINT=""
NEXUS_REPO=""
ASSETS_DIR="/firmware"
files = []

def configure_logging():
    # CONFIGURE LOGGING
    if "LOG_LEVEL" in os.environ:
        log_level = os.environ['LOG_LEVEL'].upper()
        llevel = logging.INFO
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
    else:
        log_level = "DEBUG"
        llevel = logging.DEBUG

    logging.Formatter.converter = time.gmtime
    FORMAT = '%(asctime)-15s-%(levelname)s: %(message)s'
    logging.basicConfig(format=FORMAT, level=llevel,datefmt='%Y-%m-%dT%H:%M:%SZ')
    logging.info("LOG_LEVEL: %s; value: %s", log_level, llevel)

def download_asset(asset, downloadDirectory):
    url = asset["downloadUrl"]
    fileName = pathlib.Path(url).name
    destintationPath = pathlib.Path(downloadDirectory).joinpath(fileName)
    logging.info("Dest Path: %s", destintationPath)

    # Download asset
    response = requests.get(url, stream=True, verify=False)
    with open(destintationPath, 'wb') as file:
        for chunk in response.iter_content(chunk_size=DOWNLOAD_CHUNK_SIZE):
            if chunk:
                file.write(chunk)

    # Verify checksum
    assetCheckSum = asset["checksum"]["sha256"]

    hash_sha1 = hashlib.sha256()
    with open(destintationPath, 'rb') as file:
        while True:
            chunk = file.read(DOWNLOAD_CHUNK_SIZE)
            if not chunk:
                break
            hash_sha1.update(chunk)
    downloadChecksum = hash_sha1.hexdigest()

    logging.info("Asset Digest: %s", assetCheckSum)
    logging.info("Downloaded Digest: %s", downloadChecksum)
    if assetCheckSum != downloadChecksum:
        raise RuntimeError("Invalid checksum")
        return
    files.append(str(destintationPath))
    #logging.info("rpm2cpio "+ str(destintationPath) +" | cpio -idmv")
    #os.system('rpm2cpio '+ str(destintationPath) +' | cpio -idmv')

def get_repo_artifacts(nexus_endpoint, repo):
    repomd_url = nexus_endpoint+"/repository/"+repo+"/repodata/repomd.xml"
    logging.info("Repomd URL: %s", repomd_url)
    response = requests.get(repomd_url, verify=False)
    if response.status_code != 200:
        logging.critical("Failed to get repomd.xml from repo")
        exit(1)

    href = None
    x = xmltodict.parse(response.text)
    data = x["repomd"]["data"]
    for element in data:
        if element["@type"] == 'primary':
            href = element['location']['@href']

    packages_url = nexus_endpoint+"/repository/"+repo+"/"+href
    logging.info("Packages URL: %s", packages_url)

    packages = None
    response = requests.get(packages_url, stream=True, verify=False)
    if response.status_code != 200:
        logging.critical("Failed to get package listing from repo")
        exit(1)

    with gzip.GzipFile(fileobj=response.raw) as f:
        x = xmltodict.parse(f.read())
        packages = x["metadata"]["package"]
        metadata = x["metadata"]
        count = int(metadata["@packages"])

    #artifacts[name][version]assets[]
    artifacts = {}


    if count == 0:
        return
    elif count == 1:
        name = packages["name"]
        version = packages["version"]["@ver"]
        arch = packages["arch"]
        location = packages["location"]['@href']
        checksum = {}
        checksum["sha256"] = packages["checksum"]["#text"]
        downloadUrl = nexus_endpoint+"/repository/"+repo+"/"+location
        logging.debug("Package: %s, %s, %s, %s", name, version, arch, downloadUrl)

        # Initialize the artifact dictionary if it does not exist
        if name not in artifacts:
            artifacts[name] = {}

        asset = {
            'arch': arch,
            "location": location,
            "downloadUrl": downloadUrl,
            "checksum": checksum
        }

        artifacts[name][version] = asset
    else :
        for package in packages:
            name = package["name"]
            version = package["version"]["@ver"]
            arch = package["arch"]
            location = package["location"]['@href']
            checksum = {}
            checksum["sha256"] = package["checksum"]["#text"]
            downloadUrl = nexus_endpoint+"/repository/"+repo+"/"+location
            logging.debug("Package: %s, %s, %s, %s", name, version, arch, downloadUrl)

            # Initialize the artifact dictionary if it does not exist
            if name not in artifacts:
                artifacts[name] = {}

            asset = {
                'arch': arch,
                "location": location,
                "downloadUrl": downloadUrl,
                "checksum": checksum
            }

            artifacts[name][version] = asset
    return artifacts

def main():
    # Setup logging
    configure_logging()

    # Configuration from the environment
    if "NEXUS_ENDPOINT" in os.environ:
        NEXUS_ENDPOINT = os.environ['NEXUS_ENDPOINT']
        logging.info("NEXUS_ENDPOINT: %s", NEXUS_ENDPOINT)
    else:
        logging.critical("NEXUS_ENDPOINT environment variable was not set")
        exit(1)
    if "NEXUS_REPO" in os.environ:
        NEXUS_REPO = os.environ['NEXUS_REPO']
        logging.info("NEXUS_REPO: %s", NEXUS_REPO)
    else:
        logging.critical("NEXUS_REPO environment variable was not set")
        exit(1)
# ASSETS_DIR is hardcoded variable
#    if "ASSETS_DIR" in os.environ:
#        ASSETS_DIR = os.environ['ASSETS_DIR']
#        logging.info("ASSETS_DIR: %s", ASSETS_DIR)
#    else:
#        logging.critical("ASSETS_DIR environment variable was not set")
#        exit(1)

    # Get a listing of artifacts within the badger REPO
    #artifacts[name][version]assets[]
    try:
        artifacts = get_repo_artifacts(NEXUS_ENDPOINT, NEXUS_REPO)

        # Determine what artifact version and asset to download
        assets_to_download = []
        for name, versions in artifacts.items():
            logging.info(name)

            for version, asset in versions.items():
                logging.info("  ├─ %s", version)

                downloadUrl = asset["downloadUrl"]
                logging.info("    ├─ %s", downloadUrl)
                assets_to_download.append(versions[version])

        os.system("rm -rf " + ASSETS_DIR)
        # Download Assets
        if not os.path.exists(ASSETS_DIR):
            os.mkdir(ASSETS_DIR)
        elif any(pathlib.Path(ASSETS_DIR).iterdir()):
            logging.critical("ASSET_DIR is not empty")
            exit(1)

        for asset in assets_to_download:
            download_asset(asset, ASSETS_DIR)
            logging.info(json.dumps({"files": files}))
        logging.info(json.dumps({"files": files}))
    except:
        e = sys.exc_info()[0]
        logging.critical("NEXUS ERROR unable to get artifacts -- %s", e)

if __name__ == '__main__':
    main()
