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
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
)

type Snapshots struct {
	Snapshots []Snapshot `json:"snapshots"`
}

type Snapshot struct {
	Name           string             `json:"name"`
	CaptureTime    sql.NullTime       `json:"captureTime"`
	ExpirationTime sql.NullTime       `json:"expirationTime"`
	Ready          bool               `json:"ready"`
	Devices        []Device           `json:"devices"`
	RelatedActions []uuid.UUID        `json:"relatedActions"`
	Parameters     SnapshotParameters `json:"parameters"`
}

type SnapshotStorable struct {
	Name           string                     `json:"name,omitempty"`
	CaptureTime    sql.NullTime               `json:"captureTime,omitempty"`
	ExpirationTime sql.NullTime               `json:"expirationTime,omitempty"`
	Ready          bool                       `json:"ready,omitempty"`
	Devices        []DeviceStorable           `json:"devices,omitempty"`
	RelatedActions []uuid.UUID                `json:"relatedActions,omitempty"`
	Parameters     SnapshotParametersStorable `json:"parameters,omitempty"`
}

func ToSnapshotStorable(from Snapshot) (to SnapshotStorable) {
	to = SnapshotStorable{
		Name:           from.Name,
		CaptureTime:    from.CaptureTime,
		ExpirationTime: from.ExpirationTime,
		Ready:          from.Ready,
		RelatedActions: from.RelatedActions,
		Parameters:     ToSnapshotParametersStorable(from.Parameters),
	}
	devs := ToDeviceStorable(from.Devices)
	to.Devices = append(to.Devices, devs...)
	return to
}

func ToSnapshotFromStorable(from SnapshotStorable) (to Snapshot) {
	to = Snapshot{
		Name:           from.Name,
		Ready:          from.Ready,
		CaptureTime:    from.CaptureTime,
		ExpirationTime: from.ExpirationTime,
		RelatedActions: from.RelatedActions,
		Parameters:     ToSnapshotParametersFromStorable(from.Parameters),
	}
	devs := ToDeviceFromStorable(from.Devices)
	to.Devices = append(to.Devices, devs...)

	return to
}

//TODO consider adding make/model/device type to this! REFEREENCE: CASMHMS-2781

//TODo add Error into the swagger!
type Device struct {
	Xname   string   `json:"xname,omitempty"`
	Targets []Target `json:"targets,omitempty"`
	Error   error    `json:"error,omitempty"`
}

//TODo add Error into the swagger!
type Target struct {
	Name            string    `json:"name,omitempty"`
	FirmwareVersion string    `json:"firmwareVersion,omitempty"`
	SoftwareId      string    `json:"softwareId,omitempty"`
	TargetName      string    `json:"targetName"`
	Error           error     `json:"error,omitempty"`
	ImageID         uuid.UUID `json:"imageID,omitempty"`
}

type DeviceStorable struct {
	Xname   string           `json:"xname,omitempty"`
	Targets []TargetStorable `json:"targets,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type TargetStorable struct {
	Name            string `json:"name,omitempty"`
	FirmwareVersion string `json:"firmwareVersion,omitempty"`
	Error           string `json:"error,omitempty"`
	ImageID         string `json:"imageID,omitempty"`
	SoftwareId      string `json:"softwareId"`
	TargetName      string `json:"targetName"`
}

func ToDeviceStorable(from []Device) (to []DeviceStorable) {
	for _, f := range from {
		DS := DeviceStorable{
			Xname: f.Xname,
		}
		if f.Error != nil {
			DS.Error = f.Error.Error()
		}
		TS := ToTargetStorable(f.Targets)
		DS.Targets = append(DS.Targets, TS...)
		to = append(to, DS)
	}
	return to
}

func ToDeviceFromStorable(from []DeviceStorable) (to []Device) {
	for _, f := range from {
		D := Device{
			Xname: f.Xname,
		}
		if f.Error != "" {
			D.Error = errors.New(f.Error)
		}
		T := ToTargetFromStorable(f.Targets)
		D.Targets = append(D.Targets, T...)
		to = append(to, D)
	}
	return to
}

func ToTargetStorable(from []Target) (to []TargetStorable) {
	for _, f := range from {
		T := TargetStorable{
			Name:            f.Name,
			FirmwareVersion: f.FirmwareVersion,
			ImageID:         f.ImageID.String(),
			TargetName:      f.TargetName,
			SoftwareId:      f.SoftwareId,
		}
		if f.Error != nil {
			T.Error = f.Error.Error()
		}
		to = append(to, T)
	}
	return to
}

func ToTargetFromStorable(from []TargetStorable) (to []Target) {
	for _, f := range from {
		T := Target{
			Name:            f.Name,
			FirmwareVersion: f.FirmwareVersion,
			TargetName:      f.TargetName,
			SoftwareId:      f.SoftwareId,
		}
		if f.Error != "" {
			T.Error = errors.New(f.Error)
		}
		T.ImageID, _ = uuid.Parse(f.ImageID)
		to = append(to, T)
	}
	return to
}

type SnapshotParameters struct {
	Name                    string                  `json:"name"`
	ExpirationTime          sql.NullTime            `json:"expirationTime,omitempty"`
	StateComponentFilter    StateComponentFilter    `json:"stateComponentFilters,omitempty"`
	InventoryHardwareFilter InventoryHardwareFilter `json:"inventoryHardwareFilters,omitempty"`
	TargetFilter            TargetFilter            `json:"targetFilter,omitempty"`
}

type SnapshotParametersStorable struct {
	Name                    string                  `json:"name"`
	ExpirationTime          sql.NullTime            `json:"expirationTime,omitempty"`
	StateComponentFilter    StateComponentFilter    `json:"stateComponentFilters,omitempty"`
	InventoryHardwareFilter InventoryHardwareFilter `json:"inventoryHardwareFilters,omitempty"`
	TargetFilter            TargetFilter            `json:"targetFilter,omitempty"`
}

func ToSnapshotParametersStorable(from SnapshotParameters) (to SnapshotParametersStorable) {
	to = SnapshotParametersStorable{
		Name:                    from.Name,
		ExpirationTime:          from.ExpirationTime,
		StateComponentFilter:    from.StateComponentFilter,
		InventoryHardwareFilter: from.InventoryHardwareFilter,
		TargetFilter:            from.TargetFilter,
	}
	return to
}

func ToSnapshotParametersFromStorable(from SnapshotParametersStorable) (to SnapshotParameters) {
	to = SnapshotParameters{
		Name:                    from.Name,
		ExpirationTime:          from.ExpirationTime,
		StateComponentFilter:    from.StateComponentFilter,
		InventoryHardwareFilter: from.InventoryHardwareFilter,
		TargetFilter:            from.TargetFilter,
	}
	return to
}

func (obj *SnapshotParameters) Equals(other SnapshotParameters) bool {
	if obj.ExpirationTime != other.ExpirationTime {

		if obj.ExpirationTime.Time.Round(0).Equal(other.ExpirationTime.Time.Round(0)) == false {
			logrus.Warn("ExpirationTime not equal")
			logrus.Info(obj.ExpirationTime.Time.Sub(other.ExpirationTime.Time).Seconds())
			return false
		}
	} else if obj.TargetFilter.Equals(other.TargetFilter) == false {
		logrus.Warn("TargetFilter not equal")
		return false

	} else if obj.InventoryHardwareFilter.Equals(other.InventoryHardwareFilter) == false {
		logrus.Warn("InventoryHardwareFilter not equal")
		return false

	} else if obj.StateComponentFilter.Equals(other.StateComponentFilter) == false {
		logrus.Warn("StateComponentFilter not equal")
		return false
	}
	return true
}

func (obj *Snapshot) Equals(other Snapshot) bool {

	if obj.Name != other.Name {
		logrus.Warn("Name not equal")
		return false
	} else if obj.CaptureTime != other.CaptureTime {

		if obj.CaptureTime.Time.Round(0).Equal(other.CaptureTime.Time.Round(0)) == false {
			logrus.Warn("CaptureTime not equal")
			logrus.Info(obj.CaptureTime.Time.Sub(other.CaptureTime.Time).Seconds())
			return false
		}

	} else if obj.ExpirationTime != other.ExpirationTime {

		if obj.ExpirationTime.Time.Round(0).Equal(other.ExpirationTime.Time.Round(0)) == false {
			logrus.Warn("ExpirationTime not equal")
			logrus.Info(obj.ExpirationTime.Time.Sub(other.ExpirationTime.Time).Seconds())
			return false
		}
	} else if model.UUIDSliceEquals(obj.RelatedActions, other.RelatedActions) == false {
		logrus.Warn("RelatedActions not equal")
		return false
	} else if obj.Parameters.Equals(other.Parameters) == false {
		logrus.Warn("Parameters not equal")
	} else if obj.Ready != other.Ready {
		logrus.Warn("Ready not equal")
		return false
	} else {

		objMap := make(map[string]Device)
		for _, v := range obj.Devices {
			objMap[v.Xname] = v
		}

		otherMap := make(map[string]Device)
		for _, v := range other.Devices {
			otherMap[v.Xname] = v
		}
		for k, v := range objMap {
			if sub, ok := otherMap[k]; ok {
				equals := v.Equals(sub)
				if equals == false {
					return false
				}
			} else {

				return false
			}
		}
	}

	return true
}

func (obj *Device) Equals(other Device) (equals bool) {
	equals = false

	if obj.Xname == other.Xname &&
		len(obj.Targets) == len(other.Targets) {

		if len(obj.Targets) == 0 {
			equals = true
			return
		} else {
			objMap := make(map[string]Target)
			for _, v := range obj.Targets {
				objMap[v.Name] = v
			}

			otherMap := make(map[string]Target)
			for _, v := range other.Targets {
				otherMap[v.Name] = v
			}

			for k, v := range objMap {
				if sub, ok := otherMap[k]; ok {
					equals = v.Equals(sub)
					if equals == false {
						return
					}
				} else {
					equals = false
					return
				}
			}
		}
	}
	return
}

func (obj *Target) Equals(other Target) (equals bool) {
	equals = false
	if obj.Name == other.Name &&
		obj.FirmwareVersion == other.FirmwareVersion {
		equals = true
	}
	return
}
