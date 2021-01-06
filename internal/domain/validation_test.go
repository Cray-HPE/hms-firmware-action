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
	"github.com/Masterminds/semver"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/storage"
)

type Validation_TS struct {
	suite.Suite
}


func Helper_ValidImage()(i storage.Image){
	i = storage.Image{
		ImageID:                           uuid.New(),
		DeviceType:                        "nodeBMC",
		Manufacturer:                      "intel",
		Models:                            []string{"x100","x100_A"},
		Target:                            "BIOS",
		Tags:                              []string{"default","persist"},
		FirmwareVersion:                   "123abc",
		UpdateURI:                         "",
		NeedManualReboot:                  false,
		WaitTimeBeforeManualRebootSeconds: 0,
		WaitTimeAfterRebootSeconds:        0,
		PollingSpeedSeconds:               0,
		ForceResetType:                    "",
		S3URL:                             "s3...",
		AllowableDeviceStates:             nil,
	}

	i.CreateTime.Scan(time.Now())
	i.SemanticFirmwareVersion, _ = semver.NewVersion("1.2.3")

	return i
}

// SetupSuit is run ONCE
func (suite *Validation_TS) SetupSuite() {
}

func (suite *Validation_TS) Test_ValidateActionParameters() {
	// TODO:
}
func (suite *Validation_TS) Test_ValidateSnapshotParameters() {
	// TODO:
}

func (suite *Validation_TS) Test_ValidateXnames() {
	sc := storage.StateComponentFilter{
		Xnames: []string{"x1", "x2"},
	}
	err := ValidateXnames(&sc)
	suite.True(err == nil)
	sc.Xnames = append(sc.Xnames, "x1")
	err = ValidateXnames(&sc)
	suite.False(err == nil)
	suite.True(strings.Contains(err.Error(), "x1"))
	sc.Xnames = append(sc.Xnames, "ac")
	err = ValidateXnames(&sc)
	suite.False(err == nil)
	suite.True(strings.Contains(err.Error(), "x1") && strings.Contains(err.Error(), "ac"))
}

func (suite *Validation_TS) Test_ValidateStateComponentFilter() {
	sc := storage.StateComponentFilter{
		Xnames:     []string{"x1", "x2"},
		Groups:     []string{"g1"},
		Partitions: []string{"p1"},
	}
	err := ValidateStateComponentFilter(&sc)
	suite.True(err == nil)
	sc.Partitions = append(sc.Partitions, "p2")
	err = ValidateStateComponentFilter(&sc)
	suite.True(err == nil)
	sc.Groups = append(sc.Groups, "g2")
	err = ValidateStateComponentFilter(&sc)
	suite.False(err == nil)
}

func (suite *Validation_TS) Test_ValidateImage() {
	sc := storage.StateComponentFilter{
		Xnames:     []string{"x1", "x2"},
		Groups:     []string{"g1"},
		Partitions: []string{"p1"},
	}
	err := ValidateStateComponentFilter(&sc)
	suite.True(err == nil)
	sc.Partitions = append(sc.Partitions, "p2")
	err = ValidateStateComponentFilter(&sc)
	suite.True(err == nil)
	sc.Groups = append(sc.Groups, "g2")
	err = ValidateStateComponentFilter(&sc)
	suite.False(err == nil)
}

func (suite *Validation_TS) Test_ValidateCommandParameter() {
	// TODO:
}
func (suite *Validation_TS) Test_ValidateImageFilter() {
	// TODO:
}

func (suite *Validation_TS) Test_ValidateImage_HappyPath() {
	Image := Helper_ValidImage()
	err := ValidateImageParameters(&Image)
	suite.True(err == nil)
}

func (suite *Validation_TS) Test_ValidateImage_MissingImageID() {
	Image := Helper_ValidImage()
	Image.ImageID = uuid.Nil
	err := ValidateImageParameters(&Image)
	suite.True(err != nil)
}

func (suite *Validation_TS) Test_ValidateImage_MissingDeviceType() {
	Image := Helper_ValidImage()
	Image.DeviceType = ""
	err := ValidateImageParameters(&Image)
	suite.True(err != nil)
}


func (suite *Validation_TS) Test_ValidateImage_MissingManufacturer() {
	Image := Helper_ValidImage()
	Image.Manufacturer = ""
	err := ValidateImageParameters(&Image)
	suite.True(err != nil)
}

func (suite *Validation_TS) Test_ValidateImage_MissingModels() {
	Image := Helper_ValidImage()
	Image.Models = nil
	err := ValidateImageParameters(&Image)
	suite.True(err != nil)
}

func (suite *Validation_TS) Test_ValidateImage_MissingTags() {
	Image := Helper_ValidImage()
	Image.Tags = nil
	err := ValidateImageParameters(&Image)
	suite.True(err != nil)
}

func (suite *Validation_TS) Test_ValidateImage_MissingTarget() {
	Image := Helper_ValidImage()
	Image.Target = ""
	err := ValidateImageParameters(&Image)
	suite.True(err != nil)
}

func (suite *Validation_TS) Test_ValidateImage_MissingFirmwareVersion() {
	Image := Helper_ValidImage()
	Image.FirmwareVersion = ""
	err := ValidateImageParameters(&Image)
	suite.True(err != nil)
}

func (suite *Validation_TS) Test_ValidateImage_Missings3URL() {
	Image := Helper_ValidImage()
	Image.S3URL = ""
	err := ValidateImageParameters(&Image)
	suite.True(err != nil)
}

func Test_Domain_Validation(t *testing.T) {
	//This setups the production routs and handler
	suite.Run(t, new(Validation_TS))
}
