/*
 * MIT License
 *
 * (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

package storage

import (
	"github.com/sirupsen/logrus"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
)

//StateComponentFilter -> queries hsm/state/components -> lockable components that were discovered
type StateComponentFilter struct {
	Partitions  []string `json:"partitions,omitempty"`
	Groups      []string `json:"groups,omitempty"`
	Xnames      []string `json:"xnames,omitempty"`
	DeviceTypes []string `json:"deviceTypes,omitempty"` //nodeBMC
}

func (obj *StateComponentFilter) Equals(other StateComponentFilter) bool {
	if model.StringSliceEquals(obj.Xnames, other.Xnames)  == false {
		logrus.Warn("Xnames not equal")
		return false
	} else if model.StringSliceEquals(obj.Partitions, other.Partitions) == false {
		logrus.Warn("Partitions not equal")
		return false
	} else if model.StringSliceEquals(obj.Groups, other.Groups) == false {
		logrus.Warn("Groups not equal")
		return false
	}else if model.StringSliceEquals(obj.DeviceTypes, other.DeviceTypes) == false {
		logrus.Warn("DeviceTypes not equal")
		return false
	}
	return true
}

// InventoryHardwareFilter -> queries hsm/Inventory/hardware for manufacturer/model that is in the xnameset
type InventoryHardwareFilter struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
}

func (obj *InventoryHardwareFilter) Equals(other InventoryHardwareFilter) bool {
	if obj.Manufacturer == other.Manufacturer &&
		obj.Model == other.Model {
		return true
	}
	return false
}

func (obj *InventoryHardwareFilter) Empty() bool {
	if obj.Model == "" && obj.Manufacturer == "" {
		return true
	}
	return false
}

//TargetFilter -> nodeBMC, routerBMC, etc.
type TargetFilter struct {
	Targets []string `json:"targets,omitempty"`
}

func (obj *TargetFilter) Equals(other TargetFilter) bool {
	if !(model.StringSliceEquals(obj.Targets, other.Targets)) {
		return false
	}
	return true
}
