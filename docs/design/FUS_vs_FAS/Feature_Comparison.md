## Feature Comparison Between FUS/FAS v1.0
<em> 2020-04-17, @anieuwsma </em>

#### Definitions
* ACTION - A collection of operations initiated by user request to update to the firmware images on a set of hardware. Ex) Update the gigabyte BMC targets to latest.

* OPERATION - An update (upgrade/downgrade) to a specific device's Firmware Target. Ex) Update x0c0s1b0 BIOs Target to v1.2.0

* SNAPSHOT - a point in time record of what firmware images were running on the system (a device's targets), constrained by user defined parameters (xname, model/manufacturer, etc).  Used to 'RESTORE' the system part back to specific firmware versions.

#### FEATURES

|Feature|FAS|FUS|Notes|
|---|---|---|---|
|Create an action? | Y  | Y  |   |
|Perform a dry-run (simulated action)? | Y  | Y  |   |
|Retrieve an action?  | Y  | Y  |   |
|Retrieve summary information about an action? | Y  | N  |FUS has no rollup capability, just raw data.  FAS allows for easily accessible summary information at several levels of information hierarch.  |
|Abort an action?| Y  | N  |FAS will allow users to abort an action, and set an auto-expiration policy for too-long run operations.|
|Delete the record of an action?| Y  | N  |FAS will allow users to permanently delete the record of an action.|
|Includes expanded job creation parameters?| Y  | N  |FAS will allow clients to create actions based on the added parameters of Model/Manufacture, Desired Firmware Image|
|Create a snapshot of the system?| Y  | Y  |   |
|Retrieve summary information about a snapshot?| Y  | N  |FUS has no rollup capability, just raw data.  FAS allows for easily accessible summary information at several levels of information hierarch. |
|Delete a snapshot? | Y  | N  |   |
|Restore a snapshot?| Y  | Y  |   |
|Load a backup snapshot? | Y  | N  |FAS will allow clients to upload a 'backed up' snapshot of a system, disaster recovery scenario.|
|Create an Image record? | Y  | Y*  |FUS had a partial idea of creating an image record, but the idea wasn't functionally viable, it instead depended on a file to be loaded into the service instance at initialization time. |
|Update an Image record? | Y  | N*  |FUS had a partial idea of an update, but was not functionally viable.|
|Retrieve an Image record? | Y  | Y  |   |
|Delete an Image record? | Y  | N  |FAS will allow clients to delete an image from the FAS datastore.|
|RESTful API | Y  | N |FUS is not a RESTful API|
|Firmware loader| Y  | Y  |FUS/FAS both have the capability to load system firmware into S3/FAS-FUS for later use by the system|
|Multi-instance| Y  | N  | FUS is a singleton, FAS is being designed to run in multi-instance mode. |
|Specify image / device-target dependencies?| Y*  | N*  |Fully detailed elsewhere, but FUS dependency model was deeply flawed.  In practice it was never used, and would only support a partial dependency for an update.  FAS v1.0 will not likely have dependency supported; dependency management will likely be a feature in v1.4 timeline. The high level goal for FAS would be to support dependency management of images / device-targets for upgrade and downgrade.|



## FAS/FUS Key Differences

In FUS the file was referred to as the `depfile`.  Much of the same data that was in the `depfile` is in an image record, but several things are different.

For purposes of comparison and contrast, here is an an example of an `depfile` in FUS:


```
{
    "redfish_endpoint_type": "NodeBMC",
    "target": [ "BMC" ],
    "models": [ "Intel" ],
    "version": "1.2.5",
    "manufacturer_version": "foo.123.xies",
    "latest": true,
    "need_reboot": false,
    "filename":"nCfirmware-738.itb",
    "device_states": [],
    "dependencies": [
        {
            "redfish_endpoint_type": "NodeBMC",
            "target": ["Node0.BIOS","Node1.BIOS"],
            "req_version": "1.3.0"
        }
    ]
}
```

A few of the key differences are:

    1. The `semanticFirmwareVersion` MUST be filled with a valid `sem ver.`  This `sem ver` will be used to determine what image is considered latest or earliest.
    2. The `firmwareVersion` <strong>MUST match identically</strong> to what the Redfish device will report.
    3. If the device needs a manual reboot performed, then several fields must be filled out to specify how long to wait to perform the update: `waitTimeBeforeManualRebootSeconds`, `waitTimeAfterRebotSeconds`,  `forceResetType`.  Unfortunately in the Redfish specification there is no field that indicates the progress of a firmware update.  Our testing has revealed that different devices may introduce a slight (but significant) delay before updating the firmware.  This lack of determinism makes it programmatically impossible to know if an update is in progress or has been completed on some devices, so times must be specified to control the update progress.
    4. FAS v1.0 will NOT support the concept of dependencies (that is to say that in order to update the BIOS firmware on a nodeBMC to v3.0.0 the BMC firmware must first be at v2.5.0 or higher.  The FUS concept was not used, and was not operational. This will be a feature we look at for FAS v1.1.
    5. FUS had a concept of regex'd models; but this was a hidden configuration.  Instead FAS expands to allow an array of models which puts the necessity of correctness on the creators of images.