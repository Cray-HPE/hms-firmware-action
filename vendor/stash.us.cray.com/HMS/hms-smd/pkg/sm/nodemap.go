// Copyright (c) 2018 Cray Inc. All Rights Reserved.
package sm

import (
	"encoding/json"
	"fmt"
	base "stash.us.cray.com/HMS/hms-base"
)

// An entry mapping a node xname to a NID
type NodeMap struct {
	ID       string           `json:"ID"`
	NID      int              `json:"NID"`
	Role     string           `json:"Role,omitempty"`
	SubRole  string           `json:"SubRole,omitempty"`
	NodeInfo *json.RawMessage `json:"NodeInfo,omitempty"`
}

// Named array of NodeMap entries, for representing a collection of
// them.
type NodeMapArray struct {
	NodeMaps []*NodeMap `json:"NodeMaps"`
}

// This wraps basic RedfishEndpointDescription data with the structure
// used for query responses.
func NewNodeMap(id, role, subRole string, nid int, nodeInfo *json.RawMessage) (*NodeMap, error) {
	m := new(NodeMap)
	idNorm := base.NormalizeHMSCompID(id)
	if base.GetHMSType(idNorm) != base.Node {
		err := fmt.Errorf("xname ID '%s' is invalid or not a node", id)
		return nil, err
	}
	m.ID = idNorm
	if nid > 0 {
		m.NID = nid
	} else {
		err := fmt.Errorf("NID '%d' is out of range", nid)
		return nil, err
	}
	if role != "" {
		normRole := base.VerifyNormalizeRole(role)
		if normRole != "" {
			m.Role = normRole
		} else {
			err := fmt.Errorf("Role '%s' is not valid.", role)
			return nil, err
		}
	}
	if subRole != "" {
		normSubRole := base.VerifyNormalizeSubRole(subRole)
		if normSubRole != "" {
			m.SubRole = normSubRole
		} else {
			err := fmt.Errorf("SubRole '%s' is not valid.", subRole)
			return nil, err
		}
	}
	if nodeInfo != nil {
		m.NodeInfo = nodeInfo
	}
	return m, nil
}
