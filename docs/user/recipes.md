## Recipes

The following example JSON files are useful to reference when updating specific hardware components. In all of these examples, the `overrrideDryrun` field will be set to `false`; set them to `true` to perform a live update.

When updating an entire system, walk down the device hierarchy component type by component type, starting first with 'Routers' (switches), proceeding to Chassis, and then finally to Nodes. While this is not strictly necessary, it does help eliminate confusion.

Refer to [FAS Filters for `actions` and `snapshots`](/user/filters.md) for more information on the content used in the example JSON files.

### Manufacturer : Cray

#### Device Type: RouterBMC |  Target: BMC

The BMC on the RouterBMC for a Cray includes the ASIC.  

```json
{
"inventoryHardwareFilter": {
    "manufacturer": "cray"
    },
"stateComponentFilter": {
    "deviceTypes": [
      "routerBMC"
    ]
},
"targetFilter": {
    "targets": [
      "BMC"
    ]
  },
"command": {
    "version": "latest",
    "tag": "default",
    "overrideDryrun": false,
    "restoreNotPossibleOverride": true,
    "timeLimit": 1000,
    "description": "Dryrun upgrade of Columbia and/or Colorado router BMC"
  }
}
```

#### Device Type: ChassisBMC | Target: BMC

**IMPORTANT**: Before updating a CMM, make sure all slot and rectifier power is off.

```json
{
"inventoryHardwareFilter": {
    "manufacturer": "cray"
    },
"stateComponentFilter": {
    "deviceTypes": [
      "chassisBMC"
    ]
},
"targetFilter": {
    "targets": [
      "BMC"
    ]
  },
"command": {
    "version": "latest",
    "tag": "default",
    "overrideDryrun": false,
    "restoreNotPossibleOverride": true,
    "timeLimit": 1000,
    "description": "Dryrun upgrade of Cray Chassis Controllers"
  }
}
```
#### Device Type: NodeBMC | Target: BMC

```json
{
"stateComponentFilter": {

    "deviceTypes": [
      "nodeBMC"
    ]
},
"inventoryHardwareFilter": {
    "manufacturer": "cray"
    },
"targetFilter": {
    "targets": [
      "BMC"
    ]
  },
"command": {
    "version": "latest",
    "tag": "default",
    "overrideDryrun": false,
    "restoreNotPossibleOverride": true,
    "timeLimit": 1000,
    "description": "Dryrun upgrade of Olympus node BMCs"
  }
}
```



#### Device Type: NodeBMC | Target: NodeBIOS

**IMPORTANT**: The Nodes themselves must be powered **off** in order to update the BIOS on the nodes. The BMC will still have power and will perform the update.

**IMPORTANT:** When the BMC is updated or rebooted after updating the Node0.BIOS and/or Node1.BIOS liquid-cooled nodes, the node BIOS version will not report the new version string until the nodes are powered back on. It is recommended that the Node0/1 BIOS be updated in a separate action, either before or after a BMC update and the nodes are powered back on after a BIOS update. The liquid-cooled nodes must be powered off for the BIOS to be updated.

```json
{
"stateComponentFilter": {

    "deviceTypes": [
      "nodeBMC"    ]
  },
"inventoryHardwareFilter": {
    "manufacturer": "cray"
    },
"targetFilter": {
    "targets": [
      "Node0.BIOS",
      "Node1.BIOS"
    ]
  },
"command": {
    "version": "latest",
    "tag": "default",
    "overrideDryrun": false,
    "restoreNotPossibleOverride": true,
    "timeLimit": 1000,
    "description": "Dryrun upgrade of Node BIOS"
  }
}
```

#### Device Type: NodeBMC | Target: Redstone FPGA

**IMPORTANT**: The Nodes themselves must be powered **on** in order to update the firmware of the Redstone FPGA on the nodes.  

```json
{
"stateComponentFilter": {

    "deviceTypes": [
      "nodeBMC"    ]
  },
"inventoryHardwareFilter": {
    "manufacturer": "cray"
    },
"targetFilter": {
    "targets": [
      "Node0.AccFPGA0",
      "Node1.AccFPGA0"
    ]
  },
"command": {
    "version": "latest",
    "tag": "default",
    "overrideDryrun": false,
    "restoreNotPossibleOverride": true,
    "timeLimit": 1000,
    "description": "Dryrun upgrade of Node Redstone FPGA"
  }
}
```


---

### Manufacturer: HPE
#### Device Type: NodeBMC | Target: `iLO 5` aka BMC

```json
"stateComponentFilter": {
    "deviceTypes": [
      "nodeBMC"
    ]
},
"inventoryHardwareFilter": {
    "manufacturer": "hpe"
    },
"targetFilter": {
    "targets": [
      "1"
    ]
  },
"command": {
    "version": "latest",
    "tag": "default",
    "overrideDryrun": false,
    "restoreNotPossibleOverride": true,
    "timeLimit": 1000,
    "description": "Dryrun upgrade of HPE node iLO 5"
  }
}
```

**NOTE**: `1` must be used as the `target` to indicate `iLO 5`. 

#### Device Type: NodeBMC | Target: `System ROM` aka BIOS

```json
{
"stateComponentFilter": {
    "deviceTypes": [
      "NodeBMC"
    ]
},
"inventoryHardwareFilter": {
    "manufacturer": "hpe"
    },
"targetFilter": {
    "targets": [
      "2"
    ]
  },
"command": {
    "version": "latest",
    "tag": "default",
    "overrideDryrun": false,
    "restoreNotPossibleOverride": true,
    "timeLimit": 1000,
    "description": "Dryrun upgrade of HPE node system rom"
  }
}
```

**NOTE**: `2` must be used as the `target` to indicate `System ROM`. 


---

## Manufacturer: Gigabyte

#### Device Type: NodeBMC | Target: BMC

```json
{
"stateComponentFilter": {

    "deviceTypes": [
      "nodeBMC"
    ]
},
"inventoryHardwareFilter": {
    "manufacturer": "gigabyte"
    },
"targetFilter": {
    "targets": [
      "BMC"
    ]
  },
"command": {
    "version": "latest",
    "tag": "default",
    "overrideDryrun": false,
    "restoreNotPossibleOverride": true,
    "timeLimit": 2000,
    "description": "Dryrun upgrade of Gigabyte node BMCs"
  }
}
```

*note*: The timeLimit is `2000` because the gigabytes can take a lot longer to update. 

#### Device Type: NodeBMC | Target: BIOS
```json
{
"stateComponentFilter": {

    "deviceTypes": [
      "nodeBMC"
    ]
},
"inventoryHardwareFilter": {
    "manufacturer": "gigabyte"
    },
"targetFilter": {
    "targets": [
      "BIOS"
    ]
  },
"command": {
    "version": "latest",
    "tag": "default",
    "overrideDryrun": false,
    "restoreNotPossibleOverride": true,
    "timeLimit": 2000,
    "description": "Dryrun upgrade of Gigabyte node BIOS"
  }
}
```

### Update Non-Compute Nodes (NCNs)

NCNs are compute blades. The current NCNs in use are manufactured by Gigabyte or HPE. Use the `NodeBMC` examples in this section updating NCN firmware, and include the `xname` parameter as part of the `stateComponentFilter` to target **ONLY** the xnames that have been separately identified as NCNs.  

Updating more than one NCN at a time **MAY** cause system instability. Be sure to follow the correct process for updating NCN; FAS accepts no responbility for updates that do not follow the correct process.  Firmware updates have the capacity to harm the system; follow the appropriate guides!