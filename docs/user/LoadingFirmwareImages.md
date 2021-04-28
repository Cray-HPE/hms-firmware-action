# Loading Firmware Images into FAS
**`NOTE:` These procedures are for FAS version 1.7.8 and later**

## FAS Loader
In FAS v1.7.8, FAS switched from a loader job to a API based loader.
FAS also added the ability to load individual RPM files as well as read from Nexus.
**Only one Loader command can be run at a time.**
If the loader is busy, it will return an error and you will have to try to run later.

## Loader Status
To check if the loader is currently busy and receive a list of loader run IDs:
```bash
cray fas loader list

loaderStatus = "ready"
loaderRunList = [ "6d9f57a4-3d30-47e1-81e4-c159758993df", "c8704694-2784-45b8-b92b-7da23baf7297",]
```
or if using the API:
```bash
GET fas/v1/loader

{
  "loaderStatus": "ready",
  "loaderRunList": [
    "6d9f57a4-3d30-47e1-81e4-c159758993df",
    "c8704694-2784-45b8-b92b-7da23baf7297"
  ]
}
```

## RPM Files
Firmware for FAS is released in RPM files which contains one or more firmware images along with an image meta data file which describes the image for FAS.
These images may be released as a bundle in the Nexus repository or as an individual file.

## ZIP files
The firmware loader now has the ability to extract images and image data from zip files.
These zip files must contain both the binary image and image meta data file which describes the image for FAS.

## Loading Firmware From Nexus
Firmware may be released and placed into the Nexus repository.
FAS will return a loaderRunID.
Use the loaderRunID to check the results of the loader run.
To load the firmware from Nexus into FAS, use the following command:
```bash
cray fas loader nexus update

loaderRunID = "c2b7e9bb-f428-4e4c-aa83-d8fd8bcfd820"
```
or if using the API:
```bash
POST fas/v1/loader/nexus

{"loaderRunID":"df56e72a-1d81-469d-887c-a1ed6ec3a82b"}
```

## Loading Individual RPM or ZIP into FAS
To load an RPM or ZIP into FAS on a system, copy the RPM or ZIP file to m001 or one of the ncn.
FAS will return a loaderRunID.
Use the loaderRunID to check the results of the loader run.
Run the following command (RPM is this case is firmware.rpm):
```bash
cray fas loader create firmware.rpm

loaderRunID = "dd37dd45-84ec-4bd6-b3c9-7af480048966"
```
or if using the API:
```bash
POST fas/v1/loader -F "file=@firmware.rpm"

{"loaderRunID":"dd37dd45-84ec-4bd6-b3c9-7af480048966"}
```
*`NOTE:` if firmware is not in the current directory, you will need to add the path to the filename*

## Display Results of Loader Run

Using the loaderRunID returned from the loader upload command, run the following command to get the output from the upload *(Note the --format json, this makes it easier to read)*:

```bash
cray fas loader describe dd37dd45-84ec-4bd6-b3c9-7af480048966 --format json

{
  "loaderRunOutput": [
    "2021-04-28T14:40:45Z-FWLoader-INFO-Starting FW Loader, LOG_LEVEL: INFO; value: 20",
    "2021-04-28T14:40:45Z-FWLoader-INFO-urls: {'fas': 'http://localhost:28800', 'fwloc': 'file://download/'}",
    "2021-04-28T14:40:45Z-FWLoader-INFO-Using local file: /ilo5_241.zip",
    "2021-04-28T14:40:45Z-FWLoader-INFO-unzip /ilo5_241.zip",
    "Archive:  /ilo5_241.zip",
    "  inflating: ilo5_241.bin",
    "  inflating: ilo5_241.json",
    "2021-04-28T14:40:45Z-FWLoader-INFO-Processing files from file://download/",
    "2021-04-28T14:40:45Z-FWLoader-INFO-get_file_list(file://download/)",
    "2021-04-28T14:40:45Z-FWLoader-INFO-Processing File: file://download/ ilo5_241.json",
    "2021-04-28T14:40:45Z-FWLoader-INFO-Uploading b73a48cea82f11eb8c8a0242c0a81003/ilo5_241.bin",
    "2021-04-28T14:40:45Z-FWLoader-INFO-Metadata {'imageData': \"{'deviceType': 'nodeBMC', 'manufacturer': 'hpe', 'models': ['ProLiant XL270d Gen10', 'ProLiant DL325 Gen10', 'ProLiant DL325 Gen10 Plus', 'ProLiant DL385 Gen10', 'ProLiant DL385 Gen10 Plus', 'ProLiant XL645d Gen10 Plus', 'ProLiant XL675d Gen10 Plus'], 'targets': ['iLO 5'], 'tags': ['default'], 'firmwareVersion': '2.41 Mar 08 2021', 'semanticFirmwareVersion': '2.41.0', 'pollingSpeedSeconds': 30, 'fileName': 'ilo5_241.bin'}\"}",
    "2021-04-28T14:40:46Z-FWLoader-INFO-IMAGE: {\"s3URL\": \"s3:/fw-update/b73a48cea82f11eb8c8a0242c0a81003/ilo5_241.bin\", \"target\": \"iLO 5\", \"deviceType\": \"nodeBMC\", \"manufacturer\": \"hpe\", \"models\": [\"ProLiant XL270d Gen10\", \"ProLiant DL325 Gen10\", \"ProLiant DL325 Gen10 Plus\", \"ProLiant DL385 Gen10\", \"ProLiant DL385 Gen10 Plus\", \"ProLiant XL645d Gen10 Plus\", \"ProLiant XL675d Gen10 Plus\"], \"softwareIds\": [], \"tags\": [\"default\"], \"firmwareVersion\": \"2.41 Mar 08 2021\", \"semanticFirmwareVersion\": \"2.41.0\", \"allowableDeviceStates\": [], \"needManualReboot\": false, \"pollingSpeedSeconds\": 30}",
    "2021-04-28T14:40:46Z-FWLoader-INFO-Number of Updates: 1",
    "2021-04-28T14:40:46Z-FWLoader-INFO-Iterate images",
    "2021-04-28T14:40:46Z-FWLoader-INFO-update ACL to public-read for 5ab9f804a82b11eb8a700242c0a81003/wnc.bios-1.1.2.tar.gz",
    "2021-04-28T14:40:46Z-FWLoader-INFO-update ACL to public-read for 5ab9f804a82b11eb8a700242c0a81003/wnc.bios-1.1.2.tar.gz",
    "2021-04-28T14:40:46Z-FWLoader-INFO-update ACL to public-read for 53c060baa82a11eba26c0242c0a81003/controllers-1.3.317.itb",
    "2021-04-28T14:40:46Z-FWLoader-INFO-update ACL to public-read for b73a48cea82f11eb8c8a0242c0a81003/ilo5_241.bin",
    "2021-04-28T14:40:46Z-FWLoader-INFO-finished updating images ACL",
    "2021-04-28T14:40:46Z-FWLoader-INFO-removing local file: /ilo5_241.zip",
    "2021-04-28T14:40:46Z-FWLoader-INFO-*** Number of Updates: 1 ***"
  ]
}
```
or if using the API:
```bash
GET fas/v1/loader/dd37dd45-84ec-4bd6-b3c9-7af480048966

(SAME OUTPUT AS CRAY CLI ABOVE)
```
