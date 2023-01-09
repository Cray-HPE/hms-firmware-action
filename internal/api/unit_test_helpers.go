/*
 * MIT License
 *
 * (C) Copyright [2020-2023] Hewlett Packard Enterprise Development LP
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
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Cray-HPE/hms-firmware-action/internal/domain"
	"github.com/Cray-HPE/hms-firmware-action/internal/hsm"
	"github.com/Cray-HPE/hms-firmware-action/internal/presentation"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
	"github.com/Cray-HPE/hms-firmware-action/internal/test"
)

var DSP storage.StorageProvider

var IsHandledCSFUT bool

var UTMutex sync.Mutex
var rfClientLockMock *sync.RWMutex = &sync.RWMutex{}

func GetDefaultActionParameters() (params storage.ActionParameters) {
	params.Command = storage.Command{
		OverrideDryrun:             false,
		RestoreNotPossibleOverride: false,
		TimeLimit_Seconds:          0,
		Version:                    "latest",
		Description:                "test",
	}
	return
}

func GetDefaultCompositeActionParameters() (params storage.ActionParameters) {
	params.Command = storage.Command{
		OverrideDryrun:             false,
		RestoreNotPossibleOverride: false,
		TimeLimit_Seconds:          0,
		Version:                    "latest",
		Description:                "test",
	}

	params.TargetFilter = storage.TargetFilter{[]string{"BMC", "BIOS"}}

	params.StateComponentFilter = storage.StateComponentFilter{
		Xnames: []string{"x0c0s1b0", "x0c0s2b0"},
	}
	return
}

func GetDefaultSnapshotParametersMarshaled() (params presentation.SnapshotParametersMarshaled) {
	params.ExpirationTime = (time.Now().AddDate(0, 0, 1)).String()
	params.Name = strconv.Itoa(rand.Int())
	return
}

func GetDefaultSnapshotParameters() (params storage.SnapshotParameters) {
	params.ExpirationTime.Scan((time.Now().AddDate(0, 0, 1)))
	params.Name = strconv.Itoa(rand.Int())
	return
}

func GetDefaultCompositeSnapshotParameters() (params storage.SnapshotParameters) {
	params.ExpirationTime.Scan(time.Now().AddDate(0, 0, 1))

	params.TargetFilter = storage.TargetFilter{[]string{"BMC", "BIOS"}}

	params.StateComponentFilter = storage.StateComponentFilter{
		Xnames: []string{"x0c0s1b0", "x0c0s2b0"},
	}
	return
}

//ToDo should replace storage provider with ETCD
func ConfigureSystemForUnitTesting() {
	UTMutex.Lock()
	if IsHandledCSFUT == false {
		//Storage
		var mockGlobals = test.MockGlobals{}
		mockGlobals.NewGlobals()

		localLogger := logrus.New()
		localLogger.SetLevel(logrus.TraceLevel)
		localLogger.SetReportCaller(true)
		mockGlobals.Logger.SetLevel(logrus.ErrorLevel)

		//Default to ETCD, b/c thats what we are going to use the most!
		envstr := os.Getenv("STORAGE")
		if envstr == "MEMORY" {
			tmpStorageImplementation := &storage.MemStorage{
				Logger: localLogger,
			}
			DSP = tmpStorageImplementation
		} else { //etcd
			tmpStorageImplementation := &storage.ETCDStorage{
				Logger: localLogger,
			}
			DSP = tmpStorageImplementation
		}
		DSP.Init(mockGlobals.Logger)

		//StateManager
		var mockHSM hsm.HSM_GLOBALS
		mockHSM.NewGlobals(mockGlobals.Logger, &mockGlobals.BaseTRSTask, &mockGlobals.TLOC_rf, &mockGlobals.TLOC_svc, mockGlobals.RFHttpClient, mockGlobals.SVCHttpClient, rfClientLockMock, mockGlobals.StateManagerServer, mockGlobals.VaultEnabled, mockGlobals.VaultKeypath, mockGlobals.Running, true)
		var HSM hsm.HSMProvider
		tmpHSMImplementation := &hsm.HSMv0{}
		HSM = tmpHSMImplementation
		HSM.Init(&mockHSM)

		//DOMAIN
		var mockDomain domain.DOMAIN_GLOBALS
		mockDomain.NewGlobals(&mockGlobals.BaseTRSTask, &mockGlobals.TLOC_rf, &mockGlobals.TLOC_svc, mockGlobals.RFHttpClient, mockGlobals.SVCHttpClient, rfClientLockMock, mockGlobals.Running, &DSP, &HSM, 0)
		domain.Init(&mockDomain)
		IsHandledCSFUT = true
	}
	UTMutex.Unlock()
}
