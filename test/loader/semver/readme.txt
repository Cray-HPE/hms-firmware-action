Tests the semantic versioning of the loader

To build test zip file:
$ make

To run test on loader:
$ cray fas loader create --file testsemver.zip
loaderRunID = "a93f2327-8e0e-4569-b447-0cdf1a6c1f52"

Check the loader output:
$ cray fas loader describe {loaderRunID}

Verify the loader worked:
You will get a list of version numbers, check the firmwareVersion is the same
as the semanticFirmwareVersion:

$ cray fas images list | grep Version
firmwareVersion = "20.3.4-24"
semanticFirmwareVersion = "20.3.4-24"
firmwareVersion = "20.3.0-11"
semanticFirmwareVersion = "20.3.0-11"
firmwareVersion = "20.0.0"
semanticFirmwareVersion = "20.0.0"
firmwareVersion = "20.3.5"
semanticFirmwareVersion = "20.3.5"
firmwareVersion = "20.1.0"
semanticFirmwareVersion = "20.1.0"

Note if you have more then just the test images in FAS, you can filter with the
command:
$ cray fas images list --format json | jq '.images | .[] | select(.manufacturer | contains("test"))' | grep Version
