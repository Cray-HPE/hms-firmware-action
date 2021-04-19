# Loading Firmware Images into FAS
**`NOTE:` These procedures are for v1.5 and later FAS version 1.7.5 and later**

## FAS Loader
In v1.5 (FAS v1.7.5), FAS switched from a loader job to a API based loader.
FAS also added the ability to load individual RPM files as well as read from Nexus.
**Only one Loader command can be run at a time.**
If the loader is busy, it will return an error and you will have to try to run later.

## Loader Status
To check if the loader is currently busy and receive the output from the last loader run, use the following command:
```bash
cray fas loader list
```
or if using the API:
```bash
GET fas/v1/loader
```

## RPM Files
Firmware for FAS is released in RPM files which contains one or more firmware images along with an image meta data file which describes the image for FAS.
These images may be released as a bundle in using the Nexus repository or as individual files.

## Using Nexus
Firmware may be released and placed into the Nexus repository.
To load the firmware from Nexus into FAS, use the following command:
```bash
cray fas loader update
```
or if using the API:
```bash
PUT fas/v1/loader
```

## Loading Individual RPM into FAS
To load an RPM into FAS on a system, copy the RPM to m001 or one of the ncn.
Run the following command (RPM is this case is firmware.rpm):
```bash
cray fas loader create firmware.rpm
```
or if using the API:
```bash
POST fas/v1/loader -F --file@firmware.rpm
```
*`NOTE:` if firmware is not in the current directory, you will need to add the path to the filename*

## Display Results of Loader Run
*SEE Loader Status Above*
