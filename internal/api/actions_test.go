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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Cray-HPE/hms-firmware-action/internal/domain"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/presentation"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type Update_TS struct {
	suite.Suite
}

// SetupSuit is run ONCE
func (suite *Update_TS) SetupSuite() {
}

func (suite *Update_TS) Test_POST_Action_BadXname() {

	actionParams := storage.ActionParameters{
		StateComponentFilter: storage.StateComponentFilter{
			Xnames: []string{"badXname"},
		},
	}

	apj, _ := json.Marshal(actionParams)
	aps := string(apj)

	r, _ := http.NewRequest("POST", "/actions", strings.NewReader(aps))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)

	suite.True(strings.Contains(problem.Detail, "invalid/duplicate xnames"))
}

func (suite *Update_TS) Test_POST_All_EmptyBody() {

	r, _ := http.NewRequest("POST", "/actions", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)

	suite.True(strings.Contains(problem.Detail, "empty body not allowed"))
}

func (suite *Update_TS) Test_POST_All_HappyPath() {

	actionParams := GetDefaultActionParameters()
	paramJson, _ := json.Marshal(actionParams)
	paramString := string(paramJson)
	//This method should always return 200, even if there isnt anything to return, b/c empty set is a valid set.
	r, _ := http.NewRequest("POST", "/actions", strings.NewReader(paramString))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusAccepted, resp.StatusCode)

	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	usp := presentation.CreateActionPayload{}
	_ = json.Unmarshal(body, &usp)

	logrus.Trace(usp)
}

func (suite *Update_TS) Test_POST_Actions_DuplicateXnames() {
	actionParams := storage.ActionParameters{
		StateComponentFilter: storage.StateComponentFilter{
			Xnames: []string{"x0c0s1b0", "x0c0s1b0"},
		},
	}

	apj, _ := json.Marshal(actionParams)
	aps := string(apj)

	r, _ := http.NewRequest("POST", "/actions", strings.NewReader(aps))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)

	suite.True(strings.Contains(problem.Detail, "invalid/duplicate xnames"))
}

func (suite *Update_TS) Test_POST_Actions_GoodXnamesWithCompositeTargets_HappyPath() {
	actionParams := GetDefaultCompositeActionParameters()
	paramJson, _ := json.Marshal(actionParams)
	paramString := string(paramJson)
	//This method should always return 200, even if there isnt anything to return, b/c empty set is a valid set.
	r, _ := http.NewRequest("POST", "/actions", strings.NewReader(paramString))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	defer DrainAndCloseResponseBody(resp)
	suite.Equal(http.StatusAccepted, resp.StatusCode)

	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	logrus.Debug(body)
	usp := model.IDResp{}
	_ = json.Unmarshal(body, &usp)

	logrus.Trace(usp)
}

func (suite *Update_TS) Test_DELETE_Action_NoID() {
	r, _ := http.NewRequest("DELETE", "/actions/", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	suite.Equal(http.StatusNotFound, resp.StatusCode)
}

func (suite *Update_TS) Test_DELETE_Action_BADID() {
	r, _ := http.NewRequest("DELETE", "/actions/badUUID", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (suite *Update_TS) Test_DELETE_Action_Good() {
	action := storage.HelperGetStockAction()
	action.State.SetState("completed")
	err := DSP.StoreAction(action)
	suite.True(err == nil)

	r, _ := http.NewRequest("DELETE", "/actions/"+action.ActionID.String(), nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	suite.Equal(http.StatusNoContent, resp.StatusCode)
}

func (suite *Update_TS) Test_KILL_Action_Completed() {
	action := storage.HelperGetStockAction()
	action.State.SetState("completed")
	err := DSP.StoreAction(action)
	suite.True(err == nil)

	r, _ := http.NewRequest("DELETE", "/actions/"+action.ActionID.String()+"/instance", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func (suite *Update_TS) Test_KILL_Action_NotComplete() {
	action := storage.HelperGetStockAction()
	err := DSP.StoreAction(action)
	suite.True(err == nil)

	r, _ := http.NewRequest("DELETE", "/actions/"+action.ActionID.String()+"/instance", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	suite.Equal(http.StatusAccepted, resp.StatusCode)
}

func (suite *Update_TS) Test_KILL_Action_NotFound() {
	action := storage.HelperGetStockAction()
	err := DSP.StoreAction(action)
	suite.True(err == nil)

	r, _ := http.NewRequest("DELETE", "/actions/"+uuid.New().String()+"/instance", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	suite.Equal(http.StatusNotFound, resp.StatusCode)
}

func (suite *Update_TS) Test_GET_actions_ALL_After_UpdateAll() {
	parameters := storage.ActionParameters{
		Command: storage.Command{
			OverrideDryrun: false,
		},
	}

	pb := domain.TriggerFirmwareUpdate(parameters)
	logrus.Trace(pb)

	logrus.Trace(pb.Obj.(presentation.CreateActionPayload), pb.StatusCode)
	ActionID := pb.Obj.(presentation.CreateActionPayload).ActionID
	//TODO FIX THIS::: we might be getting the ActionID too soon
	time.Sleep(5 * time.Second)
	var expectedStatusAll presentation.ActionMarshaled

	rAll, _ := http.NewRequest("GET", "/actions/"+(ActionID.String()), nil)
	wAll := httptest.NewRecorder()
	NewRouter().ServeHTTP(wAll, rAll)
	respAll := wAll.Result()

	logrus.Trace(respAll)
	//if it exists should get a 200
	suite.Equal(http.StatusOK, respAll.StatusCode)
	//read the body, unmarshall and turn into an application
	bodyAll, _ := ioutil.ReadAll(respAll.Body)
	returnedStatusAll := presentation.ActionMarshaled{}
	_ = json.Unmarshal(bodyAll, &returnedStatusAll)

	//TODO need to verify the location header

	pbAll := domain.GetAction(ActionID)
	expectedStatusAll = pbAll.Obj.(presentation.ActionMarshaled)
	logrus.Trace(expectedStatusAll, returnedStatusAll)
	suite.True(expectedStatusAll.Equals(returnedStatusAll))
	suite.True(true)
}

func (suite *Update_TS) Test_GET_actions_BadRequest() {
	path := "/actions/BadRequest"
	r, _ := http.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	suite.Equal(resp.StatusCode, http.StatusBadRequest)
}

func (suite *Update_TS) Test_GET_actions_NotFound() {
	// 0 is invalid
	path := "/actions/" + uuid.New().String()
	r, _ := http.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	suite.Equal(http.StatusNotFound, resp.StatusCode)
}

func (suite *Update_TS) Test_GET_actions_All() {
	var expectedStatusAll presentation.ActionSummaries

	rAll, _ := http.NewRequest("GET", "/actions", nil)
	wAll := httptest.NewRecorder()
	NewRouter().ServeHTTP(wAll, rAll)
	respAll := wAll.Result()
	//if it exists should get a 200
	suite.Equal(respAll.StatusCode, http.StatusOK)
	//read the body, unmarshall and turn into an application
	bodyAll, _ := ioutil.ReadAll(respAll.Body)
	returnedStatusAll := presentation.ActionSummaries{}
	_ = json.Unmarshal(bodyAll, &returnedStatusAll)

	pbAll := domain.GetAllActions()
	expectedStatusAll = pbAll.Obj.(presentation.ActionSummaries)
	logrus.Trace(expectedStatusAll)
	logrus.Trace(returnedStatusAll)
	suite.True(expectedStatusAll.Equals(returnedStatusAll))
}

func Test_API_Actions(t *testing.T) {
	//This setups the production routs and handler
	logrus.SetLevel(logrus.TraceLevel)
	CreateRouterAndHandler()
	ConfigureSystemForUnitTesting()
	suite.Run(t, new(Update_TS))
}
