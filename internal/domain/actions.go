/*
 * MIT License
 *
 * (C) Copyright [2020-2022] Hewlett Packard Enterprise Development LP
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

	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/presentation"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TriggerFirmwareUpdate - will construct an ID, and async fire off work
func TriggerFirmwareUpdate(params storage.ActionParameters) (pb model.Passback) {

	logrus.Debug("parameters:", params)

	//at this point, the parameters can be validated
	err := ValidateActionParameters(&params)
	if err != nil {
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		return
	}

	action := storage.NewAction(params)

	// GenerateOperations may find out that an action param is invalid! Like an xname doesnt exist.
	// I am planning on allowing the action to be created, but to just no act on bad data, this is best effort!

	err = StoreAction(*action)
	if err == nil {
		actionID := storage.ActionID{ActionID: action.ActionID}

		CAP := presentation.CreateActionPayload{
			ActionID:       action.ActionID,
			OverrideDryrun: params.Command.OverrideDryrun,
		}

		pb = model.BuildSuccessPassback(http.StatusAccepted, CAP)
		go GenerateOperations(actionID.ActionID) //async fire off
	} else {
		pb = model.BuildSuccessPassback(http.StatusBadRequest, err)
	}

	return pb
}

func StoreAction(action storage.Action) (err error) {
	// Get the current state of the stored action to see if
	// it has been signaled to stop
	curStAction, curExists := GetStoredAction(action.ActionID)
	if curExists == nil {
		if curStAction.State.Is("abortSignaled") {
			// Change to signal abort if possible
			if action.State.Can("signalAbort") == true {
				logrus.Info("Changed State from " + action.State.Current() + " to abortSignaled")
				action.State.Event("signalAbort")
			}
		}
	}
	err = (*GLOB.DSP).StoreAction(action)
	return err
}

func GetStoredActions() (actions []storage.Action, err error) {
	actions, err = (*GLOB.DSP).GetActions()
	return
}

func GetStoredAction(actionID uuid.UUID) (action storage.Action, err error) {
	action, err = (*GLOB.DSP).GetAction(actionID)
	return
}

func DeleteStoredAction(actionID uuid.UUID) (err error) {
	err = (*GLOB.DSP).DeleteAction(actionID)
	return
}

func StoreOperation(operation storage.Operation) (err error) {
	err = (*GLOB.DSP).StoreOperation(operation)
	return
}

func GetStoredOperations(actionID uuid.UUID) (operations []storage.Operation, err error) {
	operations, err = (*GLOB.DSP).GetOperations(actionID)
	return
}

func GetStoredOperation(operationID uuid.UUID) (operation storage.Operation, err error) {
	operation, err = (*GLOB.DSP).GetOperation(operationID)
	return
}

func DeleteStoredOperation(operationID uuid.UUID) (err error) {
	err = (*GLOB.DSP).DeleteOperation(operationID)
	return
}

func GetAllActiveOperationsFromAction(actionID uuid.UUID) []storage.Operation {
	operations, err := GetStoredOperations(actionID)
	if err != nil {
		logrus.Error(err)
	}

	var operationList []storage.Operation
	for _, operation := range operations {
		if operation.State.Is("configured") || operation.State.Is("inProgress") || operation.State.Is("needsVerified") || operation.State.Is("verifying") {
			(*GLOB.HSM).RestoreCredentials(&operation.HsmData)
			operationList = append(operationList, operation)
		}
	}
	return operationList
}

func GetAllBlockedOperationsFromAction(actionID uuid.UUID) (operationList []storage.Operation) {
	//	operations, err := (*GLOB.DSP).Get//Operations(actionID)
	//	if err != nil {
	//		logrus.Error(err)
	//	}
	//
	//	var operationList []storage.Operation
	//	for _, operation := range operations {
	//		if operation.State.Is("blocked") {
	//			operationList = append(operationList, operation)
	//		}
	//	}
	return operationList

}

// CheckBlockage scans all blockedOperations and checks if its blocker is completed,
// if it is then it updates the state to configured
func CheckBlockage(allOperations *[]storage.Operation) {

	type OpState struct {
		Op    *storage.Operation
		State string
	}
	ops := make(map[uuid.UUID]OpState)

	for operationID, _ := range *allOperations {
		var tmpOpState OpState
		operation := (*allOperations)[operationID]
		if operation.State.Is("failed") || operation.State.Is("aborted") || operation.State.Is("noOperation") || operation.State.Is("noSolution") || operation.State.Is("succeeded") {
			tmpOpState = OpState{
				Op:    &operation,
				State: "completed",
			}
		} else if operation.State.Is("blocked") {
			tmpOpState = OpState{
				Op:    &operation,
				State: "blocked",
			}
		} else {
			tmpOpState = OpState{
				Op:    &operation,
				State: "probablyInProgress",
			}
		}
		ops[operation.OperationID] = tmpOpState
	}

	for opID, op := range ops {
		if op.State == "blocked" {
			var stillBlocked bool = false
			for _, blocker := range op.Op.BlockedBy {
				blockerOp := ops[blocker]
				if blockerOp.State != "completed" {
					stillBlocked = true
					break
				}
			}
			if !stillBlocked {
				op.State = "unblock"
				op.Op.StateHelper = "unblocked"
				op.Op.State.Event("unblock")
				StoreOperation(*op.Op)
			}
			ops[opID] = op
		}
	}
}

func GetAllNonCompleteNonInitialActions() []storage.Action {
	actions, err := GetStoredActions()
	if err != nil {
		logrus.Error(err)
	}

	var actionList []storage.Action
	for _, action := range actions {
		if !action.State.Is("aborted") && !action.State.Is("completed") && !action.State.Is("new") {
			actionList = append(actionList, action)
		}
	}

	return actionList
}

func GetAllAbortSignaledActions() []storage.Action {
	actions, err := GetStoredActions()
	if err != nil {
		logrus.Error(err)
	}

	var actionList []storage.Action
	for _, action := range actions {
		if action.State.Is("abortSignaled") {
			actionList = append(actionList, action)
		}
	}

	return actionList
}

// GetAllActions - gets all actions from the system formatting them as a summary
func GetAllActions() (pb model.Passback) {
	actions, err := GetStoredActions()
	summaries := presentation.ActionSummaries{Actions: []presentation.ActionSummary{}}

	//Now convert an actions into an actions summary!
	if err == nil {
		for _, action := range actions {
			summary, err := presentation.ToActionSummaryFromAction(action)
			if err != nil {
				logrus.WithField("error", err).Error("Could not convert from action to action summary")
				break
			}
			operations, _ := GetStoredOperations(action.ActionID)
			if err != nil {
				logrus.WithFields(logrus.Fields{"ERROR": err, "actionID": action.ActionID.String()}).Error("Could not get operations from action")
				break
			}

			operationCounts, err := presentation.ToOperationCountsFromOperations(operations)
			if err != nil {
				logrus.WithFields(logrus.Fields{"ERROR": err, "actionID": action.ActionID.String()}).Error("Could not build operation data")
				break
			}
			summary.OperationCounts = operationCounts
			summaries.Actions = append(summaries.Actions, summary)
		}

		pb = model.BuildSuccessPassback(http.StatusOK, summaries)
	} else {
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
	}
	return pb
}

func GetActionStatus(actionID uuid.UUID) (pb model.Passback) {
	action, err := GetStoredAction(actionID)
	if err == nil {
		summary, err := presentation.ToActionSummaryFromAction(action)
		if err != nil {
			logrus.WithField("error", err).Error("Could not convert from action to action summary")
			pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
			return
		}
		operations, _ := GetStoredOperations(action.ActionID)
		if err != nil {
			logrus.WithFields(logrus.Fields{"ERROR": err, "actionID": action.ActionID.String()}).Error("Could not get operations from action")
			pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
			return
		}

		operationCounts, err := presentation.ToOperationCountsFromOperations(operations)
		if err != nil {
			logrus.WithFields(logrus.Fields{"ERROR": err, "actionID": action.ActionID.String()}).Error("Could not build operation data")
			pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
			return
		}
		summary.OperationCounts = operationCounts
		pb = model.BuildSuccessPassback(http.StatusOK, summary)
	} else {
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
	}
	return pb
}

func GetAllOperationsFromAction(actionID uuid.UUID) (operations []storage.Operation, err error) {
	operations, err = GetStoredOperations(actionID)
	for k, v := range operations {
		(*GLOB.HSM).RestoreCredentials(&v.HsmData)
		operations[k] = v
	}
	return
}

// GetAction - gets an action and formats it for presentation
func GetAction(id uuid.UUID) (pb model.Passback) {
	action, err := GetStoredAction(id)
	actionMarshal := presentation.ActionMarshaled{}

	if err == nil {
		actionMarshal, err = presentation.ToActionMarshaledFromAction(action)
		if err != nil {
			logrus.WithField("error", err).Error("Could not convert from action to action summary")
		} else {
			operations, err := GetStoredOperations(action.ActionID)
			if err != nil {
				logrus.WithFields(logrus.Fields{"ERROR": err, "actionID": action.ActionID.String()}).Error("Could not get operations from action")
			} else {
				operationSummary, err := presentation.ToOperationSummaryFromOperations(operations)
				if err != nil {
					logrus.WithFields(logrus.Fields{"ERROR": err, "actionID": action.ActionID.String()}).Error("Could not build operation data")
				}
				actionMarshal.OperationSummary = operationSummary
				pb = model.BuildSuccessPassback(http.StatusOK, actionMarshal)
				return pb
			}
		}
	}
	pb = model.BuildErrorPassback(http.StatusNotFound, err)
	return pb
}

// GetActionDetail - gets an action and formats it for presentation
func GetActionDetail(id uuid.UUID) (pb model.Passback) {
	action, err := GetStoredAction(id)
	actionOperationsDetail := presentation.ActionOperationsDetail{}

	if err == nil {
		actionOperationsDetail, err = presentation.ToActionOperationsDetailFromAction(action)
		if err != nil {
			logrus.WithField("error", err).Error("Could not convert from action to action summary")
		} else {
			operations, err := GetStoredOperations(action.ActionID)
			if err != nil {
				logrus.WithFields(logrus.Fields{"ERROR": err, "actionID": action.ActionID.String()}).Error("Could not get operations from action")
			} else {
				var operationsPI []presentation.OperationPlusImages
				for _, o := range operations {
					var opi presentation.OperationPlusImages
					opi.Operation = o
					opi.FromImage, _ = GetStoredImage(o.FromImageID)
					opi.ToImage, _ = GetStoredImage(o.ToImageID)
					operationsPI = append(operationsPI, opi)
				}
				operationDetail, err := presentation.ToOperationDetailFromOperations(operationsPI)
				if err != nil {
					logrus.WithFields(logrus.Fields{"ERROR": err, "actionID": action.ActionID.String()}).Error("Could not build operation data")
				}
				actionOperationsDetail.OperationDetails = operationDetail
				pb = model.BuildSuccessPassback(http.StatusOK, actionOperationsDetail)
				return pb
			}
		}
	}
	pb = model.BuildErrorPassback(http.StatusNotFound, err)
	return pb
}

func GetActionState(id uuid.UUID) (action storage.Action) {
	action, _ = GetStoredAction(id)
	return
}

func GetOperationSummaryFromAction(actionID uuid.UUID) (operationCounts presentation.OperationCounts) {
	operations, err := GetStoredOperations(actionID)
	if err != nil {
		logrus.WithFields(logrus.Fields{"ERROR": err, "actionID": actionID.String()}).Error("Could not get operations from action")
	} else {
		operationCounts, err = presentation.ToOperationCountsFromOperations(operations)
		if err != nil {
			logrus.WithFields(logrus.Fields{"ERROR": err, "actionID": actionID.String()}).Error("Could not build operation data")
		}

	}
	return operationCounts
}

// DeleteAction - deletes a non running action
func DeleteAction(id uuid.UUID) (pb model.Passback) {
	action, err := GetStoredAction(id)
	if err != nil {
		logrus.Error(err)
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
		return pb
	}

	// Only prevent if its aborting or running
	if action.State.Current() == "abortSignaled" || action.State.Current() == "running" {
		err = errors.New("cannot delete a currently running action")
		logrus.Error(err)
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		return pb
	}

	// Delete Action's Operations
	operations, err := GetStoredOperations(id)
	for _, operation := range operations {
		_ = DeleteStoredOperation(operation.OperationID)
	}
	err = DeleteStoredAction(id)
	if err == nil {
		pb = model.BuildSuccessPassback(http.StatusNoContent, nil)
		return pb
	}

	pb = model.BuildErrorPassback(http.StatusBadRequest, err)
	return pb
}

// GetActionOperationID - gets the operation for an action
func GetActionOperationID(actionID uuid.UUID, operationID uuid.UUID) (pb model.Passback) {
	if actionID != uuid.Nil {
		_, err := GetStoredAction(actionID)
		if err != nil {
			pb = model.BuildErrorPassback(http.StatusNotFound, err)
			return
		}
	}

	operation, err := GetStoredOperation(operationID)
	if err != nil {
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
		return
	}

	operationMarshal := presentation.OperationMarshaled{}
	operationMarshal, err = presentation.ToOperationMarshaledFromOperation(operation)
	if err != nil {
		logrus.WithField("error", err).Error("Could not transform operation")
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		return
	}
	toimage, err := GetStoredImage(operationMarshal.ToImageID)
	if err == nil {
		operationMarshal.ToFirmwareVersion = toimage.FirmwareVersion
		operationMarshal.ToSemanticFirmwareVersion = toimage.SemanticFirmwareVersion.String()
	}
	fromimage, err := GetStoredImage(operationMarshal.FromImageID)
	if err == nil {
		operationMarshal.FromSemanticFirmwareVersion = fromimage.SemanticFirmwareVersion.String()
	}

	pb = model.BuildSuccessPassback(http.StatusOK, operationMarshal)
	return pb
}

// AbortActionID - halt a running action
func AbortActionID(actionID uuid.UUID) (pb model.Passback) {
	action, err := GetStoredAction(actionID)
	if err != nil {
		logrus.Error(err)
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
		return pb
	}

	if action.State.Can("signalAbort") == false {
		logrus.Trace("already complete")
		pb = model.BuildSuccessPassback(http.StatusOK, nil)
		return pb
	} else {
		action.State.Event("signalAbort")
		if err != nil {
			pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
			return pb
		}
		err = StoreAction(action)
		if err != nil {
			pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		} else {
			pb = model.BuildSuccessPassback(http.StatusAccepted, nil)
		}
	}
	return pb
}

//// AbortOperation - halt a running operation
//func AbortOperation(operationID uuid.UUID) (err error) {
//	operation, err := (*GLOB.DSP).GetOperation(operationID)
//	if err != nil {
//		logrus.Error(err)
//		return err
//	}
//
//	//if it cannot abort; it must be finished
//	if operation.State.Can("abort") {
//		err = operation.State.Event("abort")
//		if err != nil {
//			logrus.Error(err)
//			return err
//		}
//		operation.EndTime.Scan(time.Now())
//		err = (*GLOB.DSP).StoreOperation(operation)
//		if err != nil {
//			logrus.Error(err)
//			return err
//		}
//	}
//	return nil
//}
