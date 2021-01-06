# Scenarios for Firmware Action

After reviewing the API, there are only a handful of requests that will cause a firmware action to occur.  In dissecting those events we have fleshed out the various outcomes for each scenario.  From this we have identified a set of questions that we will need to have answered so we can implement the right thing for FAS. 

The predominant way a firmware update is requested is by a client request to `/actions`. The request may choose to blindly perform an action on the entire system or they may filter the selection of devices by several parameters.

## Perform a Firmware Action

1. ask for 10 unique xname/targets; 10 are in HSM; all 10 are able to be updated. Call received actionID
1. ask for 10 unique xname/targets; 2 are INVALID xnames, eg: `c100d0f2`, NO operations are performed, No action created.  Caller receives error.
1. ask for 10 unique xname/targets; 10 are in HSM; 2 are in the wrong power state.  FAS does what it can, and waits for the expiration time, an order to abort, or the hardware to become available.  Caller receives actionID.  NOTE: wait may be infinite. 

The other way a firmware update is triggered is by attempting to restore a snapshot. 

## Restore a Snapshot
Scenario for snapshot: Imagine we took the snapshot on Monday, and try to restore on Friday. The size of the system is 50 devices; with 2 targets each.

1. snapshot on Monday includes ALL 50 device. The same set of hardware on Monday is available on Friday; so everything gets restored.
2. snapshot on Monday includes ALL 50 devices.  However by Friday some devices have been removed from the system.  For the remaining devices operations will be created; and those devices will be restored (if possible).
3. snapshot on Monday includes ALL 50 devices.  However by Friday some devices have been added to the system.  For all original devices the operations are created and proceed.  For the new devices no action is taken, no notification is given to report the presence of new hardware. 


## Questions pulled from scenarios:
Query: what do we do with things that can NEVER finish? say the device was taken away, never to be seen again?  Auto expiration, or manual cancel? 

Query: if the system is big (10,000 nodes); and we do an update all; but only 9,995 can be done, how do we inform users that 5 nodes didn't get updated.