// Copyright (c) 2018 Cray Inc. All Rights Reserved.
package sm

import (
	base "stash.us.cray.com/HMS/hms-base"
)

type SMEventType string

const (
	NodeStateChange       SMEventType = "NodeStateChange"
	StateChange           SMEventType = "StateChange"
	RedfishEndpointChange SMEventType = "RedfishEndpointChange"
	HWInventoryChange     SMEventType = "HWInventoryChange"
)

type SMEventSubtype string

const (
	StateTransitionOK       SMEventSubtype = "StateTransitionOK"       // StateChange - Successful state change
	StateTransitionAbnormal SMEventSubtype = "StateTransitionAbnormal" // StateChange - Change due to problem, warn/alert
	StateTransitionDisable  SMEventSubtype = "StateTransitionDisable"  // StateChange
	StateTransitionEnable   SMEventSubtype = "StateTransitionEnable"   // StateChange
	NodeAvailable           SMEventSubtype = "NodeAvailable"           // NodeStateChange
	NodeUnavailable         SMEventSubtype = "NodeUnavailable"         // NodeStateChange
	NodeFailed              SMEventSubtype = "NodeFailed"              // NodeStateChange
	NodeStandby             SMEventSubtype = "NodeStandby"             // NodeStateChange
	NodeRoleChanged         SMEventSubtype = "NodeRoleChanged"         // NodeStateChange
	NodeSubRoleChanged      SMEventSubtype = "NodeSubRoleChanged"      // NodeStateChange
	NodeNIDChanged          SMEventSubtype = "NodeNIDChanged"          // NodeStateChange
	RedfishEndpointAdded    SMEventSubtype = "RedfishEndpointAdded"    // RedfishEndpointChange
	RedfishEndpointModified SMEventSubtype = "RedfishEndpointModified" // RedfishEndpointChange
	RedfishEndpointEnabled  SMEventSubtype = "RedfishEndpointEnabled"  // RedfishEndpointChange
	RedfishEndpointDisabled SMEventSubtype = "RedfishEndpointDisabled" // RedfishEndpointChange
	RedfishEndpointRemoved  SMEventSubtype = "RedfishEndpointRemoved"  // RedfishEndpointChange
	HWInventoryAdded        SMEventSubtype = "HWInventoryAdded"        // HWInventoryChange
	HWInventoryModifed      SMEventSubtype = "HWInventoryModified"     // HWInventoryChange
	HWInventoryRemoved      SMEventSubtype = "HWInventoryRemoved"      // HWInventoryChange
)

type SMEvent struct {
	EventType    string `json:"EventType"`
	EventSubtype string `json:"EventSubtype"`

	// At least one of, as per event type:
	ComponentArray       *base.ComponentArray  `json:"ComponentArray,omitempty"`
	HWInventory          *SystemHWInventory    `json:"HWInventory,omitempty"`
	RedfishEndpointArray *RedfishEndpointArray `json:"RedfishEndpointArray,omitempty"`
}

type SMEventArray struct {
	Name      string     `json:"Name,omitempty"`
	Version   string     `json:"Version"`
	Timestamp string     `json:"Timestamp"`
	Events    []*SMEvent `json:"Events"`
}
