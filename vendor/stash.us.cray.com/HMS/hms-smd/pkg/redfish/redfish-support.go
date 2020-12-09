// Copyright (c) 2019 Cray Inc. All Rights Reserved.
package rf

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

/////////////////////////////////////////////////////////////////////////////
// Resource ID
/////////////////////////////////////////////////////////////////////////////

// Get Redfish ID portion of URI, i.e. the basename of the string
func (r *ResourceID) Basename() string {
	return path.Base(r.Oid)
}

// For sort interface, wraps []ResourceID
type ResourceIDSlice []ResourceID

// For sort interface, length function
func (rs ResourceIDSlice) Len() int {
	return len(rs)
}

// For sort interface, comparison operation
func (rs ResourceIDSlice) Less(i, j int) bool {
	cmp := strings.Compare(rs[i].Oid, rs[j].Oid)
	if cmp < 0 {
		return true
	}
	return false
}

// For sort interface, swaps at indexes
func (rs ResourceIDSlice) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

////////////////////////////////////////////////////////////////////////////
// Encoding and decoding
////////////////////////////////////////////////////////////////////////////

// Decode Event and return newly allocated pointer, or else nil/error
// Error contains loggable details.
func EventDecode(raw []byte) (*Event, error) {
	e := new(Event)
	err := json.Unmarshal(raw, e)
	if err != nil {
		fullErr := fmt.Errorf("error '%s' decoding Event '%s'", err, raw)
		return nil, fullErr
	}
	return e, nil
}

// Parse the MessageId field of an eventRecord within an event.
//
// If erec.MessageId: "Registry.1.0.MessageId" (or 1.0.0),
// Return: 'registry'="Registry", 'version'="1.0", and 'msgid'="MessageId".
//
// If only two fields, 'version' will be "", if only one, 'registry' will
// be empty as well and it will be returned as 'msgid'.
func EventRecordMsgId(erec *EventRecord) (registry, version, msgid string) {
	fields := strings.Split(strings.TrimSpace(erec.MessageId), ".")
	num := len(fields)
	if num <= 0 {
		return
	} else if num == 1 {
		msgid = fields[0]
		return
	} else {
		registry = fields[0]
		msgid = fields[num-1]
	}
	// The middle fields will be the version
	for i := 1; i < num-1; i++ {
		if i > 1 {
			version += "."
		}
		version += fields[i]
	}
	return
}

// For a version string, e.g. 1.0.0, include only the first 'num' parts.
// Return 'version' with included fields only, plus the number of fields
// 'included'.  If less are found, 'included' will be less than 'num'.
//
// Example: For vers=1.0.0 and num=2, return "1.0", 2.
func VersionFields(vers, delim string, num int) (version string, included int) {
	// Parse the MessageId field of an eventRecord within an event.
	fields := strings.Split(strings.TrimSpace(vers), delim)
	i := 0
	for ; i < len(fields) && i < num; i++ {
		if i > 0 {
			version += delim
		}
		version += fields[i]
	}
	included = i
	return
}
