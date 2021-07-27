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

//TODO needs unit testing!

package presentation

import (
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
)

type SnapshotName struct {
	Name string `json:"name"`
}

type SnapshotSummary struct {
	Name              string          `json:"name"`
	CaptureTime       string          `json:"captureTime"`
	ExpirationTime    string          `json:"expirationTime,omitempty"`
	Ready             bool            `json:"ready"`
	RelatedActions    []RelatedAction `json:"relatedActions"`
	UniqueDeviceCount int             `json:"uniqueDeviceCount"`
}

type SnapshotSummaries struct {
	Summaries []SnapshotSummary `json:"snapshots"`
}

type RelatedAction struct {
	ActionID  uuid.UUID `json:"actionID"`
	StartTime string    `json:"startTime,omitempty"`
	EndTime   string    `json:"endTime,omitempty"`
	State     string    `json:"state,omitempty"`
}

func (a RelatedAction) Equals(sub RelatedAction) bool {
	if a.ActionID == sub.ActionID &&
		a.EndTime == sub.EndTime &&
		a.StartTime == sub.StartTime &&
		a.State == sub.State {
		return true
	}
	return false
}

type SnapshotMarshaled struct {
	Name           string                      `json:"name"`
	CaptureTime    string                      `json:"captureTime"`
	ExpirationTime string                      `json:"expirationTime,omitempty"`
	Ready          bool                        `json:"ready"`
	Devices        []storage.Device            `json:"devices"`
	RelatedActions []RelatedAction             `json:"relatedActions"`
	Parameters     SnapshotParametersMarshaled `json:"parameters"`
	Errors         []string                    `json:"errors"`
}

func (obj *SnapshotMarshaled) Equals(other SnapshotMarshaled) (equals bool) {
	equals = false

	if obj.Name == other.Name &&
		obj.CaptureTime == other.CaptureTime &&
		obj.ExpirationTime == other.ExpirationTime &&
		//obj.Ready == other.Ready && -- Removed ready from check
		obj.Parameters.Equals(other.Parameters) &&
		len(obj.Devices) == len(other.Devices) &&
		len(obj.RelatedActions) == len(other.RelatedActions) {

		if len(obj.Devices) > 0 {

			objMap := make(map[string]storage.Device)
			for _, v := range obj.Devices {
				objMap[v.Xname] = v
			}

			otherMap := make(map[string]storage.Device)
			for _, v := range other.Devices {
				otherMap[v.Xname] = v
			}

			for k, v := range objMap {
				if sub, ok := otherMap[k]; ok {
					equals = v.Equals(sub)
				} else {
					return
				}
			}

		}

		if len(obj.RelatedActions) > 0 {

			objMap := make(map[uuid.UUID]RelatedAction)
			for _, v := range obj.RelatedActions {
				objMap[v.ActionID] = v
			}

			otherMap := make(map[uuid.UUID]RelatedAction)
			for _, v := range other.RelatedActions {
				otherMap[v.ActionID] = v
			}

			for k, v := range objMap {
				if sub, ok := otherMap[k]; ok {
					equals = v.Equals(sub)
				} else {
					return
				}
			}
		}
		return true
	}
	return
}

type SnapshotParametersMarshaled struct {
	Name                    string                          `json:"name"`
	ExpirationTime          string                          `json:"expirationTime,omitempty"`
	StateComponentFilter    storage.StateComponentFilter    `json:"stateComponentFilter,omitempty"`
	InventoryHardwareFilter storage.InventoryHardwareFilter `json:"inventoryHardwareFilter,omitempty"`
	TargetFilter            storage.TargetFilter            `json:"targetFilter,omitempty"`
}

func (obj *SnapshotParametersMarshaled) Equals(other SnapshotParametersMarshaled) bool {
	if !(obj.StateComponentFilter.Equals(other.StateComponentFilter)) ||
		!(obj.InventoryHardwareFilter.Equals(other.InventoryHardwareFilter)) ||
		!(obj.TargetFilter.Equals(other.TargetFilter)) ||
		!(obj.ExpirationTime == other.ExpirationTime) {
		return false
	}
	return true
}

//TODO the time is not coming through for expirationTime FIX ME
func (obj *SnapshotParametersMarshaled) ToSnapshotParameters() (other storage.SnapshotParameters) {
	other.Name = obj.Name
	other.TargetFilter = obj.TargetFilter
	other.StateComponentFilter = obj.StateComponentFilter
	other.InventoryHardwareFilter = obj.InventoryHardwareFilter
	timey, err := time.Parse(time.RFC3339, obj.ExpirationTime)
	if err == nil {
		other.ExpirationTime.Scan(timey)
	} else {
		logrus.Warn(err)
	}
	return
}

func ToSnapshotParametersMarshaled(obj *storage.SnapshotParameters) (other SnapshotParametersMarshaled) {

	other = SnapshotParametersMarshaled{
		Name:                    obj.Name,
		StateComponentFilter:    obj.StateComponentFilter,
		InventoryHardwareFilter: obj.InventoryHardwareFilter,
		TargetFilter:            obj.TargetFilter,
	}

	if obj.ExpirationTime.Valid == true {
		other.ExpirationTime = obj.ExpirationTime.Time.String()
	}

	return
}

// ToSnapshotMarshaled - transforms a snapshot to its marshaled form, will not fill RelatedDevices
func ToSnapshotMarshaled(s storage.Snapshot) (m SnapshotMarshaled) {
	//when you fill the whole struct it gets printed nicely, instead of null for a nil field.
	m = SnapshotMarshaled{
		Name:           s.Name,
		Ready:          s.Ready,
		Devices:        []storage.Device{},
		RelatedActions: []RelatedAction{},
		Errors:         []string{},
	}

	m.Parameters = ToSnapshotParametersMarshaled(&s.Parameters)

	m.Devices = append(m.Devices, s.Devices...)
	m.Errors = append(m.Errors, s.Errors...)

	if s.CaptureTime.Valid {
		m.CaptureTime = s.CaptureTime.Time.String()
	}
	if s.ExpirationTime.Valid {
		m.ExpirationTime = s.ExpirationTime.Time.String()
	}
	return m
}

func ToRelatedAction(a storage.Action) (r RelatedAction) {
	r = RelatedAction{
		ActionID: a.ActionID,
		State:    a.State.Current(),
	}
	if a.StartTime.Valid {
		r.StartTime = a.StartTime.Time.String()
	}
	if a.EndTime.Valid {
		r.EndTime = a.EndTime.Time.String()
	}
	return r
}

// ToSnapshotSummary will convert from snapshot to summary; will not fill the related devices.
func ToSnapshotSummary(s storage.Snapshot) (tmp SnapshotSummary) {

	tmp = SnapshotSummary{
		Name:              s.Name,
		Ready:             s.Ready,
		RelatedActions:    []RelatedAction{},
		UniqueDeviceCount: len(s.Devices),
	}

	if s.CaptureTime.Valid {
		tmp.CaptureTime = s.CaptureTime.Time.String()
	}
	if s.ExpirationTime.Valid {
		tmp.ExpirationTime = s.ExpirationTime.Time.String()
	}
	return tmp
}

func (obj *SnapshotSummaries) Equals(other SnapshotSummaries) (equals bool) {
	equals = false

	if len(obj.Summaries) == 0 && len(other.Summaries) == 0 {
		equals = true
		return
	} else {
		objMap := make(map[string]SnapshotSummary)
		for _, v := range obj.Summaries {
			objMap[v.Name] = v
		}

		otherMap := make(map[string]SnapshotSummary)
		for _, v := range other.Summaries {
			otherMap[v.Name] = v
		}

		for k, v := range objMap {
			if sub, ok := otherMap[k]; ok {
				equals = v.Equals(sub)
			} else {
				return
			}
		}
	}

	return
}

func (obj *SnapshotSummary) Equals(other SnapshotSummary) (equals bool) {
	equals = false
	if obj.Name == other.Name &&
		obj.UniqueDeviceCount == other.UniqueDeviceCount &&
		//obj.Ready == other.Ready && -- Removed from check
		obj.CaptureTime == other.CaptureTime {
		equals = true
	}
	return
}
