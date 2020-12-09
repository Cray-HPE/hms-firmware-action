// Copyright (c) 2020 Cray Inc. All Rights Reserved.
//
// Except as permitted by contract or express written permission of Cray Inc.,
// no part of this work or its content may be modified, used, reproduced or
// disclosed in any form. Modifications made without express permission of
// Cray Inc. may damage the system the software is installed within, may
// disqualify the user from receiving support from Cray Inc. under support or
// maintenance contracts, or require additional support services outside the
// scope of those contracts to repair the software or system.
//
// This file is contains struct defines for CompEthInterfaces

package sm

// This package defines structures for component ethernet interfaces

import (
	base "stash.us.cray.com/HMS/hms-base"
	"strings"
)

//
// Format checking for database keys and query parameters.
//

var ErrCompEthInterfaceBadMAC = base.NewHMSError("sm", "Invalid CompEthInterface MAC Address")
var ErrCompEthInterfaceBadCompID = base.NewHMSError("sm", "Invalid CompEthInterface component ID")


///////////////////////////////////////////////////////////////////////////
//
// CompEthInterface
//
///////////////////////////////////////////////////////////////////////////

// A component ethernet interface is an IP address <-> MAC address relation.
type CompEthInterface struct {
	ID         string `json:"ID"`
	Desc       string `json:"Description"`
	MACAddr    string `json:"MACAddress"`
	IPAddr     string `json:"IPAddress"`
	LastUpdate string `json:"LastUpdate"`
	CompID     string `json:"ComponentID"`
	Type       string `json:"Type"`
}

// Allocate and initialize new CompEthInterface struct, validating it.
func NewCompEthInterface(desc, macAddr, ipAddr, compID string) (*CompEthInterface, error) {
	if macAddr == "" {
		return nil, ErrCompEthInterfaceBadMAC
	}
	cei := new(CompEthInterface)
	cei.Desc = desc
	cei.MACAddr = strings.ToLower(macAddr)
	cei.ID = strings.ReplaceAll(cei.MACAddr, ":", "")
	if cei.ID == "" {
		return nil, ErrCompEthInterfaceBadMAC
	}
	cei.IPAddr = ipAddr
	if compID != "" {
		cei.CompID = base.VerifyNormalizeCompID(compID)
		if cei.CompID == "" {
			return nil, ErrCompEthInterfaceBadCompID
		}
		cei.Type = base.GetHMSTypeString(cei.CompID)
	}
	return cei, nil
}

// Patchable fields if included in payload.
type CompEthInterfacePatch struct {
	Desc   *string `json:"Description"`
	IPAddr *string `json:"IPAddress"`
	CompID *string `json:"ComponentID"`
}