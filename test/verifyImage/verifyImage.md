### Verify Image Test Program

This program will read a .rpm or a .zip file, extract the content and check for completeness.
To be used on firmware packages to be loaded into FAS using the FAS firmware loader.

To run the test:

```bash
verifyImage.py -f {filename}
```
Where {filename} is the .rpm or .zip FAS firmware package

#### Usage
```bash
usage: verifyImage.py [-h] [-f F] [-s3 S3] [--nocolors]

optional arguments:
  -h, --help  show this help message and exit
  -f F        File to process
  -s3 S3      s3URL
  --nocolors  no colors
  ```

#### Output

verifyImage uses colors to indicate good / warning / error.
To turn off color output, use the `--nocolors` command line option

verifyImage will first extract the contents of the file into a directory named `./fas_verify_download`

verifyImage will then serach for .json files in the extracted contents.

verifyImage will check each .json file found.

`CHECKING FILE: xxx.json`

First will check for required items:

`------------- REQUIRED ITEMS -------------`

Check for optional items.
If an optional item is not found, it will report what FAS will default it to.

`------------- OPTIONAL ITEMS -------------`

Check the manufacture, model, device type, and software ids.

`------------- MANUFACTUER/MODELS/DEVICETYPE/SOFTWAREIDS -------------`

Any items not recognized will be reported in the extra items section.

`------------- EXTRA ITEMS -------------`

verifyImage will then report if the image is good or has errors.

```
----------------------
******** FILE GOOD ********
----------------------
```

Using the fileName key, it will then check for the existence of the firmware file.
verifyImage will not be able to determine if the file is a valid file, only that the file exists.

`------------- FILE CHECK -------------`

verifyImage will then display the FAS json image which would be used for creating a FAS image record.
This is for informational purposes only, the loader will take care of creating this for FAS.
If you used the `-s3` command line option, that will be used in the FAS json image

`------------- FAS JSON IMAGE -------------`

#### Sample Run

```bash
$verifyImage.py -f sh-svr-5264-gpu-bios-21.03.00-r1.x86_64.rpm

/c/Users/buchmann/Downloads/sh-svr-5264-gpu-bios-21.03.00-r1.x86_64.rpm
rm -rf ./fas_verify_download; mkdir ./fas_verify_download; cd ./fas_verify_download; rpm2cpio /c/Users/buchmann/Downloads/sh-svr-5264-gpu-bios-21.03.00-r1.x86_64.rpm | cpio -idmv
./opt/cray/FW/bios/sh-svr-5264-gpu-bios
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/RBU
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/RBU/MZ92-FS0-YF_C27_F01.json
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/RBU/image.RBU
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/Relnotes_MZ92-FS0-YF_C27_F01.pdf
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/Relnotes_MZ92-FS0-YF_C27_Rome.pdf
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/Relnotes_MZ92-FS0-YF_Naples.pdf
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/SPI_UPD
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/SPI_UPD/AfuEfix64.efi
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/SPI_UPD/flash.nsh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/SPI_UPD/image.bin
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/SPI_UPD/readme.txt
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/f.nsh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/flash_R.rom
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/BMC_Release_Note_128413.doc
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/NOTICE Vertiv Update to AMI BMC Procedure 20200204.pdf
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/doc
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/doc/BMCFirmwareUpdate.txt
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/fw
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/fw/128413.bin
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/fw/128413.json
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/fw/rom.ima_enc
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/dos
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/dos/DC_SELF.EXE
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/dos/DOS4GW.EXE
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/dos/DUMPSEL.EXE
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/dos/FRU.EXE
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/dos/KCS.EXE
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/dos/SENSEL.EXE
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/windows
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/windows/IpmiIo.dll
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/windows/Senselw.exe
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/windows/dc_self.exe
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/windows/ipmilpcdriver32.sys
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/windows/ipmilpcdriver64.sys
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/others/tool/windows/kcsw.exe
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/projects.txt
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/arm-linux
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/arm-linux/NR_flashall.sh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/arm-linux/daul_flashall_chip2.sh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/arm-linux/flash.sh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/arm-linux/flashall.sh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/arm-linux/gigaflash
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/arm-linux/socflash
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux/NR_flashall32.sh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux/NR_flashall64.sh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux/daul_flashall32_chip2.sh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux/daul_flashall64_chip2.sh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux/flash32.sh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux/flash64.sh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux/flashall32.sh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux/flashall64.sh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux/gigaflash
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux/gigaflash_x64
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux/socflash
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/linux/socflash_x64
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/uefi
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/uefi/NR_flashall.nsh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/uefi/daul_flashall_chip2.nsh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/uefi/flash.nsh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/uefi/flashall.nsh
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/uefi/gigaflash.efi
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/uefi/socflash.efi
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/NR_flashall.bat
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/NR_flashall_x64.bat
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/daul_flashall_chip2.bat
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/daul_flashall_x64_chip2.bat
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/flash.bat
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/flash_x64.bat
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/flashall.bat
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/flashall_x64.bat
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/gigaflash.exe
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/gigaflash_x64.exe
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/socflash.exe
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/socflash_x64.exe
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/x64
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/x64/astio.cat
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/x64/astio.sys
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/x64_Win8
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/x64_Win8/astio.Inf
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/x64_Win8/astio.cat
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/x64_Win8/astio.sys
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/x86
./opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/utility/fwud/windows/x86/astio.sys
./opt/cray/tests/sms-destructive/sh-svr-5264-gpu-bios
./opt/cray/tests/sms-functional/sh-svr-5264-gpu-bios
./opt/cray/tests/sms-long/sh-svr-5264-gpu-bios
./opt/cray/tests/sms-resources/bin
./opt/cray/tests/sms-resources/lib
./opt/cray/tests/sms-resources/usr
./opt/cray/tests/sms-smoke/sh-svr-5264-gpu-bios
471537 blocks

CHECKING FILE: MZ92-FS0-YF_C27_F01.json
------------- REQUIRED ITEMS -------------
Required item tags found: [ default ]
Required item firmwareVersion found: C27
Required item semanticFirmwareVersion found: 21.03.00
Required item fileName found: image.RBU
Required item targets found: [ BIOS ]
------------- OPTIONAL ITEMS -------------
Optional item needManualReboot found: True
Optional item pollingSpeedSeconds found: 30
Optional item waitTimeBeforeManualRebootSeconds found: 700
Optional item waitTimeAfterRebootSeconds found: 300
Optional item forceResetType found: ForceRestart
Optional item allowableDeviceStates NOT found, defaulting to: []
Optional item updateURI NOT found, defaulting to: -
Optional item versionURI NOT found, defaulting to: -
------------- MANUFACTUER/MODELS/DEVICETYPE/SOFTWAREIDS -------------
Item manufacturer found: gigabyte
Item models found: [ R282-Z91,R282-Z91-00,R282-Z91-YF,R282-Z93,R282-Z93-00,R282-Z93-YF ]
Item deviceType found: NodeBMC
WARNING: MISSING ITEM: softwareIds :  only required if model is cray
------------- EXTRA ITEMS -------------

----------------------
******** FILE GOOD ********
----------------------

------------- FILE CHECK -------------
*** FILE FOUND :./fas_verify_download/opt/cray/FW/bios/sh-svr-5264-gpu-bios/bios/RBU/image.RBU
------------- FAS JSON IMAGE -------------
{"s3URL": "", "target": "BIOS", "deviceType": "NodeBMC", "manufacturer": "gigabyte", "models": ["R282-Z91", "R282-Z91-00", "R282-Z91-YF", "R282-Z93", "R282-Z93-00", "R282-Z93-YF"], "softwareIds": [], "tags": ["default"], "firmwareVersion": "C27", "semanticFirmwareVersion": "21.03.00", "allowableDeviceStates": [], "needManualReboot": true, "pollingSpeedSeconds": 30, "waitTimeBeforeManualRebootSeconds": 700, "waitTimeAfterRebootSeconds": 300, "forceResetType": "ForceRestart"}
----------------------

CHECKING FILE: 128413.json
------------- REQUIRED ITEMS -------------
Required item tags found: [ default ]
Required item firmwareVersion found: 12.84.13
Required item semanticFirmwareVersion found: 21.03.00
Required item fileName found: rom.ima_enc
Required item targets found: [ BMC ]
------------- OPTIONAL ITEMS -------------
Optional item needManualReboot found: False
Optional item pollingSpeedSeconds found: 30
Optional item allowableDeviceStates NOT found, defaulting to: []
Optional item updateURI NOT found, defaulting to: -
Optional item versionURI NOT found, defaulting to: -
Optional item waitTimeBeforeManualRebootSeconds NOT found, defaulting to: -
Optional item waitTimeAfterRebootSeconds NOT found, defaulting to: -
Optional item forceResetType NOT found, defaulting to: -
------------- MANUFACTUER/MODELS/DEVICETYPE/SOFTWAREIDS -------------
Item manufacturer found: gigabyte
Item models found: [ R282-Z91,R282-Z91-00,R282-Z91-YF,R282-Z93,R282-Z93-00,R282-Z93-YF ]
Item deviceType found: NodeBMC
WARNING: MISSING ITEM: softwareIds :  only required if model is cray
------------- EXTRA ITEMS -------------

----------------------
******** FILE GOOD ********
----------------------

------------- FILE CHECK -------------
*** FILE FOUND :./fas_verify_download/opt/cray/FW/bios/sh-svr-5264-gpu-bios/bmc/fw/rom.ima_enc
------------- FAS JSON IMAGE -------------
{"s3URL": "", "target": "BMC", "deviceType": "NodeBMC", "manufacturer": "gigabyte", "models": ["R282-Z91", "R282-Z91-00", "R282-Z91-YF", "R282-Z93", "R282-Z93-00", "R282-Z93-YF"], "softwareIds": [], "tags": ["default"], "firmwareVersion": "12.84.13", "semanticFirmwareVersion": "21.03.00", "allowableDeviceStates": [], "needManualReboot": false, "pollingSpeedSeconds": 30}
----------------------
```
