/*
 * MIT License
 *
 * (C) Copyright [2020-2021,2024] Hewlett Packard Enterprise Development LP
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

package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/Cray-HPE/hms-firmware-action/internal/domain"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/presentation"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// CreateAction - creates an action and will trigger an 'update'
func CreateAction(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	var pb model.Passback
	var parameters storage.ActionParameters
	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		logrus.WithFields(logrus.Fields{"body": string(body)}).Trace("Printing request body")

		if err != nil {
			pb := model.BuildErrorPassback(http.StatusInternalServerError, err)
			logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("Error detected retrieving body")
			WriteHeaders(w, pb)
			return
		}

		err = json.Unmarshal(body, &parameters)
		if err != nil {
			pb = model.BuildErrorPassback(http.StatusBadRequest, err)
			logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("Unparseable json")
			WriteHeaders(w, pb)
			return
		}
	} else {
		err := errors.New("empty body not allowed")
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("empty body")
		WriteHeaders(w, pb)
		return
	}

	pb = domain.TriggerFirmwareUpdate(parameters)

	if pb.IsError == false {
		location := "../actions/" + (pb.Obj.(presentation.CreateActionPayload).ActionID.String())

		WriteHeadersWithLocation(w, pb, location)
	} else {
		WriteHeaders(w, pb)
	}
	return
}

// GetAction - returns all actions, or action by actionID
func GetAction(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	var pb model.Passback
	params := mux.Vars(req)
	//If actionID is not in the params, then do ALL
	if _, ok := params["actionID"]; ok {
		//parse uuid and if its good then call getUpdates
		pb = GetUUIDFromVars("actionID", req)
		if pb.IsError {
			WriteHeaders(w, pb)
			return
		}
		actionID := pb.Obj.(uuid.UUID)
		pb = domain.GetAction(actionID)

	} else {
		pb = domain.GetAllActions()
	}
	WriteHeaders(w, pb)
	return
}

// GetActionIDStatus - get action with summary operations
func GetActionIDStatus(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	//If actionID is not in the params, then do ALL
	pb := GetUUIDFromVars("actionID", req)
	if pb.IsError {
		WriteHeaders(w, pb)
		return
	}
	actionID := pb.Obj.(uuid.UUID)
	pb = domain.GetActionStatus(actionID)
	WriteHeaders(w, pb)
	return
}

// GetActionIDOperations - get action with detailed operations
func GetActionIDOperations(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	pb := GetUUIDFromVars("actionID", req)
	if pb.IsError {
		WriteHeaders(w, pb)
		return
	}
	actionID := pb.Obj.(uuid.UUID)
	pb = domain.GetActionDetail(actionID)
	WriteHeaders(w, pb)
	return
}

// DeleteAction - delete action by actionID
func DeleteAction(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	pb := GetUUIDFromVars("actionID", req)
	if pb.IsError {
		WriteHeaders(w, pb)
		return
	}
	pb = domain.DeleteAction(pb.Obj.(uuid.UUID))
	WriteHeaders(w, pb)
	return
}

// AbortActionID - halt a running action
func AbortActionID(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	pb := GetUUIDFromVars("actionID", req)
	if pb.IsError {
		WriteHeaders(w, pb)
		return
	}
	actionID := pb.Obj.(uuid.UUID)

	pb = domain.AbortActionID(actionID)
	WriteHeaders(w, pb)
	return
}

// GetActionOperationID - get an operation by action/operation ids
func GetActionOperationID(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	pb := GetUUIDFromVars("actionID", req)
	if pb.IsError {
		WriteHeaders(w, pb)
		return
	}
	actionID := pb.Obj.(uuid.UUID)

	pb = GetUUIDFromVars("operationID", req)
	if pb.IsError {
		WriteHeaders(w, pb)
		return
	}
	operationID := pb.Obj.(uuid.UUID)

	pb = domain.GetActionOperationID(actionID, operationID)
	WriteHeaders(w, pb)
	return
}

// GetOperationID - get an operation by operation id
// NOTE: Returns same structure as GetActionOperationID
// Since operation ids are unique, action id
// is not needed to locate operation.
func GetOperationID(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	actionID := uuid.Nil

	pb := GetUUIDFromVars("operationID", req)
	if pb.IsError {
		WriteHeaders(w, pb)
		return
	}
	operationID := pb.Obj.(uuid.UUID)

	pb = domain.GetActionOperationID(actionID, operationID)
	WriteHeaders(w, pb)
	return
}
