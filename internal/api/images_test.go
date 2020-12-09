// MIT License
//
// (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package api

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/domain"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/model"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/presentation"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/storage"
)

type Images_TS struct {
	suite.Suite
}

func Helper_GetDefaultRawImage() (rImage presentation.RawImage) {
	rImage = presentation.RawImage{
		DeviceType:              "routerbmc",
		Manufacturer:            "cray",
		Models:                  []string{"Columbia"},
		Target:                  "BMC",
		Tags:                    []string{"default", "persist"},
		FirmwareVersion:         "sc-1.2.3-linux",
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
		Tags:                    []string{"default", "persist"},
		FirmwareVersion:         "nc-1.2.3-linux",
		SemanticFirmwareVersion: "1.2.3",
		S3URL:                   "s3://firmware/sc-1.2.3-linux.bin",
	}
	return
}

// SetupSuit is run ONCE
func (suite *Images_TS) SetupSuite() {
}

/****  Images may not be empty, can not test -- Keeping in for reference
// TEST: Test_GET_images_empty_HappyPath
// GET /images
// Returns 200 with null
func (suite *Images_TS) Test_GET_images_empty_HappyPath() {
	r, _ := http.NewRequest("GET", "/images", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	suite.Equal("null", string(body))
}
****/

// TEST: Test_GET_images_Arr_HappyPath
// GET /images
// Returns 200 with stored image
func (suite *Images_TS) Test_GET_images_Arr_HappyPath() {
	rImage := Helper_GetDefaultRawImage()
	pb := domain.CreateImage(rImage)
	imageID := pb.Obj.(storage.ImageID)

	r, _ := http.NewRequest("GET", "/images", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	imageArr := presentation.Images{}
	_ = json.Unmarshal(body, &imageArr)
	suite.True(len(imageArr.Images) >= 1)

	image, _ := rImage.NewImage()
	image.ImageID = imageID.ImageID
	for _, d := range imageArr.Images {
		if d.ImageID == imageID.ImageID {
			// Time will be different, but we do not care, so clear

			imgMar := presentation.ToImageMarshaled(image)
			imgMar.CreateTime = ""
			d.CreateTime = ""
			suite.True(d.Equals(imgMar))
		}
	}

	domain.DeleteImage(imageID.ImageID)
}

// TEST: Test_GET_images_HappyPath
// GET /images/{imageID}
// Returns 200 with stored image
func (suite *Images_TS) Test_GET_images_HappyPath() {
	rImage := Helper_GetDefaultRawImage()

	pb := domain.CreateImage(rImage)
	imageID := pb.Obj.(storage.ImageID)

	r, _ := http.NewRequest("GET", "/images/"+imageID.ImageID.String(), nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	image := storage.Image{}
	_ = json.Unmarshal(body, &image)
	suite.True(image.ImageID == imageID.ImageID)

	domain.DeleteImage(imageID.ImageID)
}

// TEST: Test_GET_images_NotFound
// GET /images/{uuid.NEW}
// Returns 404 with error
func (suite *Images_TS) Test_GET_images_NotFound() {
	r, _ := http.NewRequest("GET", "/images/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusNotFound, resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)
}

// TEST: Test_GET_images_BadID
// GET /images/bad_id
// Returns 400 with error
func (suite *Images_TS) Test_GET_images_BadID() {
	r, _ := http.NewRequest("GET", "/images/bad_id", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)
	suite.True(strings.Contains(problem.Detail, "invalid UUID length"))
}

// TEST: Test_DELETE_images_NotFound
// DELETE /images/{uuid.NEW}
// Returns 404 with error
func (suite *Images_TS) Test_DELETE_images_NotFound() {
	r, _ := http.NewRequest("DELETE", "/images/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusNotFound, resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)
}

// TEST: Test_DELETE_images_BadID
// DELETE /images/bad_id
// Returns 400 with error
func (suite *Images_TS) Test_DELETE_images_BadID() {
	r, _ := http.NewRequest("DELETE", "/images/bad_id", nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)
	suite.True(strings.Contains(problem.Detail, "invalid UUID length"))
}

// TEST: Test_DELETE_images_HappyPath
// DELETE /images/{ImageID}
// Returns 200
func (suite *Images_TS) Test_DELETE_images_HappyPath() {
	rImage := Helper_GetDefaultRawImage()

	pb := domain.CreateImage(rImage)
	imageID := pb.Obj.(storage.ImageID)
	r, _ := http.NewRequest("DELETE", "/images/"+imageID.ImageID.String(), nil)
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	suite.Equal(http.StatusNoContent, resp.StatusCode)
}

// TEST: Test_POST_images_HappyPATH
// POST /images
// Returns 204
func (suite *Images_TS) Test_POST_images_HappyPath() {
	rImage := Helper_GetDefaultRawImage()

	apj, _ := json.Marshal(rImage)
	aps := string(apj)

	r, _ := http.NewRequest("POST", "/images", strings.NewReader(aps))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	imageID := storage.ImageID{}
	_ = json.Unmarshal(body, &imageID)
	suite.Equal(http.StatusOK, resp.StatusCode)

	domain.DeleteImage(imageID.ImageID)
}

// TEST: Test_POST_images_HappyPATH
// POST /images
// Returns 200
func (suite *Images_TS) Test_POST_images_Dup_HappyPath() {
	rImage := Helper_GetDefaultRawImage()

	apj, _ := json.Marshal(rImage)
	aps := string(apj)

	r, _ := http.NewRequest("POST", "/images", strings.NewReader(aps))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	imageID := storage.ImageID{}
	_ = json.Unmarshal(body, &imageID)
	suite.Equal(http.StatusOK, resp.StatusCode)

	r, _ = http.NewRequest("POST", "/images", strings.NewReader(aps))
	w = httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	imageID2 := storage.ImageID{}
	_ = json.Unmarshal(body, &imageID2)
	suite.Equal(http.StatusOK, resp.StatusCode)

	domain.DeleteImage(imageID.ImageID)
	domain.DeleteImage(imageID2.ImageID)
}

// TEST: Test_POST_images_DependHappyPath
// POST /images
// Returns 200
func (suite *Images_TS) Test_POST_images_DepHappyPath() {
	rImage := Helper_GetDefaultRawImageWNC()
	apj, _ := json.Marshal(rImage)
	aps := string(apj)

	r, _ := http.NewRequest("POST", "/images", strings.NewReader(aps))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	imageID := storage.ImageID{}
	_ = json.Unmarshal(body, &imageID)
	suite.Equal(http.StatusOK, resp.StatusCode)

	rImage.Target = "BIOS"
	rImage.FirmwareVersion = "1.5.7"
	rImage.SemanticFirmwareVersion = "1.5.7"
	rImage.S3URL = "s3://firmware/1.5.7"
	apj, _ = json.Marshal(rImage)
	aps = string(apj)

	r, _ = http.NewRequest("POST", "/images", strings.NewReader(aps))
	w = httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)
	imageID2 := storage.ImageID{}
	_ = json.Unmarshal(body, &imageID2)
	suite.Equal(http.StatusOK, resp.StatusCode)

	domain.DeleteImage(imageID.ImageID)
	domain.DeleteImage(imageID2.ImageID)
}

// TEST: Test_PUT_images_NotFound
// GET /images/{uuid.NEW}
// Returns 201
func (suite *Images_TS) Test_PUT_images_NotFound() {
	rImage := Helper_GetDefaultRawImageWNC()

	apj, _ := json.Marshal(rImage)
	aps := string(apj)
	imageID := uuid.New()
	r, _ := http.NewRequest("PUT", "/images/"+imageID.String(), strings.NewReader(aps))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusCreated, resp.StatusCode)
	domain.DeleteImage(imageID)
}

// TEST: Test_PUT_images_BadID
// GET /images/bad_id
// Returns 400 with error
func (suite *Images_TS) Test_PUT_images_BadID() {
	rImage := Helper_GetDefaultRawImageWNC()

	apj, _ := json.Marshal(rImage)
	aps := string(apj)
	r, _ := http.NewRequest("PUT", "/images/bad_id", strings.NewReader(aps))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusBadRequest, resp.StatusCode)

	//read the body, unmarshall and turn into an application
	body, _ := ioutil.ReadAll(resp.Body)
	problem := model.Problem7807{}
	_ = json.Unmarshal(body, &problem)
	suite.True(strings.Contains(problem.Detail, "invalid UUID length"))
}

// TEST: Test_PUT_Images_Happy
// PUT /images with valid payload
// Returns 200 with imageID
func (suite *Images_TS) Test_PUT_images_Happy() {
	rImage := Helper_GetDefaultRawImageWNC()

	pb := domain.CreateImage(rImage)
	if pb.IsError == true {
		logrus.Error(pb.Error)
	}
	imageID := pb.Obj.(storage.ImageID)

	rImage.Target = "Recovery"
	apj, _ := json.Marshal(rImage)
	aps := string(apj)

	r, _ := http.NewRequest("PUT", "/images/"+imageID.ImageID.String(), strings.NewReader(aps))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()

	suite.Equal(http.StatusOK, resp.StatusCode)
	domain.DeleteImage(imageID.ImageID)
}

// TEST: Test_POST_Images_Errors
// POST /images with invalid payload
func (suite *Images_TS) Test_POST_images_Errors() {
	rImage := Helper_GetDefaultRawImageWNC()
	rImage.SemanticFirmwareVersion = "badversion"

	apj, _ := json.Marshal(rImage)
	aps := string(apj)

	r, _ := http.NewRequest("POST", "/images", strings.NewReader(aps))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	suite.Equal(http.StatusBadRequest, resp.StatusCode)

}

// TEST: Test_PUT_Images_Errors
// POST /images with invalid payload
func (suite *Images_TS) Test_PUT_images_Errors() {
	rImage := Helper_GetDefaultRawImageWNC()
	rImage.SemanticFirmwareVersion = "1.2.3"
	rImage.S3URL = ""
	apj, _ := json.Marshal(rImage)
	aps := string(apj)

	r, _ := http.NewRequest("PUT", "/images/"+uuid.New().String(), strings.NewReader(aps))
	w := httptest.NewRecorder()
	NewRouter().ServeHTTP(w, r)
	resp := w.Result()
	suite.Equal(http.StatusBadRequest, resp.StatusCode)
}

func Test_API_Images(t *testing.T) {
	//This setups the production routs and handler
	CreateRouterAndHandler()
	ConfigureSystemForUnitTesting()
	suite.Run(t, new(Images_TS))
}
