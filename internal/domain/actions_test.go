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
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/presentation"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/storage"
)

type Actions_TS struct {
	suite.Suite
}

// SetupSuit is run ONCE
func (suite *Actions_TS) SetupSuite() {
}

func (suite *Actions_TS) Test_TriggerFirmwareUpdate() {
	parameters := storage.ActionParameters{
		Command: storage.Command{
			OverrideDryrun: false,
		},
	}

	pb := TriggerFirmwareUpdate(parameters)
	suite.False(pb.IsError)

	logrus.Trace(pb.Obj.(presentation.CreateActionPayload), pb.StatusCode)
	ActionID := pb.Obj.(presentation.CreateActionPayload).ActionID
	//TODO FIX THIS::: we might be getting the ActionID too soon
	time.Sleep(5 * time.Second)

	for {
		aRet, err := (*GLOB.DSP).GetAction(ActionID)
		suite.True(err == nil)
		// TODO: State should go to complete when complete, but does not yet
		aRet.State.SetState("completed")
		err = (*GLOB.DSP).StoreAction(aRet)
		suite.True(err == nil)
		fmt.Println(aRet.State.Current())
		if aRet.State.Current() == "completed" {
			break
		}
	}
	err := (*GLOB.DSP).DeleteAction(ActionID)
	suite.True(err == nil)
}

func (suite *Actions_TS) Test_GetAllActions_Simple() {
	pb := GetAllActions()
	suite.False(pb.IsError)
}

func (suite *Actions_TS) Test_GetAllActions_withUpdate() {
	parameters := storage.ActionParameters{
		Command: storage.Command{
			OverrideDryrun: false,
		},
	}

	pb := TriggerFirmwareUpdate(parameters)
	suite.False(pb.IsError)
	logrus.Trace(pb)

	logrus.Trace(pb.Obj.(presentation.CreateActionPayload), pb.StatusCode)
	ActionID := pb.Obj.(presentation.CreateActionPayload).ActionID

	pb = GetAllActions()
	suite.False(pb.IsError)
	asm := pb.Obj.(presentation.ActionSummaries)
	logrus.Trace(asm)
	count := 0
	for _, action := range asm.Actions {
		if action.ActionID == ActionID {
			aRet, err := (*GLOB.DSP).GetAction(action.ActionID)
			suite.True(err == nil)
			aRet.State.SetState("completed") // Mark as completed so we can delete
			err = (*GLOB.DSP).StoreAction(aRet)
			count++
		}
	}
	suite.True(count == 1)
	err := (*GLOB.DSP).DeleteAction(ActionID)
	suite.True(err == nil)
}

func (suite *Actions_TS) Test_GetAction_BadID() {
	pb := GetAction(uuid.New())
	suite.True(pb.IsError)
}

func (suite *Actions_TS) Test_GetAction_Good() {
	action := storage.HelperGetStockAction()
	action.State.SetState("completed")
	err := (*GLOB.DSP).StoreAction(action)
	suite.True(err == nil)

	pb := GetAction(action.ActionID)
	suite.False(pb.IsError)
	aMarsh := pb.Obj.(presentation.ActionMarshaled)
	aRet, err := (*GLOB.DSP).GetAction(aMarsh.ActionID)
	suite.True(err == nil)
	suite.True(aRet.Equals(action))

	err = (*GLOB.DSP).DeleteAction(action.ActionID)
	suite.True(err == nil)
}

func (suite *Actions_TS) Test_DeleteAction_Good() {
	action := storage.HelperGetStockAction()
	action.State.SetState("completed") // Mark as completed so we can delete
	err := (*GLOB.DSP).StoreAction(action)
	suite.True(err == nil)

	aRet, err := (*GLOB.DSP).GetAction(action.ActionID)
	suite.True(err == nil)
	suite.True(action.Equals(aRet))

	pb := DeleteAction(action.ActionID)
	suite.False(pb.IsError)

	_, err = (*GLOB.DSP).GetAction(action.ActionID)
	suite.False(err == nil)
}

func (suite *Actions_TS) Test_DeleteAction_BadID() {
	pb := DeleteAction(uuid.New())
	suite.True(pb.IsError)
	suite.Equal(pb.StatusCode, http.StatusNotFound)
}

func (suite *Actions_TS) Test_GetActionOperationID_Good() {
	pb := GetActionOperationID(uuid.New(), uuid.New())
	logrus.Error(pb)
	suite.True(pb.IsError)
	suite.Equal(pb.StatusCode, http.StatusNotFound)

	action := storage.HelperGetStockAction()
	actionID := action.ActionID
	operation := storage.HelperGetStockOperation()
	operationID := operation.OperationID
	action.OperationIDs = append(action.OperationIDs, operationID)
	err := (*GLOB.DSP).StoreOperation(operation)
	suite.True(err == nil)
	err = (*GLOB.DSP).StoreAction(action)
	suite.True(err == nil)

	pb = GetActionOperationID(actionID, uuid.New())
	logrus.Error(pb)
	suite.True(pb.IsError)
	suite.Equal(pb.StatusCode, http.StatusNotFound)

	pb = GetActionOperationID(actionID, operationID)
	logrus.Error(pb)
	suite.False(pb.IsError)

	pb = GetActionOperationID(uuid.Nil, operationID)
	logrus.Error(pb)
	suite.False(pb.IsError)
}

func (suite *Actions_TS) Test_AbortActionID_Good() {
	action := storage.HelperGetStockAction()
	err := (*GLOB.DSP).StoreAction(action)
	suite.True(err == nil)

	pb := AbortActionID(action.ActionID)
	suite.False(pb.IsError)

	aRet, err := (*GLOB.DSP).GetAction(action.ActionID)
	suite.True(err == nil)
	logrus.Error(aRet.State.Current())
	suite.True(aRet.State.Current() == "abortSignaled")
}

func (suite *Actions_TS) Test_AbortActionID_BadID() {
	pb := AbortActionID(uuid.New())
	suite.True(pb.IsError)
	suite.Equal(pb.StatusCode, http.StatusNotFound)
}

//func (suite *Actions_TS) Test_AbortOperation() {
//	operation := storage.HelperGetStockOperation()
//	err := (*GLOB.DSP).StoreOperation(operation)
//	suite.True(err == nil)
//
//	err = AbortOperation(operation.OperationID)
//	suite.True(err == nil)
//
//	oRet, err := (*GLOB.DSP).GetOperation(operation.OperationID)
//	suite.True(err == nil)
//	logrus.Error(oRet.State.Current())
//	suite.True(oRet.State.Current() == "aborted")
//}

func Test_Domain_Actions(t *testing.T) {
	//This setups the production routs and handler
	ConfigureSystemForUnitTesting()
	suite.Run(t, new(Actions_TS))
}
