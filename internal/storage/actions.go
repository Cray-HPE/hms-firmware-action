/*
 * MIT License
 *
 * (C) Copyright [2020-2023] Hewlett Packard Enterprise Development LP
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
	"time"

	rf "github.com/Cray-HPE/hms-smd/pkg/redfish"

	"github.com/Cray-HPE/hms-firmware-action/internal/hsm"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/google/uuid"
	"github.com/looplab/fsm"
	"github.com/sirupsen/logrus"
)

type ActionID struct {
	ActionID uuid.UUID `json:"actionID"`
}

type Action struct {
	ActionID     uuid.UUID        `json:"id"`
	SnapshotID   uuid.UUID        `json:"snapshotID,omitempty"`
	Command      Command          `json:"command"`
	StartTime    sql.NullTime     `json:"startTime"`
	EndTime      sql.NullTime     `json:"endTime"`
	State        *fsm.FSM         `json:"state"`
	RefreshTime  sql.NullTime     `json:"refreshTime"`
	Parameters   ActionParameters `json:"parameters"`
	OperationIDs []uuid.UUID      `json:"operationIDs"`
	BlockedBy    []uuid.UUID      `json:"blockedBy"`
	Errors       []string         `json:"errors"`
	//Todo, need to add something like {xname, target} array; but not sure what targets we filter on; do we do it by
	// images then? WHY? so we can easily tell what we are locking
}

type ActionStorable struct {
	ActionID     uuid.UUID
	SnapshotID   uuid.UUID
	Command      Command
	StartTime    sql.NullTime     `json:"startTime"`
	EndTime      sql.NullTime     `json:"endTime"`
	State        string           `json:"state"`
	RefreshTime  sql.NullTime     `json:"refreshTime"`
	Parameters   ActionParameters `json:"parameters"`
	OperationIDs []uuid.UUID      `json:"operationIDs"`
	BlockedBy    []uuid.UUID      `json:"blockedBy"`
	Errors       []string         `json:"errors"`
}

type ActionStorableID struct {
	ActionID uuid.UUID `json:"id"`
}

func ToActionStorable(from Action) (to ActionStorable) {
	to = ActionStorable{
		ActionID:     from.ActionID,
		SnapshotID:   from.SnapshotID,
		Command:      from.Command,
		StartTime:    from.StartTime,
		EndTime:      from.EndTime,
		State:        from.State.Current(),
		RefreshTime:  from.RefreshTime,
		Parameters:   from.Parameters,
		OperationIDs: from.OperationIDs,
		BlockedBy:    from.BlockedBy,
		Errors:       from.Errors,
	}
	return
}

// Added id - workaround for incorrect storage from v1.26.0
// id will overwrite ActionID if ActionID is Nil
func ToActionFromStorable(from ActionStorable, id uuid.UUID) (to Action) {
	to = Action{
		ActionID:     from.ActionID,
		SnapshotID:   from.SnapshotID,
		Command:      from.Command,
		StartTime:    from.StartTime,
		EndTime:      from.EndTime,
		RefreshTime:  from.RefreshTime,
		Parameters:   from.Parameters,
		OperationIDs: from.OperationIDs,
		BlockedBy:    from.BlockedBy,
		Errors:       from.Errors,
	}
	if to.ActionID == uuid.Nil {
		to.ActionID = id
	}

	to.State = fsm.NewFSM(
		"new",
		fsm.Events{
			{Name: "configure", Src: []string{"new"}, Dst: "configured"}, //all the data is loaded, FAS should perform this op asap!
			{Name: "block", Src: []string{"configured"}, Dst: "blocked"},
			{Name: "unblock", Src: []string{"blocked"}, Dst: "configured"},
			{Name: "start", Src: []string{"configured"}, Dst: "running"},
			{Name: "finish", Src: []string{"new", "configured", "running"}, Dst: "completed"},
			{Name: "signalAbort", Src: []string{"running", "configured", "new", "blocked"}, Dst: "abortSignaled"},
			{Name: "abort", Src: []string{"abortSignaled"}, Dst: "aborted"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) { to.enterState(e) },
		},
	)

	to.restoreState(from.State)
	to.RefreshTime.Scan(from.RefreshTime.Time)

	return
}

func (d *Action) enterState(e *fsm.Event) {
	logrus.WithFields(logrus.Fields{"ActionID": d.ActionID, "event": e.Event, "destination": e.Dst}).Trace("transition")
	d.RefreshTime.Scan(time.Now())
}

func (op *Action) restoreState(state string) (err error) {
	allowedState := []string{"new", "running", "completed", "blocked", "configured", "abortSignaled", "aborted", "running"}
	for _, val := range allowedState {
		if val == state {
			op.State.SetState(state)
			logrus.WithFields(logrus.Fields{"ActionID": op.ActionID, "state": state}).Trace("restored state")
			return nil
		}
	}
	err = errors.New("failed to restore state, setting action to failed")
	op.State.SetState("failed")
	logrus.WithFields(logrus.Fields{"ActionID": op.ActionID, "state": state}).Error(err)
	return err
}

func NewAction(params ActionParameters) *Action {
	act := &Action{}
	act.StartTime.Scan(time.Now())
	act.RefreshTime.Scan(time.Now())
	act.Parameters = params
	act.ActionID = uuid.New()
	act.Command = params.Command
	act.State = fsm.NewFSM(
		"new",
		fsm.Events{
			{Name: "configure", Src: []string{"new"}, Dst: "configured"}, //all the data is loaded, FAS should perform this op asap!
			{Name: "block", Src: []string{"configured"}, Dst: "blocked"},
			{Name: "unblock", Src: []string{"blocked"}, Dst: "configured"},
			{Name: "start", Src: []string{"configured"}, Dst: "running"},
			{Name: "finish", Src: []string{"new", "configured", "running"}, Dst: "completed"},
			{Name: "signalAbort", Src: []string{"running", "configured", "new", "blocked"}, Dst: "abortSignaled"},
			{Name: "abort", Src: []string{"abortSignaled"}, Dst: "aborted"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) { act.enterState(e) },
		},
	)
	return act
}

func (d *Operation) enterState(e *fsm.Event) {
	logrus.WithFields(logrus.Fields{"operationID": d.OperationID, "event": e.Event, "destination": e.Dst}).Trace("transition")
	d.RefreshTime.Scan(time.Now())
}

func (op *Operation) restoreState(state string) (err error) {
	allowedState := []string{"initial", "configured", "needsVerified", "verifying", "aborted", "succeeded", "failed", "inProgress", "blocked", "noOperation", "noSolution"}
	for _, val := range allowedState {
		if val == state {
			op.State.SetState(state)
			logrus.WithFields(logrus.Fields{"operationID": op.OperationID, "state": state}).Trace("restored state")
			return nil
		}
	}
	err = errors.New("failed to restore state, setting operation to failed")
	op.State.SetState("failed")
	logrus.WithFields(logrus.Fields{"operationID": op.OperationID, "state": state}).Error(err)
	return err
}

func NewOperation() *Operation {
	op := &Operation{}
	op.OperationID = uuid.New()
	op.StartTime.Scan(time.Now())
	op.RefreshTime.Scan(time.Now())

	op.State = fsm.NewFSM(
		"initial",
		fsm.Events{
			{Name: "configure", Src: []string{"initial"}, Dst: "configured"},        //all the data is loaded, FAS should perform this op asap!
			{Name: "block", Src: []string{"initial", "configured"}, Dst: "blocked"}, //FAS would be actively performing it, but something needs to clear up first
			{Name: "unblock", Src: []string{"blocked"}, Dst: "configured"},          //FAS Can resume actitvely performing it
			{Name: "start", Src: []string{"configured"}, Dst: "inProgress"},         //FAS is actively performing this op but hasnt sent the command
			{Name: "restart", Src: []string{"inProgress"}, Dst: "inProgress"},       //FAS is actively performing this op but hasnt sent the command -> it is trying again, b/c the function died

			{Name: "needsVerify", Src: []string{"inProgress"}, Dst: "needsVerified"}, //FAS has launched the op, but need s to make sure it worked
			{Name: "verifying", Src: []string{"needsVerified"}, Dst: "verifying"},    //FAS has launched the op, but need s to make sure it worked
			{Name: "reverifying", Src: []string{"verifying"}, Dst: "verifying"},      //FAS has launched the op, but need s to make sure it worked -> trying it again, function died.

			{Name: "abort", Src: []string{"configured", "initial", "inProgress", "needsVerified", "verifying", "blocked"}, Dst: "aborted"},
			{Name: "noop", Src: []string{"initial"}, Dst: "noOperation"},                                                      //the versions are equal, nothing to do
			{Name: "nosol", Src: []string{"configured", "initial", "inProgress"}, Dst: "noSolution"},                          //cant find the  version or its disqualified
			{Name: "success", Src: []string{"inProgress", "verifying"}, Dst: "succeeded"},                                     // it worked
			{Name: "fail", Src: []string{"initial", "configured", "inProgress", "verifying", "needsVerified"}, Dst: "failed"}, // it failed
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) { op.enterState(e) },
		},
	)
	return op
}

type Operation struct {
	OperationID            uuid.UUID    `json:"operationID"`
	ActionID               uuid.UUID    `json:"actionID"`
	AutomaticallyGenerated bool         `json:"automaticallyGenerated,omitempty"`
	StartTime              sql.NullTime `json:"timeStart"`
	EndTime                sql.NullTime `json:"timeEnd"`
	ExpirationTime         sql.NullTime `json:"expirationTime"`
	RefreshTime            sql.NullTime `json:"refreshTime"` //when the record was last updated
	StateHelper            string       `json:"stateHelper"`
	State                  *fsm.FSM     `json:"state"`
	Error                  error        `json:"error"`
	Xname                  string       `json:"xname"`
	DeviceType             string       `json:"deviceType"`
	Target                 string       `json:"target"`
	TargetName             string       `json:"targetName"`
	Manufacturer           string       `json:"manufacturer"`
	Model                  string       `json:"model"`
	SoftwareId             string       `json:"softwareId"`
	FromFirmwareVersion    string       `json:"fromFirmwareVersion"` //mVersionCurrent
	FromImageID            uuid.UUID    `json:"fromImageID"`
	ToImageID              uuid.UUID    `json:"toImageID"`
	HsmData                hsm.HsmData  `json:"hsmData"`
	BlockedBy              []uuid.UUID  `json:"blockedBy"`
	TaskLink               string       `json:"taskLink"`
	UpdateInfoLink         string       `json:"updateInfoLink"`
}

type OperationStorable struct {
	OperationID            uuid.UUID       `json:"operationID"`
	ActionID               uuid.UUID       `json:"actionID"`
	AutomaticallyGenerated bool            `json:"automaticallyGenerated,omitempty"`
	StartTime              sql.NullTime    `json:"startTime"`
	EndTime                sql.NullTime    `json:"endTime"`
	ExpirationTime         sql.NullTime    `json:"expirationTime"`
	RefreshTime            sql.NullTime    `json:"refreshTime"` //when the record was last updated
	StateHelper            string          `json:"stateHelper"`
	State                  string          `json:"state"`
	Error                  string          `json:"error"`
	Xname                  string          `json:"xname"`
	DeviceType             string          `json:"deviceType"`
	Target                 string          `json:"target"`
	TargetName             string          `json:"targetName"`
	Manufacturer           string          `json:"manufacturer"`
	Model                  string          `json:"model"`
	SoftwareId             string          `json:"softwareId"`
	FromFirmwareVersion    string          `json:"fromFirmwareVersion"` //mVersionCurrent
	FromImageID            uuid.UUID       `json:"fromImageID"`
	ToImageID              uuid.UUID       `json:"toImageID"`
	HsmData                HsmDataStorable `json:"hsmData"`
	BlockedBy              []uuid.UUID     `json:"blockedBy"`
	TaskLink               string          `json:"taskLink"`
	UpdateInfoLink         string          `json:"updateInfoLink"`
}

func ToOperationStorable(from Operation) (to OperationStorable) {
	to = OperationStorable{
		OperationID:            from.OperationID,
		ActionID:               from.ActionID,
		AutomaticallyGenerated: from.AutomaticallyGenerated,
		StartTime:              from.StartTime,
		EndTime:                from.EndTime,
		ExpirationTime:         from.ExpirationTime,
		RefreshTime:            from.RefreshTime,
		StateHelper:            from.StateHelper,
		State:                  from.State.Current(),
		Xname:                  from.Xname,
		DeviceType:             from.DeviceType,
		Target:                 from.Target,
		TargetName:             from.TargetName,
		Manufacturer:           from.Manufacturer,
		Model:                  from.Model,
		SoftwareId:             from.SoftwareId,
		FromFirmwareVersion:    from.FromFirmwareVersion,
		FromImageID:            from.FromImageID,
		ToImageID:              from.ToImageID,
		HsmData:                ToHsmDataStorable(from.HsmData),
		BlockedBy:              from.BlockedBy,
		TaskLink:               from.TaskLink,
		UpdateInfoLink:         from.UpdateInfoLink,
	}
	if from.Error != nil {
		to.Error = from.Error.Error()
	}

	return to
}

func ToOperationFromStorable(from OperationStorable) (to Operation) {
	to = Operation{
		OperationID:            from.OperationID,
		ActionID:               from.ActionID,
		AutomaticallyGenerated: from.AutomaticallyGenerated,
		StartTime:              from.StartTime,
		EndTime:                from.EndTime,
		ExpirationTime:         from.ExpirationTime,
		RefreshTime:            from.RefreshTime,
		StateHelper:            from.StateHelper,
		Xname:                  from.Xname,
		DeviceType:             from.DeviceType,
		Target:                 from.Target,
		TargetName:             from.TargetName,
		Manufacturer:           from.Manufacturer,
		Model:                  from.Model,
		SoftwareId:             from.SoftwareId,
		FromFirmwareVersion:    from.FromFirmwareVersion,
		FromImageID:            from.FromImageID,
		ToImageID:              from.ToImageID,
		HsmData:                ToHsmDataFromStorable(from.HsmData),
		BlockedBy:              from.BlockedBy,
		TaskLink:               from.TaskLink,
		UpdateInfoLink:         from.UpdateInfoLink,
	}
	if from.Error != "" {
		to.Error = errors.New(from.Error)
	}

	to.State = fsm.NewFSM(
		"initial",
		fsm.Events{
			{Name: "configure", Src: []string{"initial"}, Dst: "configured"},         //all the data is loaded, FAS should perform this op asap!
			{Name: "block", Src: []string{"initial", "configured"}, Dst: "blocked"},  //FAS would be actively performing it, but something needs to clear up first
			{Name: "unblock", Src: []string{"blocked"}, Dst: "configured"},           //FAS Can resume actitvely performing it
			{Name: "start", Src: []string{"configured"}, Dst: "inProgress"},          //FAS is actively performing this op but hasnt sent the command
			{Name: "restart", Src: []string{"inProgress"}, Dst: "inProgress"},        //FAS is actively performing this op but hasnt sent the command -> it is trying again, b/c the function died
			{Name: "needsVerify", Src: []string{"inProgress"}, Dst: "needsVerified"}, //FAS has launched the op, but need s to make sure it worked
			{Name: "verifying", Src: []string{"needsVerified"}, Dst: "verifying"},    //FAS has launched the op, but need s to make sure it worked
			{Name: "reverifying", Src: []string{"verifying"}, Dst: "verifying"},      //FAS has launched the op, but need s to make sure it worked -> trying it again, function died.
			{Name: "abort", Src: []string{"configured", "initial", "inProgress", "needsVerified", "verifying", "blocked"}, Dst: "aborted"},
			{Name: "noop", Src: []string{"initial"}, Dst: "noOperation"},                                                      //the versions are equal, nothing to do
			{Name: "nosol", Src: []string{"configured", "initial", "inProgress"}, Dst: "noSolution"},                          //cant find the  version or its disqualified
			{Name: "success", Src: []string{"inProgress", "verifying"}, Dst: "succeeded"},                                     // it worked
			{Name: "fail", Src: []string{"initial", "configured", "inProgress", "verifying", "needsVerified"}, Dst: "failed"}, // it failed
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) { to.enterState(e) },
		},
	)

	to.restoreState(from.State)

	to.RefreshTime.Scan(from.RefreshTime.Time)
	return to
}

type HsmDataStorable struct {
	ID           string           `json:"id,omitempty"`
	Type         string           `json:"type"`
	Hostname     string           `json:"hostname,omitempty"`
	Domain       string           `json:"domain,omitempty"`
	FQDN         string           `json:"FQDN"`
	UpdateURI    string           `json:"updateURI"`
	InventoryURI string           `json:"inventoryURI"`
	Role         string           `json:"role"`
	Model        string           `json:"model"`
	Manufacturer string           `json:"manufacturer"`
	Error        string           `json:"error"`
	BmcPath      string           `json:"bmdPath"'`
	RfType       string           `json:"rfType"`
	DiscInfo     rf.DiscoveryInfo `json:"discoveryInfo,omitempty"`
	ActionReset  rf.ActionReset   `json:"actionReset,omitempty"`
}

func ToHsmDataStorable(from hsm.HsmData) (to HsmDataStorable) {
	to = HsmDataStorable{
		ID:           from.ID,
		Type:         from.Type,
		Hostname:     from.Hostname,
		Domain:       from.Domain,
		FQDN:         from.FQDN,
		UpdateURI:    from.UpdateURI,
		InventoryURI: from.InventoryURI,
		Role:         from.Role,
		Model:        from.Model,
		Manufacturer: from.Manufacturer,
		BmcPath:      from.BmcPath,
		RfType:       from.RfType,
		DiscInfo:     from.DiscInfo,
		ActionReset:  from.ActionReset,
	}
	if from.Error != nil {
		to.Error = from.Error.Error()
	}
	return to
}

func ToHsmDataFromStorable(from HsmDataStorable) (to hsm.HsmData) {
	to = hsm.HsmData{
		ID:           from.ID,
		Type:         from.Type,
		Hostname:     from.Hostname,
		Domain:       from.Domain,
		FQDN:         from.FQDN,
		UpdateURI:    from.UpdateURI,
		InventoryURI: from.InventoryURI,
		Role:         from.Role,
		Model:        from.Model,
		Manufacturer: from.Manufacturer,
		BmcPath:      from.BmcPath,
		RfType:       from.RfType,
		DiscInfo:     from.DiscInfo,
		ActionReset:  from.ActionReset,
	}

	if from.Error != "" {
		to.Error = errors.New(from.Error)
	}
	return to
}

func (obj *Action) Equals(other Action) bool {
	if !(obj.ActionID == other.ActionID) {
		logrus.Warn("ActionID not equal")
		return false
	} else if !(obj.SnapshotID == other.SnapshotID) {
		logrus.Warn("SnapshotID not equal")
		return false
	} else if !(obj.Command == other.Command) {
		logrus.Warn("Command not equal")
		return false
	} else if !(obj.StartTime.Time.Round(0).Equal(other.StartTime.Time.Round(0))) {
		logrus.Warn("StartTime not equal")
		return false
	} else if !(obj.EndTime.Time.Round(0).Equal(other.EndTime.Time.Round(0))) {
		logrus.Warn("EndTime not equal")
		return false
	} else if !(obj.State.Current() == other.State.Current()) {
		logrus.Warn("state not equal")
		return false
	} else if !(obj.RefreshTime.Time.Round(0).Equal(other.RefreshTime.Time.Round(0))) {
		logrus.Warn("RefreshTime not equal")
		return false
	} else if !(obj.Command.Equals(other.Command)) {
		logrus.Warn("Command not equal")
		return false
	} else if !(obj.Parameters.Equals(other.Parameters)) {
		logrus.Warn("Parameters not equal")
		return false
	} else if !(model.UUIDSliceEquals(obj.BlockedBy, other.BlockedBy)) {
		logrus.Warn("BlockedBy not equal")
		return false
	} else if !(model.UUIDSliceEquals(obj.OperationIDs, other.OperationIDs)) {
		logrus.Warn("OperationIDs not equal")
		return false
	}
	return true
}

func (obj *Operation) Equals(other Operation) bool {
	if obj.OperationID != other.OperationID {
		logrus.Warn("operationID not equal")
		return false
	} else if obj.ActionID != other.ActionID {
		logrus.Warn("operationID not equal")
		return false
	} else if obj.StartTime.Time.Round(0).Equal(other.StartTime.Time.Round(0)) == false {
		logrus.Warn("operationID not equal")
		return false
	} else if obj.EndTime.Time.Round(0).Equal(other.EndTime.Time.Round(0)) == false {
		logrus.Warn("operationID not equal")
		return false
	} else if obj.ExpirationTime.Time.Round(0).Equal(other.ExpirationTime.Time.Round(0)) == false {
		logrus.Warn("ExpirationTime not equal")
		return false
	} else if obj.RefreshTime.Time.Round(0).Equal(other.RefreshTime.Time.Round(0)) == false {
		logrus.Warn("RefreshTime not equal")
		return false
	} else if obj.StateHelper != other.StateHelper {
		logrus.Warn("StateHelper not equal")
		return false
	} else if !(obj.State.Current() == other.State.Current()) {
		logrus.Warn("State not equal")
		return false
	} else if !(obj.Error == other.Error) {
		logrus.Warn("Error not equal")
		return false
	} else if !(obj.Xname == other.Xname) {
		logrus.Warn("Xname not equal")
		return false
	} else if !(obj.DeviceType == other.DeviceType) {
		logrus.Warn("DeviceType not equal")
		return false
	} else if !(obj.Target == other.Target) {
		logrus.Warn("Target not equal")
		return false
	} else if !(obj.TargetName == other.TargetName) {
		logrus.Warn("TargetName not equal")
		return false
	} else if !(obj.Manufacturer == other.Manufacturer) {
		logrus.Warn("Manufacturer not equal")
		return false
	} else if !(obj.Model == other.Model) {
		logrus.Warn("Model not equal")
		return false
	} else if !(obj.FromFirmwareVersion == other.FromFirmwareVersion) {
		logrus.Warn("FromFirmwareVersion not equal")
		return false
	} else if !(obj.FromImageID == other.FromImageID) {
		logrus.Warn("FromImageID not equal")
		return false
	} else if !(obj.ToImageID == other.ToImageID) {
		logrus.Warn("ToImageID not equal")
		return false
	} else if obj.HsmData.Equals(other.HsmData) == false {
		logrus.Warn("hsmData not equal")
		return false
	} else if model.UUIDSliceEquals(obj.BlockedBy, other.BlockedBy) == false {
		logrus.Warn("blockedBy not equal")
		return false
	} else if !(obj.SoftwareId == other.SoftwareId) {
		logrus.Warn("softwareId not equal")
		return false
	} else if !(obj.TaskLink == other.TaskLink) {
		logrus.Warn("taskLink not equal")
		return false
	} else if !(obj.UpdateInfoLink == other.UpdateInfoLink) {
		logrus.Warn("updateInfoLink not equal")
		return false
	}
	return true
}

func NewActionID() (id ActionID) {
	id.ActionID = uuid.New()
	return
}

type ActionParameters struct {
	StateComponentFilter    StateComponentFilter    `json:"stateComponentFilter,omitempty"`
	InventoryHardwareFilter InventoryHardwareFilter `json:"inventoryHardwareFilter,omitempty"`
	ImageFilter             ImageFilter             `json:"imageFilter,omitempty"`
	TargetFilter            TargetFilter            `json:"targetFilter,omitempty"`
	Command                 Command                 `json:"command"`
}

//this may ONLY resolve to 1 imageID
type ImageFilter struct {
	ImageID       uuid.UUID `json:"imageID"`
	OverrideImage bool      `json:"overrideImage"`
}

func (obj *ImageFilter) Equals(other ImageFilter) bool {
	if obj.ImageID == other.ImageID &&
		obj.OverrideImage == other.OverrideImage {
		return true
	}
	return false
}

//The addition of TAG to command makes this more of a last second sanity check.  We had the option to add Tags to
//imageFilter, however if we did that we would always expect an image filter.  We had been implicitly using tag, in that
//we expected the image.Tags to be (contain) 'default' otherwise we would NOT match.  It is possible that by putting tag
//in command that there may be extra operations that get set at noSolution at the last possible second, whereas a more
//preemptive approach would be to filter images that dont bare that tag. B/c tag was being used implicitly we decided to
//put it in command instead of in imageFilter.
//Tags does NOT impact the use of an imageID, or the taking of a snapshot.   B/c in both of those cases the Version is
//set to explicit
type Command struct {
	OverrideDryrun             bool `json:"overrideDryrun"`
	RestoreNotPossibleOverride bool `json:"restoreNotPossibleOverride"` // it is probable in many cases that there will NOT be any return
	// image to go back to. In that case we should NOT update unless the override is set, that way we can always get back
	OverwriteSameImage bool   `json:"overwriteSameImage"`  // If to and from version are the same, update anyways
	TimeLimit_Seconds  int    `json:"timeLimit,omitempty"` //IDEA IS THAT IT WILL BE SECONDS
	Version            string `json:"version"`             //earliest, latest
	Tag                string `json:"tag"`
	Description        string `json:"description"` //WHY are you doing this action?
}

func (obj *Command) Equals(other Command) bool {
	if obj.OverrideDryrun == other.OverrideDryrun &&
		obj.RestoreNotPossibleOverride == other.RestoreNotPossibleOverride &&
		obj.TimeLimit_Seconds == other.TimeLimit_Seconds &&
		obj.Version == other.Version &&
		obj.Description == other.Description {
		return true
	}
	return false
}

func (obj *ActionParameters) Equals(other ActionParameters) bool {
	if obj.StateComponentFilter.Equals(other.StateComponentFilter) &&
		obj.InventoryHardwareFilter.Equals(other.InventoryHardwareFilter) &&
		obj.TargetFilter.Equals(other.TargetFilter) &&
		obj.ImageFilter.Equals(other.ImageFilter) &&
		obj.Command.Equals(other.Command) {
		return true
	}
	return false
}
