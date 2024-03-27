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

package presentation

import (
	"time"

	"github.com/Masterminds/semver"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
)

type Images struct {
	Images []ImageMarshaled `json:"images"'`
}

type RawImage struct {
	DeviceType                        string   `json:"deviceType"`
	Manufacturer                      string   `json:"manufacturer,omitempty"`
	Models                            []string `json:"models,omitempty"`
	SoftwareIds                       []string `json:"softwareIds,omitempty"`
	Target                            string   `json:"target,omitempty"`
	Tags                              []string `json:"tags,omitempty"`
	FirmwareVersion                   string   `json:"firmwareVersion"`
	SemanticFirmwareVersion           string   `json:"semanticFirmwareVersion,omitempty"`
	UpdateURI                         string   `json:"updateURI"`
	NeedManualReboot                  bool     `json:"needManualReboot,omitempty"`
	WaitTimeBeforeManualRebootSeconds int      `json:"waitTimeBeforeManualRebootSeconds"`
	WaitTimeAfterRebootSeconds        int      `json:"waitTimeAfterRebootSeconds"`
	PollingSpeedSeconds               int      `json:"pollingSpeedSeconds"`
	ForceResetType                    string   `json:"forceResetType"`
	S3URL                             string   `json:"s3URL"`
	TftpURL                           string   `json:"tftpURL"`
	AllowableDeviceStates             []string `json:"allowableDeviceStates,omitempty"`
}

func (obj *RawImage) Equals(other RawImage) bool {
	if obj.DeviceType != other.DeviceType ||
		obj.Manufacturer != other.Manufacturer ||
		model.StringSliceEquals(obj.Models, other.Models) ||
		model.StringSliceEquals(obj.SoftwareIds, other.SoftwareIds) ||
		obj.Target != other.Target ||
		model.StringSliceEquals(obj.Tags, other.Tags) ||
		obj.FirmwareVersion != other.FirmwareVersion ||
		obj.SemanticFirmwareVersion != other.SemanticFirmwareVersion ||
		obj.UpdateURI != other.UpdateURI ||
		obj.PollingSpeedSeconds != other.PollingSpeedSeconds ||
		obj.ForceResetType != other.ForceResetType ||
		obj.NeedManualReboot != other.NeedManualReboot ||
		obj.WaitTimeBeforeManualRebootSeconds != other.WaitTimeBeforeManualRebootSeconds ||
		obj.WaitTimeAfterRebootSeconds != other.WaitTimeAfterRebootSeconds ||
		obj.S3URL != other.S3URL ||
		obj.TftpURL != other.TftpURL ||
		model.StringSliceEquals(obj.AllowableDeviceStates, other.AllowableDeviceStates) == false {
		return false
	}
	return true
}

func (other *RawImage) NewImage() (obj storage.Image, err error) {
	obj.ImageID = uuid.New()
	obj.CreateTime.Scan(time.Now())
	obj.DeviceType = other.DeviceType
	obj.Manufacturer = other.Manufacturer
	obj.Models = append(obj.Models, other.Models...)
	obj.SoftwareIds = append(obj.SoftwareIds, other.SoftwareIds...)
	obj.Target = other.Target
	obj.FirmwareVersion = other.FirmwareVersion
	obj.Tags = append(obj.Tags, other.Tags...)
	v, err := semver.NewVersion(other.SemanticFirmwareVersion)
	if err != nil {
		logrus.Error(err)
		return obj, err
	} else {
		obj.SemanticFirmwareVersion = v
	}

	obj.UpdateURI = other.UpdateURI
	obj.NeedManualReboot = other.NeedManualReboot
	obj.WaitTimeAfterRebootSeconds = other.WaitTimeAfterRebootSeconds
	obj.WaitTimeBeforeManualRebootSeconds = other.WaitTimeBeforeManualRebootSeconds
	obj.ForceResetType = other.ForceResetType
	obj.PollingSpeedSeconds = other.PollingSpeedSeconds
	obj.S3URL = other.S3URL
	obj.TftpURL = other.TftpURL
	obj.AllowableDeviceStates = append(obj.AllowableDeviceStates, other.AllowableDeviceStates...)

	return obj, nil
}

type ImageMarshaled struct {
	ImageID                           uuid.UUID `json:"imageID"`
	CreateTime                        string    `json:"createTime,omitempty"`
	DeviceType                        string    `json:"deviceType,omitempty"`
	Manufacturer                      string    `json:"manufacturer,omitempty"`
	Models                            []string  `json:"models,omitempty"`
	SoftwareIds                       []string  `json:"softwareIds,omitempty"`
	Target                            string    `json:"target,omitempty"`
	Tags                              []string  `json:"tags,omitempty"`
	FirmwareVersion                   string    `json:"firmwareVersion,omitempty"`
	SemanticFirmwareVersion           string    `json:"semanticFirmwareVersion,omitempty"`
	UpdateURI                         string    `json:"updateURI,omitempty"`
	NeedManualReboot                  bool      `json:"needManualReboot,omitempty"`
	WaitTimeBeforeManualRebootSeconds int       `json:"waitTimeBeforeManualRebootSeconds,omitempty"`
	WaitTimeAfterRebootSeconds        int       `json:"waitTimeAfterRebootSeconds,omitempty"`
	PollingSpeedSeconds               int       `json:"pollingSpeedSeconds,omitempty"`
	ForceResetType                    string    `json:"forceResetType,omitempty"`
	S3URL                             string    `json:"s3URL,omitempty"`
	TftpURL                           string    `json:"tftpURL,omitempty"`
	AllowableDeviceStates             []string  `json:"allowableDeviceStates,omitempty"`
}

func (obj ImageMarshaled) Equals(other ImageMarshaled) bool {

	if obj.ImageID != other.ImageID {
		logrus.Warn("imageID is not equal")
		return false
	} else if obj.CreateTime != other.CreateTime {

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
	} else if obj.SemanticFirmwareVersion != other.SemanticFirmwareVersion {
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

func ToImageMarshaled(from storage.Image) (to ImageMarshaled) {
	to = ImageMarshaled{
		ImageID:                           from.ImageID,
		CreateTime:                        from.CreateTime.Time.Format(time.RFC3339),
		DeviceType:                        from.DeviceType,
		Manufacturer:                      from.Manufacturer,
		Models:                            from.Models,
		SoftwareIds:                       from.SoftwareIds,
		Target:                            from.Target,
		Tags:                              from.Tags,
		FirmwareVersion:                   from.FirmwareVersion,
		SemanticFirmwareVersion:           from.SemanticFirmwareVersion.String(),
		UpdateURI:                         from.UpdateURI,
		NeedManualReboot:                  from.NeedManualReboot,
		WaitTimeBeforeManualRebootSeconds: from.WaitTimeBeforeManualRebootSeconds,
		WaitTimeAfterRebootSeconds:        from.WaitTimeAfterRebootSeconds,
		PollingSpeedSeconds:               from.PollingSpeedSeconds,
		ForceResetType:                    from.ForceResetType,
		S3URL:                             from.S3URL,
		TftpURL:                           from.TftpURL,
		AllowableDeviceStates:             from.AllowableDeviceStates,
	}

	return to
}
