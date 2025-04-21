/*
 * MIT License
 *
 * (C) Copyright [2020-2021,2024-2025] Hewlett Packard Enterprise Development LP
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
	"strconv"
	"strings"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-firmware-action/internal/domain"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/presentation"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// GetSnapshots - returns all snapshot summaries
func GetSnapshots(w http.ResponseWriter, req *http.Request) {

	defer base.DrainAndCloseRequestBody(req)

	pb := domain.GetSnapshots()
	WriteHeaders(w, pb)
	return
}

// GetSnapshot - returns a snapshot by name
func GetSnapshot(w http.ResponseWriter, req *http.Request) {

	defer base.DrainAndCloseRequestBody(req)

	params := mux.Vars(req)
	name, _ := params["name"]
	pb := domain.GetSnapshot(name)
	WriteHeaders(w, pb)
	return
}

// CreateSnapshot - record a snapshot of the system
func CreateSnapshot(w http.ResponseWriter, req *http.Request) {

	defer base.DrainAndCloseRequestBody(req)

	var parameters presentation.SnapshotParametersMarshaled
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
			pb := model.BuildErrorPassback(http.StatusBadRequest, err)
			logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("Unparseable json")
			WriteHeaders(w, pb)
			return
		}

		sparams := parameters.ToSnapshotParameters()

		err = domain.ValidateSnapshotParameters(&sparams)
		if err != nil {
			pb := model.BuildErrorPassback(http.StatusBadRequest, err)
			WriteHeaders(w, pb)
			return
		}

		pb := domain.CreateSnapshot( sparams)
		if pb.IsError == false {
			location := "../snapshots/" + (pb.Obj.(presentation.SnapshotName).Name)
			WriteHeadersWithLocation(w, pb, location)
			return
		}
		WriteHeaders(w, pb)

	} else {
		err := errors.New("empty body not allowed")
		pb := model.BuildErrorPassback(http.StatusBadRequest, err)
		logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("empty body")
		WriteHeaders(w, pb)
		return
	}

	return
}

// StartRestoreSnapshot - triggers an action; by restoring a snapshot. Returns and actionID
func StartRestoreSnapshot(w http.ResponseWriter, req *http.Request) {

	defer base.DrainAndCloseRequestBody(req)

	var pb model.Passback
	//Check if the snapshot doesnt exist
	params := mux.Vars(req)
	name, _ := params["name"]

	//verify that confirm is set to YES, then START
	confirm, ok := req.URL.Query()["confirm"]
	if !ok || strings.ToUpper(confirm[0]) != strings.ToUpper("yes") {
		err := errors.New("missing required parameter 'confirm'")
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		WriteHeaders(w, pb)
		return
	}

	//by default make overrideDryrun true //TODO CASMHMS-3642 -> change this logic; maybe the name needs to be changed.
	overrideDryrun := false
	dryrun_p, ok := req.URL.Query()["overrideDryrun"]
	if ok && strings.ToUpper(dryrun_p[0]) == strings.ToUpper("true") {
		overrideDryrun = true
	}

	timeLimit := 0
	timeLimit_p, ok := req.URL.Query()["timeLimit"]
	if ok {
		var err error
		timeLimit, err = strconv.Atoi(timeLimit_p[0])
		if err != nil {
			pb = model.BuildErrorPassback(http.StatusBadRequest, err)
			WriteHeaders(w, pb)
			return
		}
	}

	pb = domain.StartRestoreSnapshot(name, overrideDryrun, timeLimit)
	if pb.IsError == false {
		location := "../actions/" + (pb.Obj.(presentation.CreateActionPayload).ActionID.String())
		WriteHeadersWithLocation(w, pb, location)
		return
	}
	WriteHeaders(w, pb)
	return
}


// DeleteSnapshot - delete a system snapshot
func DeleteSnapshot(w http.ResponseWriter, req *http.Request) {

	defer base.DrainAndCloseRequestBody(req)

	params := mux.Vars(req)
	name, _ := params["name"]
	pb := domain.DeleteSnapshot(name)
	WriteHeaders(w, pb)
	return
}
