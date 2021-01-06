# Firmware Action Service (FAS) Administration Guide

## Introduction

The Firmware Action Service (FAS) provides an interface for managing firmware versions of Redfish-enabled hardware in the system. FAS interacts with the Hardware State Managers (HSM), device data, and image data in order to update firmware.

FAS images contain the following information that is needed for a hardware device to update firmware versions:

* Hardware-specific information: Contains the allowed device states and how to reboot a device if necessary.
* Selection criteria: How to link a firmware image to a specific hardware type.
* Image data: Where the firmware image resides in Simple Storage Service (S3) and what firmwareVersion it will report after it is successfully applied. See "Artifact Management" in the *HPE Cray EX Administration Guide S-8001* for more information about S3. [ANIEUWSMA-LINK?]

## Warning

**WARNING:** Non-compute nodes (NCNs) should be locked with the HSM locking API to ensure they are not unintentionally updated by FAS. Research "*NCN and Management Node Locking*" for more information. [ANIEUWSMA-LINK?]  Failure to lock the NCNs could result in unintentional update of the NCNs if FAS is not used correctly; this will lead to system instability problems.

## Current Capabilities as of Shasta Release v1.4

The following table describes the hardware items that can have their firmware updated via FAS.

*Table 1. Upgradable Firmware Items*

| **Manufacturer** | **Type**   | **Target**                                                   | **New in Release 1.4**                     |
| ---------------- | ---------- | ------------------------------------------------------------ | ------------------------------------------ |
| Cray             | nodeBMC    | `BMC`, `Node0.BIOS`,  `Node1.BIOS`,  `Recovery`, `Node1.AccFPGA0`, `Node0.AccFPGA0` | Node1.AccFPGA0  and Node0.AccFPGA0 targets |
| Cray             | chassisBMC | `BMC`, `Recovery`                                            |                                            |
| Cray             | routerBMC  | `BMC`, `Recovery`                                            |                                            |
| Gigabyte         | nodeBMC    | `BMC`, `BIOS`                                                |                                            |
| HPE              | nodeBMC    | `iLO 5` (BMC aka `1` ), `System ROM` ,`Redundant System ROM` (BIOS aka `2`) | `iLO 5` and `System ROM` targets           |

## New in Release 1.4

The following enhancements have been made to FAS for the HPE Cray EX 1.4 release:

* FAS can now be used to update more hardware components, such as User Access Nodes (UANs), compute nodes (CNs), and NCNs.
* New command parameters are available for use when creating JSON files for FAS jobs. The following parameters are now supported:
  * `overrideImage`: Override a firmware image with a new one on any FAS supported hardware device, even if the firmware image doesn't match the hardware. This parameter is part of the imageFilter payload mentioned in the [*FAS* *Filters for Updates and Snapshots* ](#_bookmark1)on page 5 section [ANIEUWSMA-ANCHOR?].
  * `overWriteSameImage`: Overwrite a firmware image with the same image that is currently on the hardware device. This parameter is part of the command payload mentioned in the [*FAS* *Filters for*](#_bookmark1)[ *Updates and Snapshots* ](#_bookmark1)on page 5 section [ANIEUWSMA-ANCHOR?].
* A new API endpoint, and corresponding CLI command: `cray fas operations`, has been added to help make it easier to view the details of a FAS action. Refer to "Manage Firmware Updates with FAS" in the *HPE Cray EX Hardware Management Administration Guide S-8015* for more information.  [ANIEUWSMA-LINK?]
* A new API endpoint, and corresponding CLI command:`cray fas actions status list {actionID}`,  has been added to help make it easier to view the summary counts of a FAS action.

