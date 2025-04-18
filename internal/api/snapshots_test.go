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
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-firmware-action/internal/domain"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/presentation"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type Snapshot_TS struct {
	suite.Suite
}

// SetupSuit is run ONCE
func (suite *Snapshot_TS) SetupSuite() {
}

//GET Snapshots

func (suite *Snapshot_TS) Test_GET_Snapshots_HappyPath_PresumedEmpty() {
	r, _ := http.NewRequest("GET", "/snapshots", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer base.DrainAndCloseResponseBody(resp)
	suite.Equal(resp.StatusCode, http.StatusOK)
	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	returnedSnapshots := presentation.SnapshotSummaries{}
	_ = json.Unmarshal(body, &returnedSnapshots)

	pb := domain.GetSnapshots()
	expectedSnapshots := pb.Obj.(presentation.SnapshotSummaries)
	logrus.Debug(expectedSnapshots)
	logrus.Debug(returnedSnapshots)
	suite.True(expectedSnapshots.Equals(returnedSnapshots))
}

func (suite *Snapshot_TS) Test_GET_Snapshots_HappyPath_NonEmpty() {
	//create two snapshots
	emptyParam := storage.SnapshotParameters{}
	s1 := emptyParam
	s2 := emptyParam
	s1.Name = strconv.Itoa(rand.Int())
	s2.Name = strconv.Itoa(rand.Int())

	ss1 := domain.CreateSnapshot(s1)
	ss2 := domain.CreateSnapshot(s2)

	//check that the snapshots are equal
	r, _ := http.NewRequest("GET", "/snapshots", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer base.DrainAndCloseResponseBody(resp)
	suite.Equal(resp.StatusCode, http.StatusOK)
	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	returnedSnapshots := presentation.SnapshotSummaries{}
	_ = json.Unmarshal(body, &returnedSnapshots)

	pb := domain.GetSnapshots()
	expectedSnapshots := pb.Obj.(presentation.SnapshotSummaries)
	logrus.Debug(expectedSnapshots)
	logrus.Debug(returnedSnapshots)
	suite.True(expectedSnapshots.Equals(returnedSnapshots))

	//Clean up
	domain.DeleteSnapshot(ss1.Obj.(presentation.SnapshotName).Name)
	domain.DeleteSnapshot(ss2.Obj.(presentation.SnapshotName).Name)
}

// GET Snapshot(name)

func (suite *Snapshot_TS) Test_GET_SnapshotByName_NotFound() {
	r, _ := http.NewRequest("GET", "/snapshots/noExist", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer base.DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusNotFound, resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)

}

func (suite *Snapshot_TS) Test_GET_SnapshotByName_HappyPath() {
	//Create new snapshot
	emptyParam := GetDefaultSnapshotParameters()

	newSnapshot := domain.CreateSnapshot(emptyParam)
	newSnapshotName := newSnapshot.Obj.(presentation.SnapshotName).Name
	expectedSnapshot := domain.GetSnapshot(newSnapshotName).Obj.(presentation.SnapshotMarshaled)

	snapShotPath := filepath.Join("/snapshots", newSnapshotName)

	r, _ := http.NewRequest("GET", snapShotPath, nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer base.DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusOK, resp.StatusCode)
	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	returnedSnapshot := presentation.SnapshotMarshaled{}
	_ = json.Unmarshal(body, &returnedSnapshot)

	logrus.Debug(expectedSnapshot)
	logrus.Debug(returnedSnapshot)
	suite.True(expectedSnapshot.Equals(returnedSnapshot))

	//Clean up
	domain.DeleteSnapshot(expectedSnapshot.Name)
}

// CREATE Snapshot

func (suite *Snapshot_TS) Test_Create_Snapshot_HappyPath() {
	//Create new snapshot
	params := GetDefaultSnapshotParametersMarshaled()
	paramJson, _ := json.Marshal(params)
	paramString := string(paramJson)
	expectedSnapshotName := params.Name

	snapShotPath := filepath.Join("/snapshots")

	r, _ := http.NewRequest("POST", snapShotPath, strings.NewReader(paramString))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer base.DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusCreated, resp.StatusCode)
	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	returnedSnapshot := storage.Snapshot{}
	_ = json.Unmarshal(body, &returnedSnapshot)

	suite.Equal(expectedSnapshotName, returnedSnapshot.Name)

	//Clean up
	domain.DeleteSnapshot(expectedSnapshotName)
}

func (suite *Snapshot_TS) Test_Create_Snapshot_Duplicate() {

	emptyParam := GetDefaultSnapshotParameters()

	domain.CreateSnapshot(emptyParam)

	snapShotPath := filepath.Join("/snapshots")
	params := GetDefaultSnapshotParametersMarshaled()
	params.Name = emptyParam.Name
	paramJson, _ := json.Marshal(params)
	paramString := string(paramJson)

	r, _ := http.NewRequest("POST", snapShotPath, strings.NewReader(paramString))

	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer base.DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusConflict, resp.StatusCode)
	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)

	//Clean up
	domain.DeleteSnapshot(emptyParam.Name)
}

// DELETE Snapshot

func (suite *Snapshot_TS) Test_Delete_Snapshot_HappyPath() {
	//Create new snapshot
	emptyParam := storage.SnapshotParameters{}
	expectedSnapshotName := strconv.Itoa(rand.Int())
	emptyParam.Name = expectedSnapshotName
	_ = domain.CreateSnapshot(emptyParam)
	snapShotPath := filepath.Join("/snapshots", expectedSnapshotName)
	r, _ := http.NewRequest("DELETE", snapShotPath, nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer base.DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusNoContent, resp.StatusCode)
}

func (suite *Snapshot_TS) Test_Delete_Snapshot_NotFound() {
	expectedSnapshotName := strconv.Itoa(rand.Int())
	snapShotPath := filepath.Join("/snapshots", expectedSnapshotName)
	r, _ := http.NewRequest("DELETE", snapShotPath, nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer base.DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusNotFound, resp.StatusCode)
}

// RESTORE snapshot
//func (suite *Snapshot_TS) Test_Restore_Snapshot_HappyPath() {
//	//Create new snapshot
//	emptyParam := storage.SnapshotParameters{}
//
//	newSnapshot := domain.CreateSnapshot(strconv.Itoa(rand.Int()),emptyParam)
//	newSnapshotName := newSnapshot.Obj.(presentation.SnapshotName).Name
//	expectedSnapshot := domain.GetSnapshot(newSnapshotName).Obj.(presentation.SnapshotMarshaled)
//
//	snapShotPath := filepath.Join("/snapshots", newSnapshotName, "restore") + "?confirm=yes"
//
//	r, _ := http.NewRequest("POST", snapShotPath, nil)
//	w := httptest.NewRecorder()
//	NewRouter().ServeHTTP(w, r)
//	resp := w.Result()
//	defer base.DrainAndCloseResponseBody(resp)
//	suite.Equal(http.StatusAccepted, resp.StatusCode)
//	//read the body, unmarshall and turn into an application
//	body, _ := ioutil.ReadAll(resp.Body)
//	logrus.Debug(body)
//	usp := storage.ActionID{}
//	_ = json.Unmarshal(body, &usp)
//
//	suite.True(usp.ActionID != uuid.Nil)
//
//	//Clean up
//	domain.DeleteSnapshot(expectedSnapshot.Name)
//}

func (suite *Snapshot_TS) Test_Restore_Snapshot_NoExist() {
	//Create new snapshot
	newSnapshotName := strconv.Itoa(rand.Int())
	snapShotPath := filepath.Join("/snapshots", newSnapshotName, "restore") + "?confirm=yes"

	r, _ := http.NewRequest("POST", snapShotPath, nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer base.DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusNotFound, resp.StatusCode)
	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)
	logrus.Debug(problem)
}

func (suite *Snapshot_TS) Test_Restore_Snapshot_NoConfirm() {
	//Create new snapshot
	emptyParam := storage.SnapshotParameters{}
	newSnapshotName := strconv.Itoa(rand.Int())
	emptyParam.Name = newSnapshotName
	_ = domain.CreateSnapshot(emptyParam)
	expectedSnapshot := domain.GetSnapshot(newSnapshotName).Obj.(presentation.SnapshotMarshaled)

	snapShotPath := filepath.Join("/snapshots", newSnapshotName, "restore")

	r, _ := http.NewRequest("POST", snapShotPath, nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer base.DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)
	logrus.Debug(problem)
	suite.True(strings.Contains(problem.Detail, "missing required parameter"))

	//Clean up
	domain.DeleteSnapshot(expectedSnapshot.Name)
}

func (suite *Snapshot_TS) Test_Restore_Snapshot_BadConfirm() {
	//Create new snapshot
	emptyParam := GetDefaultSnapshotParameters()

	newSnapshot := domain.CreateSnapshot(emptyParam)
	newSnapshotName := newSnapshot.Obj.(presentation.SnapshotName).Name
	expectedSnapshot := domain.GetSnapshot(newSnapshotName).Obj.(presentation.SnapshotMarshaled)

	snapShotPath := filepath.Join("/snapshots", newSnapshotName, "restore") + "?confirm=please"

	r, _ := http.NewRequest("POST", snapShotPath, nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer base.DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)
	logrus.Debug(problem)

	suite.True(strings.Contains(problem.Detail, "missing required parameter"))
	//Clean up
	domain.DeleteSnapshot(expectedSnapshot.Name)
}

func Test_API_Snapshot(t *testing.T) {
	//This setups the production routs and handler
	logrus.SetLevel(logrus.TraceLevel)

	CreateRouterAndHandler()
	ConfigureSystemForUnitTesting()
	suite.Run(t, new(Snapshot_TS))
}
