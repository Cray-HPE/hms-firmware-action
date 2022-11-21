# Firmware Action Service

The Firmware Action Service (FAS) is the predominant tool used to affect firmware image state changes (upgrade/downgrade) for
Shasta systems. FAS is a RESTful microservice written in Go and maintained by CASMHMS. FAS as a service is deployed inside the
service mesh in the management plane. FAS performs out-of-band (oob) firmware image updates via Redfish.

## FAS Replaces FUS

FAS is the replacement for FUS (Firmware Update Service). FUS was the first implementation, and it has been decided for [many reasons that FUS must be replaced.](docs/Replacing_firmware_update_service.md)

## FAS CT Testing

In addition to the service itself, this repository builds and publishes cray-fas-hmth-test images containing tests that
verify FAS on live Shasta systems. The tests are invoked via helm test as part of the Continuous Test (CT) framework during CSM
installs and upgrades. The version of the cray-fas-hmth-test image (vX.Y.Z) should match the version of the
cray-firmware-action image being tested, both of which are specified in the helm chart for the service.

## Table of Contents

* [FAS v1.0 vs FUS Feature Comparison](docs/Feature_Comparison.md)
* [Scenarios](docs/action_scenarios.md)
* [Control Logic](docs/control_loop.md)
* [Dependency Management](docs/Dependency_Management.md)
* [Developer Environment](docs/developer_environment.md)
* [Test Environment](docs/test_environment.md)
* [Understanding Images](docs/Understanding_Images.md)

