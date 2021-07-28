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

package hsm

import (
	"testing"

	"github.com/Cray-HPE/hms-smd/pkg/redfish"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/Cray-HPE/hms-firmware-action/internal/logger"
	"github.com/Cray-HPE/hms-firmware-action/internal/test"
)

type Models_TS struct {
	suite.Suite
}

//func ToHsmDataFromRFEndpoint(src *rf.RedfishEPDescription)(dst HsmData){
func (suite *Models_TS) Test_ToHsmDataFromRFEndpoint(){
	rfep := rf.RedfishEPDescription {
		ID:           "xnameID",
		Type:         "newtype",
		Hostname:     "hostname",
		Domain:       "domain",
		FQDN:         "1.2.3.4",
		Password:     "password",
		User:         "root",
	}
  hsm := ToHsmDataFromRFEndpoint(&rfep)
  suite.Equal(hsm.ID,rfep.ID)
  suite.Equal(hsm.Type,rfep.Type)
  suite.Equal(hsm.Domain,rfep.Domain)
  suite.Equal(hsm.FQDN,rfep.FQDN)
  suite.Equal(hsm.Password,rfep.Password)
  suite.Equal(hsm.User,rfep.User)
}

func (suite *Models_TS) Test_HSMData_CopyFrom() {
	hmsd1 := HsmData {
		ID:           "xnameID",
		Type:         "newtype",
		Hostname:     "hostname",
		Domain:       "domain",
		FQDN:         "1.2.3.4",
		Password:     "password",
		User:         "root",
	}
  hmsd2 := HsmData {}
  hmsd2.CopyFrom(&hmsd1)
  suite.Equal(hmsd1.ID,hmsd2.ID)
  suite.Equal(hmsd1.Type,hmsd2.Type)
  suite.Equal(hmsd1.Domain,hmsd2.Domain)
  suite.Equal(hmsd1.FQDN,hmsd2.FQDN)
  suite.Equal(hmsd1.Password,hmsd2.Password)
  suite.Equal(hmsd1.User,hmsd2.User)
}

func (suite *Models_TS) Test_HSMData_Equals() {
	hmsd1 := HsmData {
		ID:           "xnameID",
		Type:         "newtype",
		Hostname:     "hostname",
		Domain:       "domain",
		FQDN:         "1.2.3.4",
		Password:     "password",
		User:         "root",
  }
	hmsd2 := HsmData {
		ID:           "xnameID2",
		Type:         "newtype",
		Hostname:     "hostname",
		Domain:       "domain",
		FQDN:         "1.2.3.4",
		Password:     "password",
		User:         "root",
  }
  suite.True(hmsd1.Equals(hmsd1))
  suite.True(hmsd2.Equals(hmsd2))
  suite.False(hmsd1.Equals(hmsd2))
  suite.False(hmsd2.Equals(hmsd1))
}

func (suite *Models_TS) Test_DiscoveryInfo_Equals() {
  di1 := rf.DiscoveryInfo {
		LastStatus: "OK",
		RedfishVersion: "1.0",
  }
  di2 := rf.DiscoveryInfo {
		LastStatus: "notOK",
		RedfishVersion: "1.0",
  }
  suite.True(DiscoveryInfoEquals(di1,di1))
  suite.True(DiscoveryInfoEquals(di2,di2))
  suite.False(DiscoveryInfoEquals(di2,di1))
  suite.False(DiscoveryInfoEquals(di1,di2))
}

func Test_Models(t *testing.T) {
	//ConfigureSystemForUnitTesting()

	var mockGlobals = test.MockGlobals{}
	mockGlobals.NewGlobals()

	logy := logger.Init()
	logy.SetLevel(logrus.TraceLevel)

	suite.Run(t, new(Models_TS))
}
