// Copyright (c) 2019 Cray Inc. All Rights Reserved.
//
// Except as permitted by contract or express written permission of Cray Inc.,
// no part of this work or its content may be modified, used, reproduced or
// disclosed in any form. Modifications made without express permission of
// Cray Inc. may damage the system the software is installed within, may
// disqualify the user from receiving support from Cray Inc. under support or
// maintenance contracts, or require additional support services outside the
// scope of those contracts to repair the software or system.
//
// This file is contains struct defines for PowerMaps
package sm

import (
	"fmt"
	base "stash.us.cray.com/HMS/hms-base"
)

// An entry mapping an xname to a power supplies
type PowerMap struct {
	ID        string   `json:"id"`
	PoweredBy []string `json:"poweredBy,omitempty"`
}

// This wraps basic PowerMap data with the structure used for query responses.
func NewPowerMap(id string, poweredBy []string) (*PowerMap, error) {
	m := new(PowerMap)
	idNorm := base.VerifyNormalizeCompID(id)
	if idNorm == "" {
		err := fmt.Errorf("xname ID '%s' is invalid", id)
		return nil, err
	}
	m.ID = idNorm
	if len(poweredBy) > 0 {
		for _, pwrId := range poweredBy {
			normPwrID := base.VerifyNormalizeCompID(pwrId)
			if normPwrID == "" {
				err := fmt.Errorf("Power supply xname ID '%s' is invalid", pwrId)
				return nil, err
			} else {
				m.PoweredBy = append(m.PoweredBy, normPwrID)
			}
		}
	}
	return m, nil
}
