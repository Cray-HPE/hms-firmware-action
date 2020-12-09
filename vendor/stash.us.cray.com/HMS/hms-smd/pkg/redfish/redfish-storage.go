// Copyright (c) 2020 Cray Inc. All Rights Reserved.
package rf

import (
	"encoding/json"
)

// JSON decoded struct returned from Redfish for a particular set of
// ids.  Storage Collection resources deviate from GenericCollections by
// by using the Drives array and Drives count fields instead of a Members array
// and Members count fields.
// Example: /redfish/v1/Systems/<system_id>/Storage/<storage_collection_id>
type StorageCollection struct {
	OContext     string       `json:"@odata.context"`
	OCount       int          `json:"@odata.count"` // Oldest schemas use
	Oid          string       `json:"@odata.id"`
	Otype        string       `json:"@odata.type"`
	Description  string       `json:"Description"`
	Name         string       `json:"Name"`
	Drives       []ResourceID `json:"Drives"`
	DrivesOCount int          `json:"Drives@odata.count"` // Most schemas
}

// Redfish pass-through from Redfish "Drive"
// This is the set of Redfish fields for this object that HMS understands
// and/or finds useful.  Those assigned to either the *LocationInfo
// or *FRUInfo subfields constitute the type specific fields in the
// HWInventory objects that are returned in response to queries.
type Drive struct {
	OContext string `json:"@odata.context"`
	Oid      string `json:"@odata.id"`
	Otype    string `json:"@odata.type"`

	// Embedded structs.
	DriveLocationInfoRF
	DriveFRUInfoRF

	Status StatusRF `json:"Status"`
}

// Location-specific Redfish properties to be stored in hardware inventory
// These are only relevant to the currently installed location of the FRU
// TODO: How to version these (as HMS structures).
type DriveLocationInfoRF struct {
	// Redfish pass-through from rf.Drive
	Id          string `json:"Id"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

// Durable Redfish properties to be stored in hardware inventory as
// a specific FRU, which is then link with it's current location
// i.e. an x-name.  These properties should follow the hardware and
// allow it to be tracked even when it is removed from the system.
// TODO: How to version these (as HMS structures)
type DriveFRUInfoRF struct {
	// Redfish pass-through from rf.Drive

	//Manufacture Info
	Manufacturer string `json:"Manufacturer"`
	SerialNumber string `json:"SerialNumber"`
	PartNumber   string `json:"PartNumber"`
	Model        string `json:"Model"`
	SKU          string `json:"SKU"`

	//Capabilities Info
	CapacityBytes    json.Number `json:"CapacityBytes"`
	Protocol         string      `json:"Protocol"`
	MediaType        string      `json:"MediaType"`
	RotationSpeedRPM json.Number `json:"RotationSpeedRPM"`
	BlockSizeBytes   json.Number `json:"BlockSizeBytes"`
	CapableSpeedGbs  json.Number `json:"CapableSpeedGbs"`

	//Status Info
	FailurePredicted              bool        `json:"FailurePredicted"`
	EncryptionAbility             string      `json:"EncryptionAbility"`
	EncryptionStatus              string      `json:"EncryptionStatus"`
	NegotiatedSpeedGbs            json.Number `json:"NegotiatedSpeedGbs"`
	PredictedMediaLifeLeftPercent json.Number `json:"PredictedMediaLifeLeftPercent"`
}
