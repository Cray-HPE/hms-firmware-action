# FAS | Understanding Images

### Change log 


|Date	|Author	|Description|
| ---- | ---- | ---- |
|2020-05-11	|@anieuwsma	|	initial revision|
|2020-05-15	|@anieuwsma	|	included example imagefile and depfile|
|2020-06-12	|@anieuwsma	|	included metadatafile contents|
|2020-10-15 |@mbuchmann | Added softwareIds|

## Introduction & Intended Audience
FAS (firmware action service) is the successor to FUS (firmware update service).   FAS will be part of Shasta v1.3 and FUS will be deprecated/ removed from Shasta in the v1.3.  FAS exists to provide a common interface for managing the firmware versions of hardware in a Shasta system via Redfish.

Both FAS and FUS have a RESTful API, data model, and related tools (CLI, firmware loader). The concept of an image is seen slightly differently if engaging with FAS through the API or the firmware loader, or if just considering the underlying data model.  I will clarify which perspective I am describing in this document.

FUS had the concept of `dependencies` and a dependency file (aka `depfile`) that was used by the firmware loader.  FAS has replaced that concept with `images` and an `imagefile`.  The APIs and the represent data models are similar but have very specific differences.  This paper will explain those differences and what is required by FAS.

This paper is intended for the creators of firmware who would like to enroll their firmware in FAS and allow it to be installed onto hardware targets.  e.g. updating the BIOS firmware on a nodeBMC.

For more details refer to the swagger specification for FAS.

## What is an image?
Conceptually an image is the data needed to update the firmware of a device.  This data includes a pointer to a literal binary blob (that would be loaded into a firmware slot) + the meta data that FAS needs in order to know what to target and how to update the device.  

Image data is used by FAS to know HOW/WHAT to do to perform an update.  Ex, if an admin instructs FAS to upgrade the BIOS target to latest on the nodeBMC's, FAS will use hardware state manager data, device data (via redfish), and image data, to know how/what to update.

### Image Data Model

An image contains hardware specific information (like how to reboot the device (if necessary), or the allowed device states), selection criteria (how to link a firmware Image to a specific hardware type), and image information (like where the image resides in s3, and what firmwareVersion it will report after successfully applied).

An example image (JSON representation):

```
{
  "imageID": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "createTime": "2020-05-11T17:11:07.017Z",
  "deviceType": "nodeBMC",
  "manufacturer": "cray",
  "models": ["s2600","s2600_REV_a"],
  "target": "BMC",
  "softwareIds": ["nc:*:*:*"],
  "tag": ["recovery", default"],
  "firmwareVersion": "f1.123.24xz",
  "semanticFirmwareVersion": "1.2.252",
  "updateURI": "/redfish/v1/Systems/UpdateService/BMC",
  "versionURI": "/redfish/v1/Systems/UpdateService/FirmwareInventory",
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

#### Fields

*Note: every field is case sensitive.

|Name	|DataType|Optional	|Description|Source|
| --- | --- | --- | --- | --- |
|imageID|string - UUID|automatic|The ID of the record used by FAS|API|
|createTime|string - dateTime|automatic|The TIME the record was created|API|
|deviceType|string |required - if softwareIds not defined|Type of device as seen in HSM|User|
|manufacturer|string |required - if softwareIds not defined|Manufacturer defines the redfish structure must be: 'cray', 'gigabyte', or 'hpe'|User|
|models|array of strings|required - if softwareIds not defined|Model LITERAL as reported by device via Redfish|User|
|target|string|required - if softwareIds not defined|member LITERAL to update, as reported by device via Redfish (i.e.: BIOS, BMC,...)|User|
|softwareIds|array of strings|optional|List of software ids that this firmware can be applied to.  This must match the SoftwareId returned by FirmwareInventory|User|
|tag|array of strings|required|firmware image specifier that allows for multiple  firmware images to have the same semantic firmware version. Ex) allowing 'recovery' firmware to be loaded onto a device for testing.|User|
|firmwareVersion|string|required|LITERAL firmware version reported by device via Redfish|User|
|semanticFirmwareVersion|string - semantic version|required|LITERAL semantic version|User|
|updateURI|string|optional|The redfish path to use to perform the update, will override the discovered value|User|
|needManualReboot|bool|optional - defaults to false|True if the device needs to be manually rebooted by FAS after performing a firmware action.  False if the device will automatically perform a reboot on its own, or if no reboot of any kind is needed as part of firmware action.|User|
|waitTimeBeforeManualRebootSeconds|int|required - if needManualReboot is TRUE|Amount of time to wait after issuing an update command before rebooting device|User|
|waitTimeAfterManualRebootSeconds|int|required - if needManualReboot is TRUE|amount of time to wait after issuing a reboot command before attempting to verify (should accommodate for time for OS to come up)|User|
|pollingSpeedSeconds|int|optional - default 30s|How long to wait between power state queries or version queries, to not overload device|User|
|forceResetType|string|required - if needManualReboot is TRUE|The force reset command to issue|User|
|s3URL|string - url|required|the S3 url the firmware image is located at|User|
|allowableDeviceStates|array of strings|optional - default all|the allowed device states (PowerState) the device must be in to preform the update. PowerState as reported by device via Redfish.|User|

## Image File

An imagefile is used by the firmware loader to create entries in FAS. The firmware loader is an operational tool for loading the 'shipped' firmware images into the system at time of deployment.  The firmware loader is a convenience, but is not required to use FAS, however without loading at least one image into FAS, there would be nothing that FAS could do.

The firmware loader will read image rpms from the Nexus Repository, upload the image into S3 and create an image record in FAS to use the S3 image.
The imagefile must reside in the rpm with the image.

Here is an example `imagefile`:

```
{
  "deviceType": "nodeBMC",
  "manufacturer": "cray",
  "models": ["c5000","c5000_REV_a", "c5000 mk2"],
  "targets": ["BMC"],
  "softwareIds": ["nc:*:*:*"],
  "tags": ["persist_root", "default"],
  "firmwareVersion": "f1.123.24xz",
  "semanticFirmwareVersion": "1.2.252",
  "updateURI": "/redfish/v1/UpdateService/Inventory/FOO",
  "needManualReboot": true,
  "waitTimeBeforeManualRebootSeconds": 0,
  "waitTimeAfterRebootSeconds": 0,
  "pollingSpeedSeconds": 0,
  "forceResetType": "ForceRestart",
  "fileName": "myFirmwareFile.itb"
  "allowableDeviceStates": ["On","Off"  ]
}
```

This above `imagefile` would be turned into this `image` record:

```
{
  "imageID": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "createTime": "2020-05-15T13:06:47.188Z",
  "deviceType": "nodeBMC",
  "manufacturer": "cray",
  "models": ["c5000","c5000_REV_a", "c5000 mk2"],
  "target": "BMC",
  "softwareIds": ["nc:*:*:*"],
  "tags": ["recovery", "default"],
  "firmwareVersion": "f1.123.24xz",
  "semanticFirmwareVersion": "1.2.252",
  "updateURI": "/redfish/v1/UpdateService/Inventory/FOO",
  "needManualReboot": true,
  "waitTimeBeforeManualRebootSeconds": 0,
  "waitTimeAfterRebootSeconds": 0,
  "pollingSpeedSeconds": 0,
  "forceResetType": "ForceRestart",
  "s3URL": "s3://firmware/f1.1123.24.xz.iso",
  "allowableDeviceStates": ["On","Off"  ]
}
```

## Field Descriptions

|Field	|Required|Type	|Default| Description|Note|
| --- | --- | --- | --- | --- | --- |
|`deviceType`|Required - if softwareIds not defined|String|| This is the device type for the redfish endpoint.Device Types are defined by Hardware State Manager (HSM).Below is a list of valid Device Types for FAS: <ul><li>CabinetBMC</li><li>ChassisBMC</li><li>NodeBMC</li><li>RouterBMC</li></ul> ||
|`manufacturer`|Required - if softwareIds not defined|String|| Manufacturer of the device. Currently FAS supports devices from: <ul><li> Cray </li><li> Gigabyte </li><li> HPE </li></ul> Must be a string of `cray`, `gigabyte`, or `hpe`||
|`targets`|Required - if softwareIds not defined|Array of Strings||The target field is a list of device targets for update.These are the actual devices that are being updated as defined in either the redfish `FirmwareInventory` or `SoftwareInventory`.|**NOTE: The target value is case sensitive and must match what is returned in the redfish inventory.** <p>Multiple targets can be specified in the Meta-Data File. <p>FWLoader will create a separate image entry for FAS for each target listed.|
|`models`|Required - if softwareIds not defined|Array of Strings|| The models field is used to indicate the different device models this firmware image can update. <p>The model must match the model returned from redfish.<p>Multiple model strings can be specified in this field.| **NOTE: Some devices report different model string.This field must contain all these different strings in order to update the device.** |
|`softwareIds `|Optional|Array of Strings||<p>List of software ids that this firmware can be applied to.  This must match the SoftwareId returned by FirmwareInventory||
|`tags`| Required|Array of Strings|| <p>Firmware for the same device with the same version, but have different uses can specify tags to distinguish the images from each other. <p>Multiple tags can be specified for each image. <p>One image (and only one) must be tagged as `default` which will be used when no tag is given for an update.||
|`semanticFirmwareVersion`|Required|String|| <p>This must be a valid semantic number (i.e. 1.2.3). <p>This number is used to determine the latest firmware version for a device.||
|`firmwareVersion`|Required| String|| <p>This is the version string embedded in the image, which is returned by the redfish inventory call for the firmware version.<p> Examples are `sc.prod-master-202.arm64.2019-07-03T22:02:54+00:00.fbf41d70` or `wnc.bios-0.5.1`.|**NOTE: Must match exactly to what redfish will report.**|
|`fileName`|Required|String|| <p>This is the name of the file containing the image. <p>This is how we associate this meta-data with a particular image. ||
|`needManualReboot`|Optional|Boolean|defaults to `false`|<p>This is a Boolean flag indicating that once the firmware update is complete, the device should be rebooted.<p>If the reboot occurs automatically after the update, this flag should be set to False. <p>||
|`allowableDeviceStates`|Optional|Array of Strings| defaults to `[]`|<p>The deviceStates field is a list acceptable states that the device can be in in order to perform a firmware update on the device.<p>This is intended to provide a guard rail against performing an update on a device that is not in the correct state.<p> Acceptable states are the component states defined in the hardware state manager (HSM).|The default is ```[ ]```. <p>In this case, the update will be allowed in any state.|
|`forceResetType`|Depends| string|| The redfish command used to do a manual reboot.|REQUIRED IF `needManualReboot` = true |
|`waitTimeBeforeManualRebootSeconds`|Depends | integer|| If FAS needs to perform a manual reboot, how long to wait before issuing the reboot command.| REQUIRED IF `needManualReboot` = true|
|`waitTimeAfterRebootSeconds `|Depends | integer|| Amount of time to wait after reboot command issued to check on device state and firmware version.| REQUIRED IF `needManualReboot` = true|
|`pollingSpeedSeconds`|Optional|integer| `30` | Amount of time to wait between check firmware status state (in seconds)|
|`updateURI`|Optional|String||<p>FAS automatically figures out the updateURI redfish path.<p>If the firmware needs a different path, it can be specified using this field.||
|`versionURI `|Optional|String||<p>FAS automatically figures out the versionURI redfish path.<p>If the firmware needs a different path, it can be specified using this field.||
