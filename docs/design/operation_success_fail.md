### How FAS determines success or failure of an operation

Once FAS sends the update command to redfish (redfish/v1/UpdateService) it begins checking if the operation succeeded or failed.

Process for determining success or failure:

1. FAS checks the version of the target updating by comparing the target version string returned by a redfish call (redfish/v1/UpdateService/FirmwareInventory) and the version string in the image record along with the version that was present at the start of the operation (FromFirmwareVersion)

  a. If the version is the same as the image record version string then FAS declares successful operation.

  b. If the version is not the same as the image record version string, but is different from the FromFirmwareVersion, FAS will declared it failed and present both strings to the user to determine if the operation was actually successful.
  The version strings (image record and redfish) must match exactly to be successful.
  If the image version string is incorrect, the update will be declared a fail, but most likely update was successful.

  c. Upon finding no change in the firmware version string after the time limit has expired, FAS will declare the operation has failed.

2. On some target updates, the version string returned from redfish will not be updated until a node is rebooted or some other action takes place.
To determine if the operation has completed FAS will perform additional checks.

  a. iLO devices will create a task upon receiving a firmware update request.
  The task id and link is returned with the update request.
  FAS will check that task link and use that to determine if the operation has successfully completed or had an error.

  b. Gigabyte devices have an UpdateInformation section in the for the UpdateService (redfish/v1/UpdateService).
  This information will show the update status of the current operation on that device.
  FAS will use that update status to determine if the update has successfully completed or had an error.
