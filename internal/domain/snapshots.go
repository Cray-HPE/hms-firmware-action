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

package domain

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/Cray-HPE/hms-firmware-action/internal/hsm"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/presentation"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
)

func GetAllExpiredSnapshots() (expiredSnapshots storage.Snapshots) {
	snapshots, err := (*GLOB.DSP).GetSnapshots()
	if err != nil {
		logrus.Error(err)
		return
	}

	for _, s := range snapshots {
		if s.ExpirationTime.Valid && s.ExpirationTime.Time.Before(time.Now()) {
			expiredSnapshots.Snapshots = append(expiredSnapshots.Snapshots, s)
		}
	}
	return
}

// GetSnapshots - gets all snapshot summaries
func GetSnapshots() (pb model.Passback) {
	snapshots, err := (*GLOB.DSP).GetSnapshots()
	if err != nil {
		logrus.Error(err)
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		return
	}

	summaries := presentation.SnapshotSummaries{Summaries: []presentation.SnapshotSummary{}}
	for _, snapshot := range snapshots {
		summary := presentation.ToSnapshotSummary(snapshot)
		relatedAction := presentation.RelatedAction{}
		for _, actionID := range snapshot.RelatedActions {
			action, _ := (*GLOB.DSP).GetAction(actionID)
			if err != nil {
				logrus.WithFields(logrus.Fields{"ERROR": err, "actionID": action.ActionID.String()}).Error("Could not get associated action")
				break
			}
			relatedAction = presentation.ToRelatedAction(action)
			summary.RelatedActions = append(summary.RelatedActions, relatedAction)
		}
		summaries.Summaries = append(summaries.Summaries, summary)
	}
	pb = model.BuildSuccessPassback(http.StatusOK, summaries)
	return pb
}

// GetSnapshot - gets a snapshot summary
func GetSnapshot(name string) (pb model.Passback) {
	snapshot, err := (*GLOB.DSP).GetSnapshot(name)
	if err != nil {
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
		return
	}
	snapshotMarshaled := presentation.ToSnapshotMarshaled(snapshot)
	for _, actionID := range snapshot.RelatedActions {
		action, _ := (*GLOB.DSP).GetAction(actionID)
		if err != nil {
			logrus.WithFields(logrus.Fields{"ERROR": err, "actionID": action.ActionID.String()}).Error("Could not get associated action")
			break
		}
		relatedAction := presentation.ToRelatedAction(action)
		snapshotMarshaled.RelatedActions = append(snapshotMarshaled.RelatedActions, relatedAction)
	}
	pb = model.BuildSuccessPassback(http.StatusOK, snapshotMarshaled)
	return pb
}

//TODO if a ImageID isnt set for a device target, we cant restore it! -> therefore we should probably point that out in the return data!
func CreateSnapshot(parameters storage.SnapshotParameters) (pb model.Passback) {
	//check if the name already exists; if it does CONFLICT!
	_, err := (*GLOB.DSP).GetSnapshot(parameters.Name)
	if err == nil {
		pb = model.BuildErrorPassback(http.StatusConflict, errors.New("Snapshot with same name already exists"))
		return
	}

	var snapshot storage.Snapshot
	snapshot.CaptureTime.Scan(time.Now())
	snapshot.Ready = false
	snapshot.Name = parameters.Name
	snapshot.Parameters = parameters

	if parameters.ExpirationTime.Valid {
		snapshot.ExpirationTime.Scan(parameters.ExpirationTime.Time)
	}

	err = (*GLOB.DSP).StoreSnapshot(snapshot)
	go BuildSnapshot(snapshot)
	if err != nil {
		logrus.Error(err)
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		return pb
	}
	var ssn presentation.SnapshotName
	ssn.Name = snapshot.Name
	pb = model.BuildSuccessPassback(http.StatusCreated, ssn)
	return pb
}

func BuildSnapshot(snapshot storage.Snapshot) {
	devices, errlist := GetCurrentFirmwareVersionsFromParams(snapshot.Parameters)
	snapshot.Errors = append(snapshot.Errors, errlist...)
	snapshot.Errors = model.RemoveDuplicateStrings(snapshot.Errors)
	snapshot.Devices = devices
	snapshot.Ready = true
	err := (*GLOB.DSP).StoreSnapshot(snapshot)
	if err != nil {
		logrus.Error(err)
	}
}

func StartRestoreSnapshot(name string, overrideDryrun bool, timeLimit int) (pb model.Passback) {
	//pb = GetSnapshot(name)
	snapshot, err := (*GLOB.DSP).GetSnapshot(name)
	if err != nil {
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
		return
	}

	//snapshot := pb.Obj.(storage.Snapshot)
	actionParams := storage.ActionParameters{}

	actionParams.Command = storage.Command{
		OverrideDryrun:             overrideDryrun,
		RestoreNotPossibleOverride: true,
		TimeLimit_Seconds:          timeLimit,
		Version:                    "explicit",
		Description:                "restore snapshot " + snapshot.Name,
	}
	action := storage.NewAction(actionParams)

	err = (*GLOB.DSP).StoreAction(*action)
	if err != nil {
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
	} else {
		actionID := storage.ActionID{ActionID: action.ActionID}
		CAP := presentation.CreateActionPayload{
			ActionID:       actionID.ActionID,
			OverrideDryrun: actionParams.Command.OverrideDryrun,
		}
		pb = model.BuildSuccessPassback(http.StatusAccepted, CAP)
		go RestoreSnapshot(*action, snapshot)
	}
	return pb
}

func DeleteSnapshot(name string) (pb model.Passback) {
	pb = GetSnapshot(name)
	if pb.IsError == true {
		logrus.Error(pb.Error)
		return
	}

	if err := (*GLOB.DSP).DeleteSnapshot(name); err != nil {
		logrus.Error(err)
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		return
	}
	pb = model.BuildSuccessPassback(http.StatusNoContent, nil)
	return
}

/*  THIS logic is 'similar' but not identical to CreateAction.  So if you change one, change the other!
Ok, b.c there is NO dependency, this is pretty easy.
1) we already have a new action
  we are going to create action params (stateComponentFilter) by using EXPLICIT xnames (loop through devices)

Every device / target is an operation.  If there is an ImageID, then GREAT!,we can do it.  If there is not image ID, then its a NoSol.
For everything WITH an imageID, need to see if its a valid op (vs noOp)

3) actions.params.stateComponents.xnames = devices.xnames
4) refil hsmDataMap
5) Generate the operations that I can
6) mark the action as ready?
7) save it back to disc
*/

func RestoreSnapshot(action storage.Action, snapshot storage.Snapshot) {
	//flush out the action params
	for _, device := range snapshot.Devices {
		action.Parameters.StateComponentFilter.Xnames = append(action.Parameters.StateComponentFilter.Xnames, device.Xname)
	}

	emptyStringSlice := []string{}

	hsmDataMap, errs := (*GLOB.HSM).FillHSMData(action.Parameters.StateComponentFilter.Xnames,
		emptyStringSlice,
		emptyStringSlice,
		emptyStringSlice)

	if len(errs) > 0 {
		logrus.Error(errs)
		for _, value := range errs {
			action.Errors = append(action.Errors, value.Error())
		}
	}
	err := (*GLOB.DSP).StoreAction(action)
	if err != nil {
		logrus.Error(err)
	}

	var XnameTargets []hsm.XnameTarget

	for _, device := range snapshot.Devices {
		for _, targets := range device.Targets {
			tar := hsm.XnameTarget{
				Xname:  device.Xname,
				Target: targets.Name,
			}
			XnameTargets = append(XnameTargets, tar)
		}
	}

	XnameTargetHSMMap := make(map[hsm.XnameTarget]hsm.HsmData)
	for key, value := range XnameTargets {
		XnameTargetHSMMap[XnameTargets[key]] = hsmDataMap[value.Xname]
	}

	//TODO Perhaps move THIS to a global? Not going to do it In June of 2020 b.c it works and
	//  I dont want to spend the time monkeying around with it!
	specialTargets := make(map[string]string)
	specialTargets["node0.bios"] = "/redfish/v1/Systems/Node0"
	specialTargets["node1.bios"] = "/redfish/v1/Systems/Node1"

	(*GLOB.HSM).RefillModelRF(&XnameTargetHSMMap, specialTargets)

	candidateOperations := make(map[uuid.UUID]storage.Operation)

	for _, device := range snapshot.Devices {
		hData := hsmDataMap[device.Xname]
		for _, targets := range device.Targets {
			op := storage.NewOperation()
			op.Xname = device.Xname
			op.Target = targets.Name
			op.ActionID = action.ActionID
			op.AutomaticallyGenerated = true
			//op.FromImageID -> need to fill by scanning!
			//op.FromFirmwareVersion -> need to fill by scanning!
			op.ToImageID = targets.ImageID
			op.HsmData = hData
			op.Manufacturer = hData.Manufacturer
			op.Model = hData.Model
			candidateOperations[op.OperationID] = *op
		}
	}

	//load all the FromFirmwareVerions! into the operation by looping through the deviceMap
	// I am intentionally getting this for ALL Targets b/c of the recursion needed for operations may need this data.
	// I think this is the lesser of two evils, to get a bit more data, that I may need, then to do a very expensive query MANY times!
	deviceMap, errlist := GetCurrentFirmwareVersionsFromHsmDataAndTargets(XnameTargetHSMMap)
	logrus.Error("ERROR LIST")
	logrus.Error(errlist)
	action.Errors = append(action.Errors, errlist...)
	err = (*GLOB.DSP).StoreAction(action)
	if err != nil {
		logrus.Error(err)
	}

	//6b -> get all images
	imageMap := GetImageMap()

	for operationID, _ := range candidateOperations {
		operation := candidateOperations[operationID]
		if operation.State.Can("configure") {
			//set up a few times
			operation.StartTime.Scan(time.Now()) //The clock starts NOW, set the expiration time!
			if snapshot.ExpirationTime.Valid {
				operation.ExpirationTime.Scan(snapshot.ExpirationTime.Time)
			}
			// load the FromFirmwareVersion -> so I can look up the FromFirmwareID
			SetFirmwareVersion(&operation, &deviceMap)
			//if candidateOperation
			//try to lookup the actual images IDs!
			FillInImageId(&operation, &imageMap, action.Parameters)

			// SetNoSolutionOperations!
			//  At this point, every candidate operation should have a ToImageID; if it doesnt then END IT!
			SetNoSolOp(&operation)

			//SetNoOperationOperations!
			SetNoOpOp(&operation, false)

			//Not a NoSOl nor a NoOP
			if operation.State.Can("configure") {
				//TODO here is a good place to add in DependencyManagment -> see develop @ tag: with-dependency to see what we had.
				operation.State.Event("configure")
			}
		}
	}

	//Start or Finish the Action!
	if len(candidateOperations) == 0 {
		action.EndTime.Scan(time.Now())
		action.State.Event("finish")
	} else {
		if action.State.Can("configure") { //if it cant start its because it got kicked out!
			action.State.Event("configure")
		}
	}

	//Figure out if there are any sibling blockers (xname == xname)
	//store the operations and load the OperationIDs into the action
	xnameOps := make(map[string][]uuid.UUID)
	for k, v := range candidateOperations {
		if v.State.Is("configured") {
			if xnameOp, ok := xnameOps[v.Xname]; ok {
				//there is at least one other entry!
				lastOp := xnameOp[len(xnameOps)-1]

				v.BlockedBy = append(v.BlockedBy, lastOp)
				v.State.Event("block")
				v.StateHelper = "blocked by sibling"
				candidateOperations[k] = v

			} else { //else, we are the first one here, so create it
				xnameOps[v.Xname] = append(xnameOps[v.Xname], v.OperationID)
			}
		}

		//regardless of that state, save it to the action
		action.OperationIDs = append(action.OperationIDs, k)
		err := (*GLOB.DSP).StoreOperation(v)
		if err != nil {
			logrus.Error(err)
		}
	}

	//Store the action
	(*GLOB.DSP).StoreAction(action)
}
