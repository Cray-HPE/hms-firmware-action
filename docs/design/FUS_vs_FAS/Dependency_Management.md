# Dependency Management
## Overview 
The approach used in FUS for Dependency Management is insufficient. It has requirements gaps that weren't factored into the original architecture or design. After a comprehensive analysis we have found many of the gaps, but believe that the complicated nature of dependency management requires extensive requirement validation and a different approach from FUS or from what FAS originally planned. 

The following explains the FUS approach and some of the intricate complications and essential limitations of the domain.  It further recommends a course of action that we should consider.

## FUS Approach
### Overview

Part of the FUS API is `dependencies`.  This API is responsible for storing firmware image data, which includes image data & dependency data. 

<em>abstract representation</em>

```
{
	"deviceType": "nodeBMC",
	"target": "BIOS",
	"semanticVersion": "1.2.0",
	"firmwareVersion": "avb2p.91",
	"manufacturer": "Gigabyte",
	"model": "g200",
	"s3URL": "s3://foo.img",
	"dependencies":[
		{
			"deviceType": "nodeBMC",
			"target": "NIC",
			"reqVersion": "1.4.0"
		}
	 ]
}	
```
 
The image data is VERY important because it is how FUS knows how to match up WHAT firmware is running on a device as well as provides a pointer to the actual image to use to update with. 

In addition this API also remembers 'dependency' relationships between firmware images as a way of 'chaining' operations together.  These are a listing of the 'pre-req' operations that must be completed before we can update the selected target. Dependency management has been asked for because there can be limitations & incompatibilities to what versions of firmware can run on devices that have to interact.

Example:

```
The BMC version must be at 4.0 or greater if the BIOS version is at 1.5
The NIC version cannot be between 12.0 and 15.0 if the NODE0.BMC version is 1.0.2
```

An example use case for performing an update when dependencies are present is: 

```
Perform any prerequisite updates to the device. 
	UPDATE BMC to latest
	UPDATE NIC to latest
Perform an update to the BIOS target of Gigabyte nodeBMC x0c0s1b0 to go to latest.
```

### Limitations
A few essential difficulties:

 1. It is difficult to match hardware models to images; as the image must contain EXACTLY what the device reports.  Device manufacturers can report their model information via different RF API paths (Intel vs Cray vs Gigabyte), and the data that is returned can change from model to model for a specific manufacturer (Gigabyte vs Gigabyte). 
 2. It is hard to make sense of WHAT is on a device target (x0c0s1b0 BIOS); because device manufacturers have NO standard firmware version schema (sem ver for example).  It *could be sem ver, or a date-time, or any 'random' character set; all are valid.  
 3. It is hard to look up 'random' firmware versions against our list of images because of non-standard naming.
 4. We do not have all images available. It is quite possible that the firmware the device is shipped with from the OEM (Intel for example) is not available to us in our repository. This makes restoring a snapshot or performing a rollback operationally impossible.  We do the best we can, but there are TOO many wholes.
 5. Because we don't have ALL possible firmware, and devices report firmware version in non-standard naming is it really hard to determine some sort of ordering (or dependency).  e.g. How do I know what is latest for Gigabyte Node0.BIOS? `MZ62-HD0-YF_C14_F01b_Armor` or `MZ92-FS0-YF_C14_F00a_Armor`? -> YOU DONT. 
 6. The data model is incomplete... Dependencies represent 'abstract' hardware/image relationships.  The data model has a hard time compensating for the lack of firmware information or its lack of cleanliness.   When an actual update is performed those 'dependencies' must be transformed to 'instance' dependencies where we consider the LITERAL state of the device and what firmware versions are reported and identified. 

Furthermore FUS has a limited 'firmware image dependency management' concept. 

* Image dependencies only work for upgrades to 'latest'.  Cannot actually specify go to v1.3.5 (even though I will use specific examples like that).
* FUS Dependency management cannot detect/prevent cycles.  
* FUS can not downgrade with dependency.
* FUS cannot represent dependencies on different devices -> cannot cross the xname barrier (which I think is actually a reasonable limitation, because the logic would get orders of magnitude more difficult and the likelihood of doing something wrong would increase equally. 
 
### Scenarios
The following scenarios help expose the limitations of dependency management in FUS.

#### Cycle Detection - Cross Target

 1. IF gigabyte, g200, nodeBMC: BMC version 1.5 requires BIOS version 2.0.0
 2. IF gigabyte, g200, nodeBMC: BIOS version 2.0.0 requires BIOS version 1.4
 3. RESULT: FUS would run infinitely to perform the update, and would keep creating un-satisfiable operations.
 
#### Cycle Detection - Same Target

 1. IF gigabyte, g200, nodeBMC: BMC version 1.5 requires BMC version 1.4
 2. IF gigabyte, g200, nodeBMC: BMC version 1.4 requires BMC version 1.5
 3. RESULT: FUS would run infinitely to perform the update, and would keep creating un-satisfiable operations.

#### Downgrade Failure

 1. IF gigabyte, g200, nodeBMC: BMC version 1.5 is INCOMPATIBLE with BIOS version < 10.
 2. AND BMC version is at 1.5
 3. AND BIOS version is at 12.
 4. IF BIOS is downgraded to 9
 4. RESULT: BMC version will be untouched, and will likely be in a fault state b/c the BIOS is TOO low for it.

#### Multiple req'd versions

 1. IF gigabyte, g200, nodeBMC: BMC version 1.5 requires BIOS version 2.0.0
 2. AND requires BIOS version 3.1.0
 3. RESULT: FUS will run infinitely flip flopping the BIOS version

#### Snapshot Failure

 1. IF  gigabyte, g200, nodeBMC: BMC is currently at version 1.5 
 1. AND needs rolled back to v1.0 (original)
 2. AND the image isnt present b.c we do not have it.
 3. RESULT: the gigabyte device cannot be rolled back. It is stuck at the current version. 

#### Rollback Failure

 1. IF gigabyte, g200, nodeBMC: BMC version 1.5 requires BIOS version 2.0.0
 2. AND FUS upgrades BIOS to 2.0.0
 3. AND then BMC fails to upgrade to 1.5
 4. RESULT: BIOS stays at version 2.0.0; which it may be incompatible with

 
## FAS Initial Approach
FAS has simplified the `images` API and streamlined its data model.  

The data model is very similar to the FUS approach:

```
{
	"imageID": 1
	"deviceType": "nodeBMC",
	"target": "BIOS",
	"tag": "default"
	"semanticVersion": "1.2.0",
	"firmwareVersion": "avb2p.91",
	"manufacturer": "Gigabyte",
	"model": "g200",
	"s3URL": "s3://foo.img",
	"dependsOn":[2,3,4]
}	
```
Except for addition of TAG, which helps differentiate between two images at the same: deviceType, target, manufacturer, model, & semantic version; but are two different 'releases' : e.g. a TEST image vs a PRODUCTION image.

Additionally: the dependsOn is a lookup ID.

However, this approach to dependency management is plagued with all the same issues that FUS is.

## Recommendation
Given the goal of ensuring that FAS replaces FUS in Shasta v1.3 I think we should remove the concept of 'dependencies' from FAS (keeping images of course). I think dependencies could be added to FAS v1.1; (in the Shasta v1.4 timeline) after extensive investigation and requirements validation.

Frankly, we need to validate the requirements of dependencies and work through the numerous conditions that would cause dependency management to cause undesired behavior in the system. We need to engage our stakeholders and understand their use cases.  The data model itself must be updated to allow for correct dependency relationship mapping.

To date, NO dependency has ever been specified for an image (aka meta-data file) according to our review of the metadata files.  In addition to our knowledge no ticket has ever been created regarding the 'limitations' of FUS dependency management.   In my personal opinion we really need to ascertain if the concept of dependency management is truly a necessary feature for performing firmware actions.  It is not yet clear to me that a sufficient business case exists that warrants its implementation. 

Regarding timelines, I think we should continue FAS development as planned (sans dependencies).  I think we should and can engage stakeholders and work through requirements for dependencies in Q2, but given the relative size of the work that will be needed to implement dependencies we should not include it in MVP for FAS v1.1

### Update - 2020-06-26

It has been decided at the T2 level to not pursue the concept of dependency management for FAS at this time.  Our contractual deliverables must take a higher precedence. After a cursory examination it is not clear that there is sufficient ROI to engage with delivering dependency management. Previously, in the Cascade product, a concept similar to dependency management was available, but was seldom, if ever used.  Frankly admins desire to have greater control over the system. 

We currently have no slated timeline for considering dependency management.   