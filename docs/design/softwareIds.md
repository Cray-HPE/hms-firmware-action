# Software IDs

### Change log

|Date	|Author	|Description|
| ---- | ---- | ---- |
|2021-03-26	|@mbuchmann	|	initial revision|

Cray products introduced software ids in firmware release v1.4

*NOTE:* software ids are only supported on Cray redfish devices

Software ids are used to match firmware images in FAS with the correct firmware target on the device.  Each updatable target will have a firmware id which can be queried from redfish in the FirmwareInventory.

Example from Redfish:

```json
{
  "@odata.context": "/redfish/v1/$metadata#UpdateService/FirmwareInventory/Members/$entity",
  "@odata.etag": "W/\"1616104009\"",
  "@odata.id": "/redfish/v1/UpdateService/FirmwareInventory/Node0.BIOS",
  "@odata.type": "#SoftwareInventory.v1_1_0.SoftwareInventory",
  "Description": "Node0.BIOS",
  "Id": "Node0.BIOS",
  "Name": "Node0.BIOS",
  "SoftwareId": "bios.ex425.*.*",
  "Status": {
    "Health": "OK",
    "State": "Enabled"
  },
  "Updateable": true,
  "Version": "ex425.bios-1.4.3"
}
```

Software id for this firmware is `bios.ex425.*.*`

The image record stored in FAS looks like this:

```json
{
  "imageID": "2e062f62-7831-48d3-9c99-884d87d75c3c",
  "createTime": "2021-03-10T19:55:09Z",
  "deviceType": "nodeBMC",
  "manufacturer": "cray",
  "models": [
    "HPE CRAY EX425"
  ],
  "softwareIds": [
    "bios.ex425.*.*"
  ],
  "target": "Node0.BIOS",
  "tags": [
    "default"
  ],
  "firmwareVersion": "ex425.bios-1.4.3",
  "semanticFirmwareVersion": "1.4.3",
  "pollingSpeedSeconds": 30,
  "s3URL": "s3:/fw-update/8472b3e481da11eb8a6e22841ed8e377/ex425.bios-1.4.3.tar.gz"
}
```

When FAS encounters a software id from the redfish Firmware Inventory, it searches the available image records for a matching software id.
FAS will ignore all other fields (such as models, manufacturer, and target).
Once FAS finds all the valid image records containing the exact same software id as the redfish firmware inventory, it will select the correct image based off of the semanticFirmwareVersion key (latest or earliest).

If the redfish Firmware Inventory contains a software id, but no matching software id can be found in the FAS images, FAS will not flash the target and return `No image found` error.
