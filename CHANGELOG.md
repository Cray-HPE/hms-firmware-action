# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.7.8] - 2021-05-04

### Changed

- Updated docker-compose files to pull images from Artifactory instead of DTR.
- Increased timeouts for FAS actions and snapshots CT test cases.

## [1.7.7] - 2021-04-21

### Changed

- Updated Dockerfiles to pull base images from Artifactory instead of DTR.

## [1.7.6] - 2021-04-19

### Changed

- Updated FAS /action CT test to poll for completed action before attempting to delete it.

## [1.7.5] - 2021-04-08

### Changed

- Update to TRS module 1.5.1 for HTTP timeout bug fix

## [1.7.4] - 2021-04-07

### Changed

- Bumped service version to fix release/csm-1.0 build.

## [1.7.3] - 2021-03-23

### Changed

- Added ability to select HPE targets by target name (i.e. "iLO 5", "System ROM")

## [1.7.2] - 2021-02-23

### Changed

- refactored docs

## [1.7.1] - 2021-02-08

### Changed

- Added User-Agent headers to outbound HTTP requests.

## [1.7.0] - 2021-02-05

### Changed

- revendored and updated license on all source code.

## [1.6.2] - 2021-02-04

### Changed

- resolved an issue where xnames did not get unlocked properly

## [1.6.1] - 2021-01-19

### Changed

- Change S3_ENDPOINT from s3_endpoint to fw_s3_endpoint

## [1.6.0] - 2021-01-14

### Changed

- Updated license file.

## [1.5.2] - 2020-11-20

- Fixed the s3_endpoint config to point to the configmap

## [1.5.1] - 2020-11-16

- CASMHMS-4216 - Added final CA bundle configmap handling to Helm chart.


## [1.5.0] - 2020-11-06

### Fixed
- bug in abort sequence that would cause control loop to painc
- removed dependency on s3 (the loader still needs it; but the service does not)
- added extra error handling in doLaunch to fail an operation that gets a 4xx or 5xx status code from the http.do request


## [1.4.6] - 2020-11-06

- Added ability to use TLS certs/CA trust bundles for Redfish communications.

## [1.4.5] - 2020-11-05

### Fixed
- add coping to loader to ignore bad json files
- fix GB payload format

## [1.4.4] - 2020-11-03

## Security

- CASMHMS-4148 - Updated Go module vendor code for security update.

## [1.4.3] - 2020-10-29

## Security
- CASMHMS-4105 - Updated base Golang Alpine image to resolve libcrypto vulnerability.

## [1.4.2] - 2020-10-23

## Added
- CASMHMS-4125: Added support for HPE iLO 5 devices
- Example runs for Gigabyte hardware
- Scripts to aid in setting up Nexus in test environment

## [1.4.1] - 2020-10-22

## Fixed
- Upgraded s3 client; and added logic to check that the s3 file actually exists.

## [1.4.0] - 2020-10-20

## Changed
- Updated FAS to use HSM v2; with new reservation library.

## [1.3.4] - 2020-10-02

## Changed
- Updated cray-service version to 2.0.1

## [1.3.3] - 2020-09-30

## Added
- CASMHMS-4019 - Added support for SoftwareId for images and operations (redfish)

## Changed
- Fixed bug for Model Selection
- Renamed fus_examples.md to fas_examples.md

## [1.3.2] - 2020-09-22

## Changed
- Update nexus loader to check all files in nexus repository

## [1.3.1] - 2020-09-17

## Added
- Overwrite Same Image feature CASMHMS-4007

## Changed
- Updated docker-compose files for vault and smd changes
- Corrected output for error status code

## [1.3.0] - 2020-09-14

## Changed
These are changes to charts in support of:
moving to Helm v1/Loftsman v1
the newest 2.x cray-service base chart
upgraded to support Helm v3
modified containers/init containers, volume, and persistent volume claim value definitions to be objects instead of arrays
the newest 0.2.x cray-jobs base chart
upgraded to support Helm v3
Modifications of your chart values.yaml/requirements.yaml was, in part, automated, so you may see formatting changes or removed whitespace in certain cases since the tools to help us automate would've done this.

Some other info on chart changes:
Starting with Helm v3, helm itself does some rendered resource validation against the k8s api. So, certain charts had invalid properties, invalid values, etc. defined but were being silently ignored by Helm itself. Some changes will be related to removing invalid things that have been in your chart.
All of the changes were a product of testing or known changes in other dependencies.

## [1.2.9] - 2020-09-09

## Changed
- Added additional messages to StateHelper for display on actions output.

## [1.2.8] - 2020-09-01

## Added
- Added overrideImage Flag to actions
- Added code to ping for etcd check and hsm checks
- Increased liveness probe timeout to 5 seconds

## [1.2.7] - 2020-08-26

## Changed
- CASMHMS-3407 - Updated FAS to use trusted baseOS images.
- CASMHMS-2732 - Updated FAS to no longer use deprecated hms-common packages.
- CASMHMS-3695 - Added new CT functional tests for FAS.

## [1.2.6] - 2020-08-05

## Added
- Added operations/{operationID}
- Added actions/{actionID}/operations
- Added actions/{actionID}/status

## Changed
- Increased most timeouts to help with empty return values

## [1.2.5] - 2020-07-24

## Changed
- Increased resource limits for k8s in values.yaml

## [1.2.4] - 2020-07-22

## Changed
- Update to FW-loader script to detect missing files and not crash.

## [1.2.3] - 2020-07-09

## Changed
- CASMHMS-3716: Updated the Dockerfiles to explicitly install the pip package

## [1.2.2] - 2020-07-01

### Added

- CASMHMS-3606 - Added CT smoke test for FAS.

## [1.2.1] - 2020-06-29

### Changed

- converted dryrun to overrideDryrun; which inverts the logic to default to always running a dryrun

## [1.1.1] - 2020-06-29

### Fixed

- turned off SSL verification for fw loader.
- changed state rollup logic

### Added

- added admin documentation

## [1.1.0] - 2020-06-25

### update

- Update for cray base-service chart to 1.11

## [1.0.0] - 2020-06-25

### Added

- Release of v1.0.0 of FAS; includes final changes needed to fw-loader to get it to pull from nexus correctly

## [0.4.0] - 2020-06-25

### Added

- Storage of Image File Data into S3 repository

## [0.3.0] - 2020-06-25

### Added

- enhanced the reporting available for a running action so we can see all possible states.
- expaned the operation summary to include the stateHelper and the fromFirmwareVersion
- updated swagger to indicate these changes
- added more information to swagger to describe possible action/operation states

### Fixed

- fixed inconsistent captialization

## [0.2.0] - 2020-06-24

### Fixed

- snapshot dryrun was not being honored correctly, it now is handled the right way
- We now have a 'specialTargets' map that allows us to get the right model for Node1.BIOS and Node0.BIOS -> CASMHMS-3132
- updated swagger to make tags an array (as it is in code)

## [0.1.4] - 2020-06-18

### Fixed

- several bugs in the update scheduler were removed to ensure the correct headers were used as part of the RF payload
- changed firmware loader to not duplicate into s3 and provide uniqueness to image paths
- added 'automatic' timing and auto retries to doVerify.  By default if an automatic reboot is done we will wait 2 mins
-  then try to verify the version 15 times, with a 30 polling delay.
- fixed a bug in the FSM for operations that prevented the operation from being completed.
- fixed a bug in abort process that would cause the thread of execution to terminate

## [0.1.3] - 2020-06-12

### Fixed

- update path in swagger so that craycli generation works correctly.
- updated documentation to clarify imagefiles
- fixed a path in the loader so that it downloads and parses correctly.


## [0.1.2] - 2020-06-12

### Changed

- This will be the first major release of FAS as part of Shasta v1.3   It is operational.
- We will begin more intentional changelog tracking after this point.
- Converted base chart to 1.8.0-0 so that we could take advantage of the latest ETCD changes

## [0.0.1] - 2020-03-31

### Changed

- updated swagger

### Added

- Added changelog
### Fixed
### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
