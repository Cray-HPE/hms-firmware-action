#!/usr/bin/env python3
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
import datetime


class StateComponentFilter:

    def __init__(self, xnames=[], partitions=[], groups=[], deviceTypes=[]):
        self.xnames = xnames
        self.partitions = partitions
        self.groups = groups
        self.deviceTypes = deviceTypes


class InventoryHardwareFilter:

    def __init__(self, manufacturer="", model=""):
        self.manufacturer = manufacturer
        self.model = model


class ImageFilter:

    def __init__(self, imageID=""):
        self.imageID = imageID


class TargetFilter:

    def __init__(self, targets=[]):
        self.targets = targets


class Command:

    def __init__(self, version="", overrideDryrun=False, restoreNotPossibleOverride=False, timeLimit=0, description="", tag=""):
        self.version = version
        self.overRideDryrun = overrideDryrun
        self.restoreNotPossibleOverride = restoreNotPossibleOverride
        self.timeLimit = timeLimit
        self.description = description
        self.tag = tag

class ActionParameter:

    def __init__(self, stateComponentFilter, inventoryHardwareFilter, targetFilter, imageFilter, command):
        self.stateComponentFilter = None
        self.inventoryHardwareFilter = None
        self.targetFilter = None
        self.imageFilter = None
        self.command = None
        if isinstance(stateComponentFilter, StateComponentFilter):
            self.stateComponentFilter = stateComponentFilter

        if isinstance(inventoryHardwareFilter, InventoryHardwareFilter):
            self.inventoryHardwareFilter = inventoryHardwareFilter

        if isinstance(targetFilter, TargetFilter):
            self.targetFilter = targetFilter

        if isinstance(imageFilter, ImageFilter):
            self.imageFilter = imageFilter

        if isinstance(command, Command):
            self.command = command

        self.__validate()

    def __validate(self):
        if self.command is None:
            raise NameError('command cannot be None')


class SnapshotParameter:

    def __init__(self, name, stateComponentFilter, inventoryHardwareFilter, targetFilter,  expirationTime):
        self.name = None
        self.stateComponentFilter = None
        self.inventoryHardwareFilter = None
        self.targetFilter = None
        self.expirationTime = None
        if isinstance(stateComponentFilter, StateComponentFilter):
            self.stateComponentFilter = stateComponentFilter

        if isinstance(inventoryHardwareFilter, InventoryHardwareFilter):
            self.inventoryHardwareFilter = inventoryHardwareFilter

        if isinstance(targetFilter, TargetFilter):
            self.targetFilter = targetFilter

        if isinstance(expirationTime, datetime.datetime):
            self.expirationTime = expirationTime

        if name is not None:
            self.name = name

        self.__validate()

    def __validate(self):
        if self.name is None:
            raise NameError('name cannot be None')


class Image:

    def __init__(self, deviceType, manufacturer, models, target, s3URL, firmwareVersion, semanticFirmwareVersion, updateURI="", versionURI="", needReboot=False,  allowableDeviceStates=[], tags=["default"]):

        self.deviceType = deviceType
        self.manufacturer = manufacturer
        self.models = models
        self.target = target
        self.tags = tags
        self.firmwareVersion = firmwareVersion
        self.semanticFirmwareVersion =semanticFirmwareVersion
        self.updateURI = updateURI
        self.versionURI = versionURI
        self.needReboot =needReboot
        self.s3URL = s3URL
        self.allowableDeviceStates = allowableDeviceStates

        self.__validate()

    def __validate(self):
        if self.deviceType is None:
            raise NameError('deviceType cannot be None')

        if self.manufacturer is None:
            raise NameError('manufacturer cannot be None')

        if self.models is None:
            raise NameError('models cannot be None')

        if self.target is None:
            raise NameError('target cannot be None')

        if self.firmwareVersion is None:
            raise NameError('firmwareVersion cannot be None')

        if self.semanticFirmwareVersion is None:
            raise NameError('semanticFirmwareVersion cannot be None')

        if self.s3URL is None:
            raise NameError('s3URL cannot be None')
