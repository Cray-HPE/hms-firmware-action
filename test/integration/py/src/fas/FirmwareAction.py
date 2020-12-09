#!/usr/bin/env python3
# MIT License
#
# (C) Copyright [2019, 2021] Hewlett Packard Enterprise Development LP
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
import logging
import requests
import os
import time
import json
import models

# API_URL="host.docker.internal"

class FirmwareAction:

    def __init__(self, init_API_URL, init_API_SERVER_PORT="", init_API_BASE_PATH="", init_VERIFY_SSL=True,
                 log_level="INFO"):

        if log_level == "DEBUG":
            self.log_level = logging.DEBUG
        elif log_level == "INFO":
            self.log_level = logging.INFO
        elif log_level == "WARNING":
            self.log_level = logging.WARNING
        elif log_level == "ERROR":
            self.log_level = logging.ERROR
        elif log_level == "NOTSET":
            self.log_level = logging.NOTSET
        else:
            self.log_level = logging.INFO

        logging.Formatter.converter = time.gmtime
        FORMAT = '%(asctime)-15s-%(levelname)s-%(message)s'
        logging.basicConfig(format=FORMAT, level=self.log_level, datefmt='%Y-%m-%dT%H:%M:%SZ')
        logging.debug("LOG_LEVEL: %s", log_level)

        self.API_URL = init_API_URL
        self.API_SERVER_PORT = init_API_SERVER_PORT
        self.API_BASE_PATH = init_API_BASE_PATH

        logging.debug("Configuring API connection: %s, %s, %s", self.API_URL, self.API_SERVER_PORT, self.API_BASE_PATH)

        self.VERIFY_SSL = init_VERIFY_SSL
        logging.debug("Configuring requests library ssl connection to use trusted connection: %s", self.VERIFY_SSL)

    def test_connection(self):
        response = self.get_service_status()
        if response.status_code == 200:
            return True
        return False

    def get_actions(self):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/actions"
        payload = ""
        headers = {
            'cache-control': "no-cache",
        }
        logging.debug("get_actions: url: %s", url)
        return requests.request("GET", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def get_action(self, actionID):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/actions/" + actionID
        payload = ""
        headers = {
            'cache-control': "no-cache",
        }
        return requests.request("GET", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def get_operation(self, actionID, operationID):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/actions/" + actionID + "/operations/" + operationID
        payload = ""
        headers = {
            'cache-control': "no-cache",
        }
        return requests.request("GET", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def delete_action(self, actionID):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/actions/" + actionID
        payload = ""
        headers = {
            'Content-Type': "application/json",
            'cache-control': "no-cache",
        }
        return requests.request("DELETE", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def abort_action(self, actionID):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/actions/" + actionID + "/instance"
        payload = ""
        headers = {
            'Content-Type': "application/json",
            'cache-control': "no-cache",
        }
        return requests.request("DELETE", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def delete_actions(self, actions):
        statuses = []
        apps = json.loads(actions)
        appList = apps["actions"]
        for app in appList:
            d = self.delete_action(app["actionID"])
            statuses.append(d)
        return statuses

    def create_action(self, actionParameter):
        if isinstance(actionParameter, models.ActionParameter):
            # DO SOMETHING
            ap = {}
            if actionParameter.stateComponentFilter is not None:
                scp = {}
                if len(actionParameter.stateComponentFilter.xnames) > 0:
                    scp['xnames'] = actionParameter.stateComponentFilter.xnames
                if len(actionParameter.stateComponentFilter.partitions) > 0:
                    scp['partitions'] = actionParameter.stateComponentFilter.partitions
                if len(actionParameter.stateComponentFilter.groups) > 0:
                    scp['groups'] = actionParameter.stateComponentFilter.groups
                if len(actionParameter.stateComponentFilter.deviceTypes) > 0:
                    scp['deviceTypes'] = actionParameter.stateComponentFilter.deviceTypes
                if len(scp) > 0:
                    ap['stateComponentFilter'] = scp
            if actionParameter.inventoryHardwareFilter is not None:
                ihf = {}
                if len(actionParameter.inventoryHardwareFilter.manufacturer) > 0:
                    ihf['manufacturer'] = actionParameter.inventoryHardwareFilter.manufacturer
                if len(actionParameter.inventoryHardwareFilter.model) > 0:
                    ihf['model'] = actionParameter.inventoryHardwareFilter.model
                if len(ihf) > 0:
                    ap['inventoryHardwareFilter'] = ihf
            if actionParameter.imageFilter is not None:
                iff = {}
                if len(actionParameter.imageFilter.imageID) > 0:
                    iff['imageID'] = actionParameter.imageFilter.imageID
                    ap['imageFilter'] = iff
            if actionParameter.targetFilter is not None:
                tf = {}
                if len(actionParameter.targetFilter.targets) > 0:
                    tf['targets'] = actionParameter.targetFilter.targets
                    ap['targetFilter'] = tf
            if actionParameter.command is not None:
                c = {}
                if len(actionParameter.command.version) > 0:
                    c['version'] = actionParameter.command.version
                c['overRideDryrun'] = actionParameter.command.overRideDryrun
                c['restoreNotPossibleOverride'] = actionParameter.command.restoreNotPossibleOverride
                c['timeLimit'] = actionParameter.command.timeLimit
                if len(actionParameter.command.description) > 0:
                    c['description'] = actionParameter.command.description
                if len(c) > 0:
                    ap['command'] = c
            url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/actions"
            payload = json.dumps(ap)
            headers = {
                'Content-Type': "application/json",
                'cache-control': "no-cache",
            }
            print(url)
            print(payload)
            return requests.request("POST", url, data=payload, headers=headers, verify=self.VERIFY_SSL)
        else:
            raise NameError('actionParameter is not ActionParameter')

    def get_images(self):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/images"
        payload = ""
        headers = {
            'cache-control': "no-cache",
        }
        logging.debug(url)
        return requests.request("GET", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def get_image(self, imageID):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/images/" + imageID
        payload = ""
        headers = {
            'cache-control': "no-cache",
        }
        return requests.request("GET", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def delete_image(self, imageID):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/images/" + imageID
        payload = ""
        headers = {
            'Content-Type': "application/json",
            'cache-control': "no-cache",
        }
        return requests.request("DELETE", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def delete_images(self, images):
        statuses = []
        data = json.loads(images)
        images = data["images"]
        for image in images:
            d = self.delete_image(image["imageID"])
            statuses.append(d)
        return statuses

    def create_image(self, image):
        if isinstance(image, models.Image):
            im = {}
            im['deviceType'] = image.deviceType
            im['manufacturer'] = image.manufacturer
            im['models'] = []
            im['models'] += image.models
            im['s3URL'] = image.s3URL
            im['tags'] = []
            im['tags'] += image.tags
            im['target'] = image.target
            im['firmwareVersion'] = image.firmwareVersion
            im['semanticFirmwareVersion'] = image.semanticFirmwareVersion
            im['needReboot'] = image.needReboot
            im['updateURI'] = image.updateURI
            im['versionURI'] = image.versionURI
            im['allowableDeviceStates'] = []
            im['allowableDeviceStates'] += image.allowableDeviceStates

            url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/images"
            logging.debug(url)
            payload = json.dumps(im)
            logging.debug(payload)
            headers = {
                'Content-Type': "application/json",
                'cache-control': "no-cache",
            }
            print(payload)
            return requests.request("POST", url, data=payload, headers=headers, verify=self.VERIFY_SSL)
        else:
            raise NameError('image is not Image')

    def update_image(self, imageID, image):
        if isinstance(image, models.Image):
            im = {}
            im['deviceType'] = image.deviceType
            im['manufacturer'] = image.manufacturer
            im['models'] = []
            im['models'] += image.models
            im['s3URL'] = image.s3URL
            im['tags'] = []
            im['tags'] += image.tags
            im['target'] = image.target
            im['firmwareVersion'] = image.firmwareVersion
            im['semanticFirmwareVersion'] = image.semanticFirmwareVersion
            im['needReboot'] = image.needReboot
            im['updateURI'] = image.updateURI
            im['versionURI'] = image.versionURI
            im['allowableDeviceStates'] = []
            im['allowableDeviceStates'] += image.allowableDeviceStates

            url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/images" + imageID
            logging.debug(url)
            payload = json.dumps(im)
            logging.debug(payload)
            headers = {
                'Content-Type': "application/json",
                'cache-control': "no-cache",
            }
            return requests.request("PUT", url, data=payload, headers=headers, verify=self.VERIFY_SSL)
        else:
            raise NameError('image is not Image')

    def get_snapshots(self):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/snapshots"
        payload = ""
        headers = {
            'cache-control': "no-cache",
        }
        logging.debug(url)
        return requests.request("GET", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def get_snapshot(self, name):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/snapshots/" + name
        payload = ""
        headers = {
            'cache-control': "no-cache",
        }
        return requests.request("GET", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def delete_snapshot(self, name):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/snapshots/" + name
        payload = ""
        headers = {
            'Content-Type': "application/json",
            'cache-control': "no-cache",
        }
        return requests.request("DELETE", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def delete_snapshots(self, snapshots):
        statuses = []
        data = json.loads(snapshots)
        snapshots = data["snapshots"]
        for snapshot in snapshots:
            d = self.delete_snapshot(snapshot["name"])
            statuses.append(d)
        return statuses

    def create_snapshot(self, snapshotParameter):
        if isinstance(snapshotParameter, models.SnapshotParameter):
            # DO SOMETHING
            ap = {}
            if snapshotParameter.stateComponentFilter is not None:
                scp = {}
                if len(snapshotParameter.stateComponentFilter.xnames) > 0:
                    scp['xnames'] = snapshotParameter.stateComponentFilter.xnames
                if len(snapshotParameter.stateComponentFilter.partitions) > 0:
                    scp['partitions'] = snapshotParameter.stateComponentFilter.partitions
                if len(snapshotParameter.stateComponentFilter.groups) > 0:
                    scp['groups'] = snapshotParameter.stateComponentFilter.groups
                if len(snapshotParameter.stateComponentFilter.deviceTypes) > 0:
                    scp['deviceTypes'] = snapshotParameter.stateComponentFilter.deviceTypes
                if len(scp) > 0:
                    ap['stateComponentFilter'] = scp
            if snapshotParameter.inventoryHardwareFilter is not None:
                ihf = {}
                if len(snapshotParameter.inventoryHardwareFilter.manufacturer) > 0:
                    ihf['manufacturer'] = snapshotParameter.inventoryHardwareFilter.manufacturer
                if len(snapshotParameter.inventoryHardwareFilter.model) > 0:
                    ihf['model'] = snapshotParameter.inventoryHardwareFilter.model
                if len(ihf) > 0:
                    ap['inventoryHardwareFilter'] = ihf
            if snapshotParameter.targetFilter is not None:
                tf = {}
                if len(snapshotParameter.targetFilter.targets) > 0:
                    tf['targets'] = snapshotParameter.targetFilter.targets
                    ap['targetFilter'] = tf
            if snapshotParameter.expirationTime is not None:
                if len(snapshotParameter.expirationTime) > 0:
                    ap['expirationTime'] = snapshotParameter.expirationTime
            if snapshotParameter.name is not None:
                if len(snapshotParameter.name) > 0:
                    ap['name'] = snapshotParameter.name
            url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/snapshots"
            logging.debug(url)
            payload = json.dumps(ap)
            logging.debug(payload)
            headers = {
                'Content-Type': "application/json",
                'cache-control': "no-cache",
            }
            return requests.request("POST", url, data=payload, headers=headers, verify=self.VERIFY_SSL)
        else:
            raise NameError('snapshotParameter is not SnapshotParameter')

    def restore_snapshot(self, name, confirm, timeLimit):
        if name is not None:
            timeLimitStr = ""
            if timeLimit is not None:
                timeLimitStr = "&timeLimit=" + str(timeLimit)
            url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/snapshots" + name + "/restore?confirm=" + confirm + timeLimitStr
            logging.debug(url)
            payload = ""
            headers = {
                'Content-Type': "application/json",
                'cache-control': "no-cache",
            }
            return requests.request("POST", url, payload, headers=headers, verify=self.VERIFY_SSL)

    # TODO eventually add support for PUT of snapshot (really device states)

    def get_service_status(self):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/service/status"
        payload = ""
        headers = {
            'cache-control': "no-cache",
        }
        logging.debug(url)
        return requests.request("GET", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def get_service_status_details(self):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/service/status/details"
        payload = ""
        headers = {
            'cache-control': "no-cache",
        }
        logging.debug(url)
        return requests.request("GET", url, data=payload, headers=headers, verify=self.VERIFY_SSL)

    def get_service_version(self):
        url = self.API_URL + self.API_SERVER_PORT + self.API_BASE_PATH + "/service/version"
        payload = ""
        headers = {
            'cache-control': "no-cache",
        }
        logging.debug(url)
        return requests.request("GET", url, data=payload, headers=headers, verify=self.VERIFY_SSL)
