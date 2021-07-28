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
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/Cray-HPE/hms-firmware-action/internal/presentation"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
)

func Helper_GetDefaultRawImage() (rImage presentation.RawImage) {
	rImage = presentation.RawImage{
		DeviceType:              "routerbmc",
		Manufacturer:            "cray",
		Models:                  []string{"Columbia"},
		Target:                  "BMC",
		FirmwareVersion:         "sc-1.2.3-linux",
		Tags:                    []string{"default"},
		SemanticFirmwareVersion: "1.2.3",
		S3URL:                   "s3://firmware/sc-1.2.3-linux.bin",
	}
	return
}

func Helper_GetDefaultRawImageWNC() (rImage presentation.RawImage) {
	rImage = presentation.RawImage{
		DeviceType:              "nodebmc",
		Manufacturer:            "cray",
		Models:                  []string{"WindomNodeBoard"},
		Target:                  "BMC",
		FirmwareVersion:         "nc-1.2.3-linux",
		SemanticFirmwareVersion: "1.2.3",
		S3URL:                   "s3://firmware/sc-1.2.3-linux.bin",
	}
	return
}

type Images_TS struct {
	suite.Suite
}

// SetupSuit is run ONCE
func (suite *Images_TS) SetupSuite() {
}

// TEST: Test_CreateImage_GoodImage
func (suite *Images_TS) Test_CreateImage_GoodImage() {
	rawImage := Helper_GetDefaultRawImage()
	pb := CreateImage(rawImage)
	suite.False(pb.IsError)
	imageID := pb.Obj.(storage.ImageID)

	// Make sure image is in array
	pb = GetImages()
	suite.False(pb.IsError)

	marshaledImages := pb.Obj.(presentation.Images)
	suite.True(len(marshaledImages.Images) >= 1)
	image, _ := rawImage.NewImage()
	image.ImageID = imageID.ImageID
	count := 0
	for _, imageMarshaled := range marshaledImages.Images {
		if imageMarshaled.ImageID == imageID.ImageID {
			imgMar := presentation.ToImageMarshaled(image)
			// Time will be different, but we do not care, so clear
			imgMar.CreateTime = ""
			imageMarshaled.CreateTime = ""
			suite.True(imageMarshaled.Equals(imgMar))
			count++
		}
	}
	suite.True(count == 1)
	// Retrieve Image and Check
	pb = GetImage(imageID.ImageID)
	suite.False(pb.IsError)
	retImage := pb.Obj.(presentation.ImageMarshaled)

	imgMar := presentation.ToImageMarshaled(image)
	// Time could be different, but we do not care, so clear
	imgMar.CreateTime = ""
	retImage.CreateTime = ""
	suite.True(retImage.Equals(imgMar))
	logrus.Info(imgMar)
	logrus.Info(retImage)

	// Delete Image
	pb = DeleteImage(imageID.ImageID)
	suite.False(pb.IsError)
}

func (suite *Images_TS) Test_GetImage_Good() {
	rImage := Helper_GetDefaultRawImage()
	pb := CreateImage(rImage)
	suite.False(pb.IsError)
	imageID := pb.Obj.(storage.ImageID)

	// Make sure image is in array
	pb = GetImage(imageID.ImageID)
	suite.False(pb.IsError)
	//iRet := pb.Obj.(storage.Image)

	// Delete Image
	pb = DeleteImage(imageID.ImageID)
	suite.False(pb.IsError)
}

func (suite *Images_TS) Test_GetImage_BadImage() {
	pb := GetImage(uuid.New())
	suite.True(pb.IsError)
	suite.Equal(pb.StatusCode, http.StatusNotFound)
}

func (suite *Images_TS) Test_CreateImage_BadImage() {
	rImage := Helper_GetDefaultRawImage()
	rImage.Models = nil
	pb := CreateImage(rImage)
	suite.True(pb.IsError)
	suite.Equal(http.StatusBadRequest, pb.StatusCode)
}

func (suite *Images_TS) Test_GetImages() {
	rImage := Helper_GetDefaultRawImage()

	pb := CreateImage(rImage)
	suite.False(pb.IsError)
	imageID := pb.Obj.(storage.ImageID)

	rImage2 := Helper_GetDefaultRawImage()

	pb = CreateImage(rImage)
	suite.False(pb.IsError)
	imageID2 := pb.Obj.(storage.ImageID)

	// Make sure image is in array
	pb = GetImages()
	suite.False(pb.IsError)

	imageArr := pb.Obj.(presentation.Images)
	suite.True(len(imageArr.Images) >= 2)
	image, _ := rImage.NewImage()
	image2, _ := rImage2.NewImage()
	image.ImageID = imageID.ImageID
	image2.ImageID = imageID2.ImageID
	count1 := 0
	count2 := 0
	for _, d := range imageArr.Images {
		if d.ImageID == imageID.ImageID {
			// Time will be different, but we do not care, so clear
			imgMar := presentation.ToImageMarshaled(image)
			imgMar.CreateTime = ""
			d.CreateTime = ""
			suite.True(d.Equals(imgMar))
			count1++
		}
		if d.ImageID == imageID2.ImageID {

			imgMar2 := presentation.ToImageMarshaled(image2)
			// Time will be different, but we do not care, so clear
			imgMar2.CreateTime = ""
			d.CreateTime = ""
			suite.True(d.Equals(imgMar2))
			count2++
		}
	}
	suite.True(count1 == 1)
	suite.True(count2 == 1)
	// Retrieve Image and Check
	pb = GetImage(imageID.ImageID)
	suite.False(pb.IsError)
	retImage := pb.Obj.(presentation.ImageMarshaled)

	imgMar := presentation.ToImageMarshaled(image)
	// Time will be different, but we do not care, so clear
	imgMar.CreateTime = ""
	retImage.CreateTime = ""
	suite.True(retImage.Equals(imgMar))

	// Delete Image
	pb = DeleteImage(imageID.ImageID)
	suite.False(pb.IsError)
}

func (suite *Images_TS) Test_DeleteImage_Valid() {
	rImage := Helper_GetDefaultRawImage()

	pb := CreateImage(rImage)
	suite.False(pb.IsError)
	imageID := pb.Obj.(storage.ImageID)

	// Get Image Array Size
	pb = GetImages()
	suite.False(pb.IsError)

	imageArr := pb.Obj.(presentation.Images)
	count := len(imageArr.Images)

	// Delete Image
	pb = DeleteImage(imageID.ImageID)
	suite.False(pb.IsError)

	// Make sure Image removed from Array
	pb = GetImages()
	suite.False(pb.IsError)

	imageArr = pb.Obj.(presentation.Images)
	suite.True(count == len(imageArr.Images)+1)
}

func (suite *Images_TS) Test_DeleteImage_Invalid() {
	pb := DeleteImage(uuid.New())
	suite.True(pb.IsError)
}

func (suite *Images_TS) Test_UpdateImage_GoodImage() {
	rImage := Helper_GetDefaultRawImage()

	pb := CreateImage(rImage)
	suite.False(pb.IsError)
	imageID := pb.Obj.(storage.ImageID)
	rImage.Models = nil
	rImage.Models = append(rImage.Models, "Colorado")
	pb = UpdateImage(rImage, imageID.ImageID)
	suite.False(pb.IsError)

	pb = DeleteImage(imageID.ImageID)
	suite.False(pb.IsError)
}

func (suite *Images_TS) Test_UpdateImage_BadImage() {
	rImage := Helper_GetDefaultRawImage()

	pb := CreateImage(rImage)
	suite.False(pb.IsError)
	imageID := pb.Obj.(storage.ImageID)
	rImage.Models = nil
	pb = UpdateImage(rImage, imageID.ImageID)
	suite.True(pb.IsError)

	pb = DeleteImage(imageID.ImageID)
	suite.False(pb.IsError)
}

func Test_Domain_Images(t *testing.T) {
	//This setups the production routs and handler
	ConfigureSystemForUnitTesting()
	suite.Run(t, new(Images_TS))
}
