# Errors from FAS
Starting with FAS version 1.9.4 FAS will return a list of errors with each action, snapshot, and  operation.

This is a list of possible errors and how to correct:

### "No Valid XNames Found in HSM State Components"
### "No Viable Targets"
FAS could not find any components to create operations.
* Check Hardware State Manager (HSM) for components.
* Check action json file for valid targets

### "Error retrieving data from *xname*"
FAS could not retrieve data from xname.
* Check that xname is powered up and reachable

### "*xname* discovery status: *DiscoveryStatus*",
If a node has not yet been completely discovered or has discover issues, FAS can not update that node.
* Check HSM for errors discovering the node.
* Once node discover status is "DiscoverOK" FAS can update the node.
