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

	"github.com/sirupsen/logrus"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/model"
)

func ServiceStatusDetails(check CheckServiceStatus) (pb model.Passback) {
	var fusStatus ServiceStatus
	pb.StatusCode = http.StatusOK //optimistic initialization

	if check.Status {
		if *GLOB.Running {
			fusStatus.Status = "running"
		} else {
			fusStatus.Status = "not running"
			pb.StatusCode = http.StatusServiceUnavailable
			logrus.Error("ERROR: FAS Running = FALSE")
		}
	}

	if check.Version {
		if dat, err := ioutil.ReadFile(".version"); err == nil {
			fusStatus.Version = strings.TrimSpace(string(dat))
		} else if dat, err := ioutil.ReadFile("../../.version"); err == nil {
			fusStatus.Version = strings.TrimSpace(string(dat))
		} else {
			pb.StatusCode = http.StatusInternalServerError
			err = errors.New("could not find version file")
			logrus.Error(err)
			fusStatus.Version = err.Error()
		}
	}

	if check.HSMStatus {
		if err := (*GLOB.HSM).Ping(); err == nil {
			fusStatus.HSMStatus = "connected"
		} else {
			fusStatus.HSMStatus = "not connected"
			pb.StatusCode = http.StatusInternalServerError
			logrus.Error("ERROR: HSM Status = NOT CONNECTED")
		}
	}

	if check.StorageStatus {
		if err := (*GLOB.DSP).Ping(); err == nil {
			fusStatus.StorageStatus = "connected"
		} else {
			fusStatus.StorageStatus = "not connected"
			pb.StatusCode = http.StatusInternalServerError
			logrus.Error("ERROR: DATABASE Status = NOT CONNECTED")
		}
	}

	if check.RFTransportStatus {
		if *(*GLOB).RFTransportReady == true {
			fusStatus.RFTransportStatus = "ready"
		} else {
			fusStatus.RFTransportStatus = "not ready"
			pb.StatusCode = http.StatusServiceUnavailable
			logrus.Error("ERROR: RF TRANSPORT Status = NOT READY")
		}
	}

	pb = model.BuildSuccessPassback(pb.StatusCode, fusStatus)

	return pb
}

type ServiceStatus struct {
	Version           string `json:"serviceVersion,omitempty"`
	Status            string `json:"serviceStatus,omitempty"`
	HSMStatus         string `json:"hsmStatus,omitempty"`
	StorageStatus     string `json:"storageStatus,omitempty"`
	RFTransportStatus string `json:"rfTransportStatus,omitempty"`
}

type CheckServiceStatus struct {
	Version           bool
	Status            bool
	HSMStatus         bool
	StorageStatus     bool
	RFTransportStatus bool
}
