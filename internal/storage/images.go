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

package storage

import (
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"

	"github.com/Masterminds/semver"
	"github.com/google/uuid"
)

type ImageID struct {
	ImageID uuid.UUID `json:"imageID"`
}

//TODO flush this out new rule in documentation!: the firmware version must be unique for the devicetype/manf/model;
//  b.c I cannot figure out WHAT tag they are running just by looking at the firmware on the device!!!
type Image struct {
	ImageID                           uuid.UUID       `json:"imageID"`
	CreateTime                        sql.NullTime    `json:"createTime"`
	DeviceType                        string          `json:"deviceType"`
	Manufacturer                      string          `json:"manufacturer,omitempty"`
	Models                            []string        `json:"models,omitempty"`
	SoftwareIds                       []string        `json:"softwareIds,omitempty"`
	Target                            string          `json:"target,omitempty"`
	Tags                              []string        `json:"tags,omitempty"`
	FirmwareVersion                   string          `json:"firmwareVersion"`
	SemanticFirmwareVersion           *semver.Version `json:"semanticFirmwareVersion,omitempty"`
	UpdateURI                         string          `json:"updateURI"`
	NeedManualReboot                  bool            `json:"needManualReboot"`
	WaitTimeBeforeManualRebootSeconds int             `json:"waitTimeBeforeManualRebootSeconds"`
	WaitTimeAfterRebootSeconds        int             `json:"waitTimeAfterRebootSeconds"`
	PollingSpeedSeconds               int             `json:"pollingSpeedSeconds"`
	ForceResetType                    string          `json:"forceResetType"`
	S3URL                             string          `json:"s3URL"`
	TftpURL                           string          `json:"tftpURL"`
	AllowableDeviceStates             []string        `json:"allowableDeviceStates,omitempty"`
}

func (obj *Image) Equals(other Image) bool {

	if obj.ImageID != other.ImageID {
		logrus.Warn("imageID is not equal")
		return false
	} else if obj.CreateTime.Time.Round(0).Equal(other.CreateTime.Time.Round(0)) == false {
		logrus.Warn("CreateTime is not equal")
		return false
	} else if obj.DeviceType != other.DeviceType {
		logrus.Warn("DeviceType is not equal")
		return false
	} else if obj.Manufacturer != other.Manufacturer {
		logrus.Warn("Manufacturer is not equal")
		return false
	} else if model.StringSliceEquals(obj.Models, other.Models) == false {
		logrus.Warn("Models is not equal")
		return false
	} else if model.StringSliceEquals(obj.SoftwareIds, other.SoftwareIds) == false {
		logrus.Warn("SoftwareIds is not equal")
		return false
	} else if obj.Target != other.Target {
		logrus.Warn("Target is not equal")
		return false
	} else if model.StringSliceEquals(obj.Tags, other.Tags) == false {
		logrus.Warn("Tags is not equal")
		return false
	} else if obj.FirmwareVersion != other.FirmwareVersion {
		logrus.Warn("FirmwareVersion is not equal")
		return false
	} else if obj.SemanticFirmwareVersion.Equal(other.SemanticFirmwareVersion) == false {
		logrus.Warn("SemanticFirmwareVersion is not equal")
		return false
	} else if obj.UpdateURI != other.UpdateURI {
		logrus.Warn("UpdateURI is not equal")
		return false
	} else if obj.NeedManualReboot != other.NeedManualReboot {
		logrus.Warn("NeedManualReboot is not equal")
		return false
	} else if obj.WaitTimeBeforeManualRebootSeconds != other.WaitTimeBeforeManualRebootSeconds {
		logrus.Warn("WaitTimeBeforeManualRebootSeconds is not equal")
		return false
	} else if obj.WaitTimeAfterRebootSeconds != other.WaitTimeAfterRebootSeconds {
		logrus.Warn("WaitTimeAfterRebootSeconds is not equal")
		return false
	} else if obj.PollingSpeedSeconds != other.PollingSpeedSeconds {
		logrus.Warn("PollingSpeedSeconds is not equal")
		return false
	} else if obj.ForceResetType != other.ForceResetType {
		logrus.Warn("ForceResetType is not equal")
		return false
	} else if obj.S3URL != other.S3URL {
		logrus.Warn("S3URL is not equal")
		return false
	} else if obj.TftpURL != other.TftpURL {
		logrus.Warn("TftpURL is not equal")
		return false
	} else if model.StringSliceEquals(obj.AllowableDeviceStates, other.AllowableDeviceStates) == false {
		logrus.Warn("AllowableDeviceStates is not equal")
		return false
	}
	return true
}
