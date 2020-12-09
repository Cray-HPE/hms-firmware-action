// Copyright 2020 Hewlett Packard Enterprise Development LP

// This file is contains struct defines for HSN NICs

package sm

// This package defines structures for HSN interfaces

///////////////////////////////////////////////////////////////////////////
//
// HSNInterface
//
///////////////////////////////////////////////////////////////////////////

type HSNNICLocationInfo struct {
	ID             string `json:"ID"`
	GroupID        int    `json:"GroupID,omitempty"`
	PortNum        int    `json:"PortNum,omitempty"`
	PortDesignator string `json:"PortDesignator,omitempty"`
	SwitchNum      int    `json:"SwitchNum,omitempty"`
}

type HSNNICFRUInfo struct {
	Manufacturer string `json:"Manufacturer,omitempty"`
	PartNumber   string `json:"PartNumber,omitempty"`
	SerialNumber string `json:"SerialNumber"`
}
