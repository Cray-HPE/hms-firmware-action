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

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/Cray-HPE/hms-firmware-action/internal/domain"
)

type ServiceStatus_TS struct {
	suite.Suite
}

// SetupSuit is run ONCE
func (suite *ServiceStatus_TS) SetupSuite() {
}

func (suite *ServiceStatus_TS) Test_GET_service_status_HappyPath() {
	var expectedServiceStatus domain.ServiceStatus

	r, _ := http.NewRequest("GET", "/service/status", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusOK, resp.StatusCode)

	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	returnedServiceStatus := domain.ServiceStatus{}
	_ = json.Unmarshal(body, &returnedServiceStatus)

	check := domain.CheckServiceStatus{
		Status: true,
	}

	pb := domain.ServiceStatusDetails(check)
	expectedServiceStatus = pb.Obj.(domain.ServiceStatus)
	suite.Equal(expectedServiceStatus, returnedServiceStatus)
}

func (suite *ServiceStatus_TS) Test_GET_service_version_HappyPath() {
	var expectedServiceStatus domain.ServiceStatus

	r, _ := http.NewRequest("GET", "/service/version", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusOK, resp.StatusCode)

	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	returnedServiceStatus := domain.ServiceStatus{}
	_ = json.Unmarshal(body, &returnedServiceStatus)

	check := domain.CheckServiceStatus{
		Version: true,
	}

	pb := domain.ServiceStatusDetails(check)
	expectedServiceStatus = pb.Obj.(domain.ServiceStatus)
	suite.Equal(expectedServiceStatus, returnedServiceStatus)
}

func (suite *ServiceStatus_TS) Test_GET_service_status_details_HappyPath() {
	var expectedServiceStatus domain.ServiceStatus

	r, _ := http.NewRequest("GET", "/service/status/details", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusOK, resp.StatusCode)

	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	returnedServiceStatus := domain.ServiceStatus{}
	_ = json.Unmarshal(body, &returnedServiceStatus)

	check := domain.CheckServiceStatus{
		Version:       true,
		Status:        true,
		HSMStatus:     true,
		StorageStatus: true,
	}

	pb := domain.ServiceStatusDetails(check)
	expectedServiceStatus = pb.Obj.(domain.ServiceStatus)
	suite.Equal(expectedServiceStatus, returnedServiceStatus)
}

func Test_API_ServiceStatus(t *testing.T) {
	//This setups the production routs and handler
	CreateRouterAndHandler()
	ConfigureSystemForUnitTesting()
	suite.Run(t, new(ServiceStatus_TS))
}
