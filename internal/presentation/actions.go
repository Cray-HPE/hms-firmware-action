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
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/model"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/storage"
)

type OperationKey struct {
	OperationID         uuid.UUID `json:"operationID"`
	Xname               string    `json:"xname"`
	Target              string    `json:"target"`
	TargetName          string    `json:"targetName"`
	FromFirmwareVersion string    `json:"fromFirmwareVersion"`
	StateHelper         string    `json:"stateHelper"`
}

type ActionSummary struct {
	ActionID        uuid.UUID       `json:"actionID"`
	SnapshotID      uuid.UUID       `json:"snapshotID,omitempty"`
	Command         storage.Command `json:"command"`
	StartTime       string          `json:"startTime"`
	EndTime         string          `json:"endTime,omitempty"`
	State           string          `json:"state"`
	OperationCounts OperationCounts `json:"operationCounts"`
	BlockedBy       []uuid.UUID     `json:"blockedBy"`
}

type OperationCounts struct {
	Total         int `json:"total"`
	Initial       int `json:"initial"`
	Configured    int `json:"configured"`
	Blocked       int `json:"blocked"`
	NeedsVerified int `json:"needsVerified"`
	Verifying     int `json:"verifying"`
	InProgress    int `json:"inProgress"`
	Failed        int `json:"failed"`
	Succeeded     int `json:"succeeded"`
	NoOperation   int `json:"noOperation"`
	NoSolution    int `json:"noSolution"`
	Aborted       int `json:"aborted"`
	Unknown       int `json:"unknown"`
}

// OperationSummary and OperationDetail both need to updated if
// states are added or removed
type OperationKeys struct {
	OperationsKeys []OperationKey `json:"operationKeys"`
}
type OperationSummary struct {
	Initial       OperationKeys `json:"initial"'`
	Configured    OperationKeys `json:"configured"`    //Not done yet
	Blocked       OperationKeys `json:"blocked"`       //Not done yet
	InProgress    OperationKeys `json:"inProgress"`    //Not done yet
	NeedsVerified OperationKeys `json:"needsVerified"` //Not done yet
	Verifying     OperationKeys `json:"verifying"`     //Not done yet
	Failed        OperationKeys `json:"failed"`        //Done it failed
	Succeeded     OperationKeys `json:"succeeded"`     // Done, it worked
	NoOperation   OperationKeys `json:"noOperation"`   //Nothing done
	NoSolution    OperationKeys `json:"noSolution"`    //nothing CAN be done
	Aborted       OperationKeys `json:"aborted"`       //IT was aborted
	Unknown       OperationKeys `json:"unknown"`       //the state isnt set, but an op exists
}

// OperationSummary and OperationDetail both need to updated if
// states are added or removed
type OperationKeysDetail struct {
	OperationsKeys []OperationMarshaled `json:"operationKeys"`
}
type OperationDetail struct {
	Initial       OperationKeysDetail `json:"initial"'`
	Configured    OperationKeysDetail `json:"configured"`    //Not done yet
	Blocked       OperationKeysDetail `json:"blocked"`       //Not done yet
	InProgress    OperationKeysDetail `json:"inProgress"`    //Not done yet
	NeedsVerified OperationKeysDetail `json:"needsVerified"` //Not done yet
	Verifying     OperationKeysDetail `json:"verifying"`     //Not done yet
	Failed        OperationKeysDetail `json:"failed"`        //Done it failed
	Succeeded     OperationKeysDetail `json:"succeeded"`     // Done, it worked
	NoOperation   OperationKeysDetail `json:"noOperation"`   //Nothing done
	NoSolution    OperationKeysDetail `json:"noSolution"`    //nothing CAN be done
	Aborted       OperationKeysDetail `json:"aborted"`       //IT was aborted
	Unknown       OperationKeysDetail `json:"unknown"`       //the state isnt set, but an op exists
}

type ActionSummaries struct {
	Actions []ActionSummary `json:"actions"`
}

type ActionMarshaled struct {
	ActionID         uuid.UUID                `json:"actionID"`
	SnapshotID       uuid.UUID                `json:"snapshotID,omitempty"`
	Command          storage.Command          `json:"command"`
	StartTime        string                   `json:"startTime"`
	EndTime          string                   `json:"endTime,omitempty"`
	State            string                   `json:"state"`
	Parameters       storage.ActionParameters `json:"parameters,omitempty"`
	OperationSummary OperationSummary         `json:"operationSummary"`
	BlockedBy        []uuid.UUID              `json:"blockedBy"`
}

type ActionOperationsDetail struct {
	ActionID         uuid.UUID                `json:"actionID"`
	SnapshotID       uuid.UUID                `json:"snapshotID,omitempty"`
	Command          storage.Command          `json:"command"`
	StartTime        string                   `json:"startTime"`
	EndTime          string                   `json:"endTime,omitempty"`
	State            string                   `json:"state"`
	Parameters       storage.ActionParameters `json:"parameters,omitempty"`
	OperationDetails OperationDetail          `json:"operationDetails"`
	BlockedBy        []uuid.UUID              `json:"blockedBy"`
}

type OperationPlusImages struct {
	Operation storage.Operation `json:"operation"`
	ToImage   storage.Image     `json:"toImage"`
	FromImage storage.Image     `json:"fromImage"`
}

type OperationMarshaled struct {
	OperationID                 uuid.UUID   `json:"operationID"`
	ActionID                    uuid.UUID   `json:"actionID"`
	State                       string      `json:"state"`
	StateHelper                 string      `json:"stateHelper"`
	StartTime                   string      `json:"startTime"`
	EndTime                     string      `json:"endTime,omitempty"`
	RefreshTime                 string      `json:"refreshTime"`
	ExpirationTime              string      `json:"expirationTime"`
	Xname                       string      `json:"xname"`
	DeviceType                  string      `json:"deviceType"`
	Target                      string      `json:"target"`
	TargetName                  string      `json:"targetName"`
	Manufacturer                string      `json:"manufacturer"`
	Model                       string      `json:"model"`
	SoftwareId                  string      `json:"softwareId"`
	FromImageID                 uuid.UUID   `json:"fromImageID"`
	FromSemanticFirmwareVersion string      `json:"fromSemanticFirmwareVersion"` //versionCurrent
	FromFirmwareVersion         string      `json:"fromFirmwareVersion"`         //mVersionCurrent
	FromImageURL                string      `json:"fromImageURL"`
	FromTag                     string      `json:"fromTag"`
	ToImageID                   uuid.UUID   `json:"toImageID"`
	ToSemanticFirmwareVersion   string      `json:"toSemanticFirmwareVersion"` //versionUpdate
	ToFirmwareVersion           string      `json:"toFirmwareVersion"`         //mVersionUpdate
	ToImageURL                  string      `json:"toImageURL"`
	ToTag                       string      `json:"toTag"`
	BlockedBy                   []uuid.UUID `json:"blockedBy"`
}

func (obj *ActionSummaries) Equals(other ActionSummaries) (equals bool) {
	equals = false
	if len(obj.Actions) != len(other.Actions) {
		return
	}
	if len(other.Actions) == 0 {
		equals = true
	}

	objVDMap := make(map[uuid.UUID]ActionSummary)
	for _, v := range obj.Actions {
		objVDMap[v.ActionID] = v
	}

	otherVDMap := make(map[uuid.UUID]ActionSummary)
	for _, v := range other.Actions {
		otherVDMap[v.ActionID] = v
	}

	for _, v := range obj.Actions {
		if sub, ok := otherVDMap[v.ActionID]; ok {
			if equals = v.Equals(sub); !equals {
				return
			}
		} else {
			return
		}
	}
	return
}

func (obj *OperationMarshaled) Equals(other OperationMarshaled) bool {
	if obj.Xname != other.Xname ||
		obj.Target != other.Target ||
		obj.TargetName != other.TargetName ||
		obj.FromFirmwareVersion != other.FromFirmwareVersion ||
		obj.ToImageID != other.ToImageID ||
		obj.FromImageID != other.FromImageID ||
		obj.State != other.State {
		return false
	}
	return true
}

func (obj *OperationSummary) Equals(other OperationSummary) bool {
	if obj.InProgress.Equals(other.InProgress) == false ||
		obj.Initial.Equals(other.Initial) == false ||
		obj.Aborted.Equals(other.Aborted) == false ||
		obj.Unknown.Equals(other.Unknown) == false ||
		obj.Configured.Equals(other.Configured) == false ||
		obj.Blocked.Equals(other.Blocked) == false ||
		obj.NeedsVerified.Equals(other.NeedsVerified) == false ||
		obj.Verifying.Equals(other.Verifying) == false ||
		obj.Failed.Equals(other.Failed) == false ||
		obj.Succeeded.Equals(other.Succeeded) == false ||
		obj.NoOperation.Equals(other.NoOperation) == false ||
		obj.NoSolution.Equals(other.NoSolution) == false {
		return false
	}
	return true
}

func (obj *OperationKeys) Equals(other OperationKeys) bool {

	if len(obj.OperationsKeys) != len(other.OperationsKeys) {
		return false
	}
	objMap := make(map[uuid.UUID]OperationKey)
	otherMap := make(map[uuid.UUID]OperationKey)

	for _, objE := range obj.OperationsKeys {
		objMap[objE.OperationID] = objE
	}
	for _, otherE := range other.OperationsKeys {
		otherMap[otherE.OperationID] = otherE
	}
	for objKey, objVal := range objMap {
		if otherVal, ok := otherMap[objKey]; ok {
			if objVal.Equals(otherVal) == false {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func (obj *ActionMarshaled) Equals(other ActionMarshaled) bool {
	if obj.ActionID != other.ActionID ||
		obj.SnapshotID != other.SnapshotID ||
		obj.Command.Equals(other.Command) == false ||
		obj.StartTime != other.StartTime ||
		obj.EndTime != other.EndTime ||
		!obj.Parameters.Equals(other.Parameters) ||
		!obj.OperationSummary.Equals(other.OperationSummary) ||
		obj.State != other.State ||
		model.UUIDSliceEquals(obj.BlockedBy, other.BlockedBy) == false {
		return false
	}
	return true
}
func (obj *ActionSummary) Equals(other ActionSummary) bool {
	if obj.ActionID == other.ActionID &&
		obj.Command.Equals(other.Command) &&
		obj.StartTime == other.StartTime &&
		obj.EndTime == other.EndTime &&
		obj.OperationCounts == other.OperationCounts &&
		obj.SnapshotID == other.SnapshotID &&
		obj.State == other.State &&
		model.UUIDSliceEquals(obj.BlockedBy, other.BlockedBy) == true {
		return true
	}
	return false
}

func (obj *OperationKey) Equals(other OperationKey) bool {
	if obj.Xname != other.Xname ||
		obj.OperationID != other.OperationID ||
		obj.Target != other.Target ||
		obj.TargetName != other.TargetName {
		return false
	}
	return true
}

func ToActionSummaryFromAction(a storage.Action) (s ActionSummary, err error) {
	s.ActionID = a.ActionID
	s.SnapshotID = a.SnapshotID
	s.Command = a.Command
	s.State = a.State.Current()

	if len(a.BlockedBy) == 0 {
		s.BlockedBy = []uuid.UUID{}
	} else {
		s.BlockedBy = a.BlockedBy
	}

	if a.StartTime.Valid {
		s.StartTime = a.StartTime.Time.String()
	}
	if a.EndTime.Valid {
		s.EndTime = a.EndTime.Time.String()
	}
	return s, err
}

func ToOperationCountsFromOperations(o []storage.Operation) (c OperationCounts, err error) {
	for _, op := range o {
		c.Total++

		if op.State != nil {
			if op.State.Is("initial") {
				c.Initial++
			} else if op.State.Is("configured") {
				c.Configured++
			} else if op.State.Is("blocked") {
				c.Blocked++
			} else if op.State.Is("inProgress") {
				c.InProgress++
			} else if op.State.Is("needsVerified") {
				c.NeedsVerified++
			} else if op.State.Is("verifying") {
				c.Verifying++
			} else if op.State.Is("failed") {
				c.Failed++
			} else if op.State.Is("succeeded") {
				c.Succeeded++
			} else if op.State.Is("noOperation") {
				c.NoOperation++
			} else if op.State.Is("noSolution") {
				c.NoSolution++
			} else if op.State.Is("aborted") {
				c.Aborted++
			}
		} else {
			c.Unknown++
			logrus.WithField("operationID", op.OperationID).Warn("Cannot count operation, the State is nil")
		}

	}
	return
}

func ToActionMarshaledFromAction(a storage.Action) (m ActionMarshaled, err error) {
	m = ActionMarshaled{
		ActionID:   a.ActionID,
		SnapshotID: a.SnapshotID,
		Command:    a.Command,
		State:      a.State.Current(),
		Parameters: a.Parameters,
	}

	if len(a.BlockedBy) == 0 {
		m.BlockedBy = []uuid.UUID{}
	} else {
		m.BlockedBy = a.BlockedBy
	}

	if a.StartTime.Valid {
		m.StartTime = a.StartTime.Time.String()
	}
	if a.EndTime.Valid {
		m.EndTime = a.EndTime.Time.String()
	}
	return m, err
}

func ToActionOperationsDetailFromAction(a storage.Action) (m ActionOperationsDetail, err error) {
	m = ActionOperationsDetail{
		ActionID:   a.ActionID,
		SnapshotID: a.SnapshotID,
		Command:    a.Command,
		State:      a.State.Current(),
		Parameters: a.Parameters,
	}

	if len(a.BlockedBy) == 0 {
		m.BlockedBy = []uuid.UUID{}
	} else {
		m.BlockedBy = a.BlockedBy
	}

	if a.StartTime.Valid {
		m.StartTime = a.StartTime.Time.String()
	}
	if a.EndTime.Valid {
		m.EndTime = a.EndTime.Time.String()
	}
	return m, err
}

func ToOperationMarshaledFromOperation(o storage.Operation) (m OperationMarshaled, err error) {
	m = OperationMarshaled{
		OperationID:         o.OperationID,
		ActionID:            o.ActionID,
		State:               o.State.Current(),
		StateHelper:         o.StateHelper,
		Xname:               o.Xname,
		DeviceType:          o.DeviceType,
		Target:              o.Target,
		TargetName:          o.TargetName,
		Manufacturer:        o.Manufacturer,
		Model:               o.Model,
		SoftwareId:          o.SoftwareId,
		FromFirmwareVersion: o.FromFirmwareVersion,
		FromImageID:         o.FromImageID,
		ToImageID:           o.ToImageID,
	}
	if len(o.BlockedBy) == 0 {
		m.BlockedBy = []uuid.UUID{}
	} else {
		m.BlockedBy = o.BlockedBy
	}

	if o.StartTime.Valid {
		m.StartTime = o.StartTime.Time.String()
	}
	if o.EndTime.Valid {
		m.EndTime = o.EndTime.Time.String()
	}
	if o.RefreshTime.Valid {
		m.RefreshTime = o.RefreshTime.Time.String()
	}
	if o.ExpirationTime.Valid {
		m.ExpirationTime = o.ExpirationTime.Time.String()
	}
	return m, err
}

func ToOperationSummaryFromOperations(o []storage.Operation) (c OperationSummary, err error) {

	c.Initial = OperationKeys{OperationsKeys: []OperationKey{}}
	c.Configured = OperationKeys{OperationsKeys: []OperationKey{}}
	c.Blocked = OperationKeys{OperationsKeys: []OperationKey{}}
	c.InProgress = OperationKeys{OperationsKeys: []OperationKey{}}
	c.NeedsVerified = OperationKeys{OperationsKeys: []OperationKey{}}
	c.Verifying = OperationKeys{OperationsKeys: []OperationKey{}}
	c.Succeeded = OperationKeys{OperationsKeys: []OperationKey{}}
	c.Failed = OperationKeys{OperationsKeys: []OperationKey{}}
	c.NoOperation = OperationKeys{OperationsKeys: []OperationKey{}}
	c.NoSolution = OperationKeys{OperationsKeys: []OperationKey{}}
	c.Aborted = OperationKeys{OperationsKeys: []OperationKey{}}
	c.Unknown = OperationKeys{OperationsKeys: []OperationKey{}}

	for _, op := range o {
		opkey := OperationKey{
			OperationID:         op.OperationID,
			Xname:               op.Xname,
			Target:              op.Target,
			TargetName:          op.TargetName,
			FromFirmwareVersion: op.FromFirmwareVersion,
			StateHelper:         op.StateHelper,
		}
		if op.State != nil {
			if op.State.Is("initial") {
				c.Initial.OperationsKeys = append(c.Initial.OperationsKeys, opkey)
			} else if op.State.Is("configured") {
				c.Configured.OperationsKeys = append(c.Configured.OperationsKeys, opkey)
			} else if op.State.Is("blocked") {
				c.Blocked.OperationsKeys = append(c.Blocked.OperationsKeys, opkey)
			} else if op.State.Is("inProgress") {
				c.InProgress.OperationsKeys = append(c.InProgress.OperationsKeys, opkey)
			} else if op.State.Is("needsVerified") {
				c.NeedsVerified.OperationsKeys = append(c.NeedsVerified.OperationsKeys, opkey)
			} else if op.State.Is("verifying") {
				c.Verifying.OperationsKeys = append(c.Verifying.OperationsKeys, opkey)
			} else if op.State.Is("failed") {
				c.Failed.OperationsKeys = append(c.Failed.OperationsKeys, opkey)
			} else if op.State.Is("succeeded") {
				c.Succeeded.OperationsKeys = append(c.Succeeded.OperationsKeys, opkey)
			} else if op.State.Is("noOperation") {
				c.NoOperation.OperationsKeys = append(c.NoOperation.OperationsKeys, opkey)
			} else if op.State.Is("noSolution") {
				c.NoSolution.OperationsKeys = append(c.NoSolution.OperationsKeys, opkey)
			} else if op.State.Is("aborted") {
				c.Aborted.OperationsKeys = append(c.Aborted.OperationsKeys, opkey)
			}
		} else {
			c.Unknown.OperationsKeys = append(c.Unknown.OperationsKeys, opkey)
		}
		logrus.WithField("operationID", op.OperationID).Warn("Cannot summarize operation, the State is nil")
	}
	return
}

func ToOperationDetailFromOperations(o []OperationPlusImages) (c OperationDetail, err error) {

	c.Initial = OperationKeysDetail{OperationsKeys: []OperationMarshaled{}}
	c.Configured = OperationKeysDetail{OperationsKeys: []OperationMarshaled{}}
	c.Blocked = OperationKeysDetail{OperationsKeys: []OperationMarshaled{}}
	c.InProgress = OperationKeysDetail{OperationsKeys: []OperationMarshaled{}}
	c.NeedsVerified = OperationKeysDetail{OperationsKeys: []OperationMarshaled{}}
	c.Verifying = OperationKeysDetail{OperationsKeys: []OperationMarshaled{}}
	c.Succeeded = OperationKeysDetail{OperationsKeys: []OperationMarshaled{}}
	c.Failed = OperationKeysDetail{OperationsKeys: []OperationMarshaled{}}
	c.NoOperation = OperationKeysDetail{OperationsKeys: []OperationMarshaled{}}
	c.NoSolution = OperationKeysDetail{OperationsKeys: []OperationMarshaled{}}
	c.Aborted = OperationKeysDetail{OperationsKeys: []OperationMarshaled{}}
	c.Unknown = OperationKeysDetail{OperationsKeys: []OperationMarshaled{}}

	for _, opi := range o {
		op := opi.Operation
		opkey, _ := ToOperationMarshaledFromOperation(op)
		im := opi.FromImage
		opkey.FromImageURL = im.S3URL
		if im.SemanticFirmwareVersion != nil {
			opkey.FromSemanticFirmwareVersion = im.SemanticFirmwareVersion.String()
		}
		im = opi.ToImage
		opkey.ToImageURL = im.S3URL
		if im.SemanticFirmwareVersion != nil {
			opkey.ToSemanticFirmwareVersion = im.SemanticFirmwareVersion.String()
		}
		opkey.ToFirmwareVersion = im.FirmwareVersion
		if op.State != nil {
			if op.State.Is("initial") {
				c.Initial.OperationsKeys = append(c.Initial.OperationsKeys, opkey)
			} else if op.State.Is("configured") {
				c.Configured.OperationsKeys = append(c.Configured.OperationsKeys, opkey)
			} else if op.State.Is("blocked") {
				c.Blocked.OperationsKeys = append(c.Blocked.OperationsKeys, opkey)
			} else if op.State.Is("inProgress") {
				c.InProgress.OperationsKeys = append(c.InProgress.OperationsKeys, opkey)
			} else if op.State.Is("needsVerified") {
				c.NeedsVerified.OperationsKeys = append(c.NeedsVerified.OperationsKeys, opkey)
			} else if op.State.Is("verifying") {
				c.Verifying.OperationsKeys = append(c.Verifying.OperationsKeys, opkey)
			} else if op.State.Is("failed") {
				c.Failed.OperationsKeys = append(c.Failed.OperationsKeys, opkey)
			} else if op.State.Is("succeeded") {
				c.Succeeded.OperationsKeys = append(c.Succeeded.OperationsKeys, opkey)
			} else if op.State.Is("noOperation") {
				c.NoOperation.OperationsKeys = append(c.NoOperation.OperationsKeys, opkey)
			} else if op.State.Is("noSolution") {
				c.NoSolution.OperationsKeys = append(c.NoSolution.OperationsKeys, opkey)
			} else if op.State.Is("aborted") {
				c.Aborted.OperationsKeys = append(c.Aborted.OperationsKeys, opkey)
			}
		} else {
			c.Unknown.OperationsKeys = append(c.Unknown.OperationsKeys, opkey)
		}
		logrus.WithField("operationID", op.OperationID).Warn("Cannot summarize operation, the State is nil")
	}
	return
}

type CreateActionPayload struct {
	ActionID       uuid.UUID `json:"actionID"`
	OverrideDryrun bool      `json:"overrideDryrun"`
}
