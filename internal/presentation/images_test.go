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
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type Images_TS struct {
	suite.Suite
}

// SetupSuit is run ONCE
func (suite *Images_TS) SetupSuite() {
}

func (suite *Images_TS) Test_NewImage() {
	rImage := RawImage{
		DeviceType:              "routerbmc",
		Manufacturer:            "cray",
		Models:                  []string{"Columbia"},
		Target:                  "BMC",
		FirmwareVersion:         "sc-1.2.3-linux",
		SemanticFirmwareVersion: "1.2.3",
		S3URL:                   "s3://firmware/sc-1.2.3-linux.bin",
	}
	image, err := rImage.NewImage()
	suite.True(err == nil)
	suite.True(image.ImageID != uuid.Nil)
	rImage.SemanticFirmwareVersion = "badnum"
	image, err = rImage.NewImage()
	suite.False(err == nil)
	suite.True(err.Error() == "Invalid Semantic Version")
}

func (suite *Images_TS) Test_Equals() {
	rImage1 := RawImage{
		DeviceType:              "routerbmc",
		Manufacturer:            "cray",
		Models:                  []string{"Columbia"},
		Target:                  "BMC",
		FirmwareVersion:         "sc-1.2.3-linux",
		SemanticFirmwareVersion: "1.2.3",
		S3URL:                   "s3://firmware/sc-1.2.3-linux.bin",
	}
	rImage2 := RawImage{
		DeviceType:              "nodebmc",
		Manufacturer:            "cray",
		Models:                  []string{"Columbia"},
		Target:                  "BMC",
		FirmwareVersion:         "sc-1.2.3-linux",
		SemanticFirmwareVersion: "1.2.3",
		S3URL:                   "s3://firmware/sc-1.2.3-linux.bin",
	}
	suite.False(rImage2.Equals(rImage1))
	suite.False(rImage1.Equals(rImage2))
	rImage2.DeviceType = rImage1.DeviceType
	suite.True(rImage2.Equals(rImage1))
	suite.True(rImage1.Equals(rImage2))
	rImage1.AllowableDeviceStates = append(rImage1.AllowableDeviceStates, "On")
	rImage1.AllowableDeviceStates = append(rImage1.AllowableDeviceStates, "Off")
	suite.False(rImage2.Equals(rImage1))
	suite.False(rImage1.Equals(rImage2))
	rImage2.AllowableDeviceStates = append(rImage2.AllowableDeviceStates, "On")
	suite.False(rImage2.Equals(rImage1))
	suite.False(rImage1.Equals(rImage2))
	rImage2.AllowableDeviceStates = append(rImage2.AllowableDeviceStates, "Off")
	suite.True(rImage2.Equals(rImage1))
	suite.True(rImage1.Equals(rImage2))
}

func Test_Presentation_Images(t *testing.T) {
	//This setups the production routs and handler
	suite.Run(t, new(Images_TS))
}
