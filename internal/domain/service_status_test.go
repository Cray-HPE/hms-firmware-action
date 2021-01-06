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
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ServiceStatus_TS struct {
	suite.Suite
}

func HelperVersionNum() (ver string, err error) {
	if dat, err := ioutil.ReadFile(".version"); err == nil {
		ver = strings.TrimSpace(string(dat))
	} else if dat, err := ioutil.ReadFile("../../.version"); err == nil {
		ver = strings.TrimSpace(string(dat))
	} else {
		err = errors.New("could not find version file")
	}
	return
}

// SetupSuit is run ONCE
func (suite *ServiceStatus_TS) SetupSuite() {
}

func (suite *ServiceStatus_TS) Test_GET_service_status_HappyPath() {
	var expectedServiceStatus ServiceStatus
	expectedServiceStatus.Status = "running"

	check := CheckServiceStatus{
		Status: true,
	}

	pb := ServiceStatusDetails(check)
	returnedServiceStatus := pb.Obj.(ServiceStatus)
	suite.Equal(pb.StatusCode, http.StatusOK)
	suite.Equal(expectedServiceStatus, returnedServiceStatus)
}

func (suite *ServiceStatus_TS) Test_GET_service_version_HappyPath() {
	var expectedServiceStatus ServiceStatus
	var err error

	expectedServiceStatus.Version, err = HelperVersionNum()
	suite.True(err == nil)

	check := CheckServiceStatus{
		Version: true,
	}

	pb := ServiceStatusDetails(check)
	returnedServiceStatus := pb.Obj.(ServiceStatus)
	suite.Equal(pb.StatusCode, http.StatusOK)
	suite.Equal(expectedServiceStatus, returnedServiceStatus)
}

func (suite *ServiceStatus_TS) Test_GET_service_status_ALL_HappyPath() {
	var expectedServiceStatus ServiceStatus
	var err error
	expectedServiceStatus.Status = "running"
	expectedServiceStatus.StorageStatus = "connected"
	expectedServiceStatus.Version, err = HelperVersionNum()
	suite.True(err == nil)

	check := CheckServiceStatus{
		Version:       true,
		Status:        true,
		HSMStatus:     true,
		StorageStatus: true,
	}

	pb := ServiceStatusDetails(check)
	returnedServiceStatus := pb.Obj.(ServiceStatus)
	suite.Equal(expectedServiceStatus.Status, returnedServiceStatus.Status)
	suite.Equal(expectedServiceStatus.Version, returnedServiceStatus.Version)
	suite.Equal(expectedServiceStatus.StorageStatus, returnedServiceStatus.StorageStatus)

	if returnedServiceStatus.HSMStatus == "error" {
		suite.Equal(pb.StatusCode, http.StatusInternalServerError)
		expectedServiceStatus.HSMStatus = "error"
		suite.Equal(expectedServiceStatus.HSMStatus, returnedServiceStatus.HSMStatus)
	} else {
		suite.Equal(pb.StatusCode, http.StatusOK)
		expectedServiceStatus.HSMStatus = "connected"
		suite.True(strings.Contains(returnedServiceStatus.HSMStatus, expectedServiceStatus.HSMStatus))
	}
}

func Test_DOMAIN_ServiceStatus(t *testing.T) {
	ConfigureSystemForUnitTesting()
	suite.Run(t, new(ServiceStatus_TS))
}
