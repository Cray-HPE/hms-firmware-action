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

package test

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"github.com/Cray-HPE/hms-certs/pkg/hms_certs"
	"github.com/Cray-HPE/hms-firmware-action/internal/logger"
	trsapi "github.com/Cray-HPE/hms-trs-app-api/pkg/trs_http_api"
	"time"
)


type MockGlobals struct {
	Logger             *logrus.Logger
	BaseTRSTask        trsapi.HttpTask
	Running            *bool
	TLOC_rf            trsapi.TrsAPI
	TLOC_svc           trsapi.TrsAPI
	RFHttpClient       *hms_certs.HTTPClientPair
	SVCHttpClient      *hms_certs.HTTPClientPair
	StateManagerServer string
	VaultEnabled       bool
	VaultKeypath       string
	LockEnabled        bool
}

func (glob *MockGlobals) NewGlobals() () {

	//Logging
	glob.Logger = logger.Init()
	//Logging
	glob.Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	glob.Logger.SetReportCaller(true)

	//HSM
	glob.VaultEnabled = true // since we are now using real HMS, need to use real vault
	glob.VaultKeypath = "hms-creds"    //need to use the correct keypath

	envstr := os.Getenv("SMS_SERVER")
	if envstr != "" {
		glob.StateManagerServer = "http://cray-smd:27779"
	} else {
		glob.StateManagerServer = "http://localhost:27779"
	}

	//TRS
	glob.BaseTRSTask.ServiceName = "cray-hms-firmware-action"
	glob.BaseTRSTask.Timeout = time.Second * 30
	glob.BaseTRSTask.Request, _ = http.NewRequest("GET", "", nil)
	glob.BaseTRSTask.Request.Header.Set("Content-Type", "application/json")

	//var TLOC trsapi.TrsAPI

	worker := &trsapi.TRSHTTPLocal{}
	worker.Logger = glob.Logger
	glob.TLOC_rf = worker
	glob.TLOC_rf.Init("cray-hms-firmware-action", glob.Logger)
	glob.TLOC_svc = worker
	glob.TLOC_svc.Init("cray-hms-firmware-action", glob.Logger)

	//HTTP clients

	hc,_ := hms_certs.CreateHTTPClientPair("",10)
	glob.RFHttpClient = hc
	glob.SVCHttpClient = hc

	tmpBool := true
	glob.Running = &tmpBool

	glob.LockEnabled = true

}
