// Copyright 2020 Cray Inc. All Rights Reserved.
//
// Except as permitted by contract or express written permission of Cray Inc.,
// no part of this work or its content may be modified, used, reproduced or
// disclosed in any form. Modifications made without express permission of
// Cray Inc. may damage the system the software is installed within, may
// disqualify the user from receiving support from Cray Inc. under support or
// maintenance contracts, or require additional support services outside the
// scope of those contracts to repair the software or system.

package sm

import (
	"strings"
	base "stash.us.cray.com/HMS/hms-base"
)

var ErrHWHistEventTypeInvalid = base.NewHMSError("sm", "Invalid hardware inventory history event type")
var ErrHWInvHistFmtInvalid = base.NewHMSError("sm", "Invalid HW Inventory History format")

// Valid values for event types
const (
	HWInvHistEventTypeAdded   = "Added"
	HWInvHistEventTypeRemoved = "Removed"
	HWInvHistEventTypeScanned = "Scanned"
)

// For case-insensitive verification and normalization of state strings
var hwInvHistEventTypeMap = map[string]string{
	"added":   HWInvHistEventTypeAdded,
	"removed": HWInvHistEventTypeRemoved,
	"scanned": HWInvHistEventTypeScanned,
}

type HWInvHistFmt int

const (
	HWInvHistFmtByLoc HWInvHistFmt = iota
	HWInvHistFmtByFRU
)

type HWInvHist struct {
	ID        string `json:"ID"`        // xname location where the event happened
	FruId     string `json:"FRUID"`     // FRU ID of the affected FRU
	Timestamp string `json:"Timestamp"` // Timestamp of the event
	EventType string `json:"EventType"` // (i.e. Added, Removed, Scanned)
}

type HWInvHistArray struct {
	ID      string       `json:"ID"`      // xname or FruId (if ByFRU)
	History []*HWInvHist `json:"History"`
}

type HWInvHistResp struct {
	Components []HWInvHistArray `json:"Components"`
}

// Create formatted HWInvHistResp from a random array of HWInvHist entries.
// No sorting is done (with components of the same type), so pre/post-sort if
// needed.
func NewHWInvHistResp(hwHists []*HWInvHist, format HWInvHistFmt) (*HWInvHistResp, error) {
	compHist := new(HWInvHistResp)
	compHistMap := make(map[string]int)
	if !(format == HWInvHistFmtByLoc || format == HWInvHistFmtByFRU) {
		return nil, ErrHWInvHistFmtInvalid
	}
	var idx int
	for _, hwHist := range hwHists {
		id := hwHist.ID
		if format == HWInvHistFmtByFRU {
			id = hwHist.FruId
		}
		index, ok := compHistMap[id]
		if !ok {
			compHistMap[id] = idx
			idx++
			compHistArray := HWInvHistArray{
				ID: id,
				History: []*HWInvHist{hwHist},
			}
			compHist.Components = append(compHist.Components, compHistArray)
		} else {
			compHist.Components[index].History = append(compHist.Components[index].History, hwHist)
		}
	}
	return compHist, nil
}

// Validate and Normalize event types used in queries
func VerifyNormalizeHWInvHistEventType(eventType string) string {
	evtLower := strings.ToLower(eventType)
	value, ok := hwInvHistEventTypeMap[evtLower]
	if !ok {
		return ""
	} else {
		return value
	}
}
