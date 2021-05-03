## Introduction

The Firmware Action Service (FAS) provides an interface for managing firmware versions of Redfish-enabled hardware in the system. FAS interacts with the Hardware State Managers (HSM), device data, and image data in order to update firmware.

FAS images contain the following information that is needed for a hardware device to update firmware versions:

* Hardware-specific information: Contains the allowed device states and how to reboot a device if necessary.
* Selection criteria: How to link a firmware image to a specific hardware type.
* Image data: Where the firmware image resides in Simple Storage Service (S3) and what firmwareVersion it will report after it is successfully applied. See "Artifact Management" in the *HPE Cray EX Administration Guide S-8001* for more information about S3. [ANIEUWSMA-LINK?]

### Warning

**WARNING:** Non-compute nodes (NCNs) should be locked with the HSM locking API to ensure they are not unintentionally updated by FAS. Research "*NCN and Management Node Locking*" for more information. [ANIEUWSMA-LINK?]  Failure to lock the NCNs could result in unintentional update of the NCNs if FAS is not used correctly; this will lead to system instability problems.

### Current Capabilities as of Shasta Release v1.4

The following table describes the hardware items that can have their firmware updated via FAS.

*Table 1. Upgradable Firmware Items*

| **Manufacturer** | **Type**   | **Target**                                                   | **New in Release 1.4**                     |
| ---------------- | ---------- | ------------------------------------------------------------ | ------------------------------------------ |
| Cray             | nodeBMC    | `BMC`, `Node0.BIOS`,  `Node1.BIOS`,  `Recovery`, `Node1.AccFPGA0`, `Node0.AccFPGA0` | Node1.AccFPGA0  and Node0.AccFPGA0 targets |
| Cray             | chassisBMC | `BMC`, `Recovery`                                            |                                            |
| Cray             | routerBMC  | `BMC`, `Recovery`                                            |                                            |
| Gigabyte         | nodeBMC    | `BMC`, `BIOS`                                                |                                            |
| HPE              | nodeBMC    | `iLO 5` (BMC aka `1` ), `System ROM` ,`Redundant System ROM` (BIOS aka `2`) | `iLO 5` and `System ROM` targets           |

### New in Release 1.4

The following enhancements have been made to FAS for the HPE Cray EX 1.4 release:

* FAS can now be used to update more hardware components, such as User Access Nodes (UANs), compute nodes (CNs), and NCNs.
* New command parameters are available for use when creating JSON files for FAS jobs. The following parameters are now supported:
  * `overrideImage`: Override a firmware image with a new one on any FAS supported hardware device, even if the firmware image doesn't match the hardware. This parameter is part of the imageFilter payload mentioned in the [*FAS* *Filters for Updates and Snapshots* ](#_bookmark1)on page 5 section [ANIEUWSMA-ANCHOR?].
  * `overWriteSameImage`: Overwrite a firmware image with the same image that is currently on the hardware device. This parameter is part of the command payload mentioned in the [*FAS* *Filters for*](#_bookmark1)[ *Updates and Snapshots* ](#_bookmark1)on page 5 section [ANIEUWSMA-ANCHOR?].
* A new API endpoint, and corresponding CLI command: `cray fas operations`, has been added to help make it easier to view the details of a FAS action. Refer to "Manage Firmware Updates with FAS" in the *HPE Cray EX Hardware Management Administration Guide S-8015* for more information.  [ANIEUWSMA-LINK?]
* A new API endpoint, and corresponding CLI command:`cray fas actions status list {actionID}`,  has been added to help make it easier to view the summary counts of a FAS action.

### FAS Use Cases

There are several use cases for using the FAS to update firmware on the system. These use cases are intended to be run by system administrators with a good understanding of firmware. Under no circumstances should non-admin users attempt to use FAS or perform a firmware update.

-   Perform a firmware update: Update the firmware of an xname's target to the latest, earliest, or an explicit version.
-   Determine what hardware can be updated by performing a dry-run: The easiest was to determine what can be updated is to perform a dry-run of the update.
-   Take a snapshot of the system: Record the firmware versions present on each target for the identified xnames. If the firmware version corresponds to an image available in the images repository, link the `imageID` to the record.
-   Restore the snapshot of the system: Take the previously recorded snapshot and use the related `imageIDs` to put the xname/targets back to the firmware version they were at, at the time of the snapshot.
-   Provide firmware for updating: FAS can only update an xname/target if it has an image record that is applicable. Most admins will not encounter this use case.

### Firmware Actions

An action is collection of operations, which are individual firmware update tasks. Only one FAS action can be run at a time. Any other attempted action will be queued. Additionally, only one operation can be run on an xname at a time. For example, if there are 1000 xnames with 5 targets each to be updated, all 1000 xnames can be updating a target, but only 1 target on each xname will be updated at a time.

The life cycle of any action can be divided into the static and dynamic portions of the life cycle.

The static portion of the life cycle is where the action is created and configured. It begins with a request to create an action through either of the following requests:

-   Direct: Request to /actions API.
-   Indirect: Request to restore a snapshot via the /snapshots API.

The dynamic portion of the life cycle is where the action is executed to completion. It begins when the actions is transitioned from the `new` to `configured` state. The action will then be ultimately transitioned to an end state of `aborted` or `completed`.

### FAS Filters for Updates and Snapshots

FAS uses five primary filters to determine what operations to create. The filters are listed below:

-   `command`
-   `stateComponentFilter`
-   `targetFilter`
-   `inventoryHardwareFilter`
-   `imageFilter`

All filters are logically connected with `AND` logic. Only the `stateComponentFilter`, `targetFilter`, and `inventoryHardwareFilter` are used for snapshots.

- **`command`**

  The command group is the most important part of an action command and controls if the action is executed as dry-run.

  It also determines whether or not to override an operation that would normally not be executed if there is no way to return the xname/target to the previous firmware version. This happens if an image does not exist in the image repository.

- **`stateComponentFilter`**

  The state component filter allows users to select hardware to update. Hardware can be selected individually with xnames, or in groups by leveraging the Hardware State Manager \(HSM\) groups and partitions features.

- **`targetFilter`**

  The target filter removes targets from the candidate list when they do not match the targets. For example, if the user specifies only the BIOS target, FAS would remove any candidate operation that was not explicitly for BIOS.

  A Redfish device has potentially many targets \(members\). Targets for FAS are case sensitive and must match Redfish.

  Examples include, but are not limited to the following:

  -   BIOS
  -   BMC
  -   NIC
  -   Node0.BIOS
  -   Node1.BIOS
  -   Recovery

- **`inventoryHardwareFilter`**

  The inventory hardware filter takes place after the state component filter has been applied. It will remove any devices that do not conform to the identified manufacturer or models determined by querying the Redfish endpoint.

  **Important:** There can be a mismatch of hardware models. The `model` field is human-readable and is human-programmable. In some cases, there can be typos where the wrong model is programmed, which causes issues filtering. If this occurs, query the hardware, find the model name, and add it to the images repository on the desired image.

- **`imageFilter`**

  FAS applies images to xname/targets. The image filter is a way to specify an explicit image that should be used. When included with other filters, the image filter reduces the devices considered to only those devices where the image can be applied.

  For example, if a user specifies an image that only applies to gigabyte, nodeBMCs, BIOS targets. If all hardware in the system is targeted with an empty `stateComponentFilter`, FAS would find all devices in the system that can be updated via Redfish, and then the image filter would remove all xname/targets that this image could not be applied. In this example, FAS would remove any device that is not a gigabyte nodeBMC, as well as any target that is not BIOS.


### Firmware Images

FAS requires images in order to update firmware for any device on the system. An image contains the data that allows FAS to establish a link between an admin command, available devices \(xname/targets\), and available firmware.

The following is an example of an image:

```screen
{
  "imageID": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "createTime": "2020-05-11T17:11:07.017Z",
  "deviceType": "nodeBMC",
  "manufacturer": "intel",
  "model": ["s2600","s2600_REV_a"],
  "target": "BIOS",
  "tag": ["recovery", default"],
  "firmwareVersion": "f1.123.24xz",
  "semanticFirmwareVersion": "v1.2.252",
  "updateURI": "/redfish/v1/Systems/UpdateService/BIOS",
  "needManualReboot": true,
  "waitTimeBeforeManualRebootSeconds": 600,
  "waitTimeAfterRebootSeconds": 180,
  "pollingSpeedSeconds": 30,
  "forceResetType": "ForceRestart",
  "s3URL": "s3://firmware/f1.1123.24.xz.iso",
  "allowableDeviceStates": [
    "On",
    "Off"
  ]
}
```

The main components of an image are described below:

- **Key**

  This includes the `deviceType`, `manufacturer`, `model`, `target`, `tag`, `semanticFirmwareVersion` \(firmware version\) fields.

  These fields are how admins assess what firmware is on a device, and if an image is applicable to that device.

- **Process guides**

  This includes the `forceResetType`, `pollingSpeedSeconds`, `waitTime(s)`, `allowableDeviceStates` fields.

  FAS gets information about how to update the firmware from these fields. These values determine if FAS is responsible for rebooting the device, and what communication pattern to use.

- **`s3URL`**

  The URL that FAS uses to get the firmware binary and the download link that is supplied to Redfish devices. Redfish devices are not able to directly communicate with S3.