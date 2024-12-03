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

package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/Cray-HPE/hms-firmware-action/internal/domain"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/presentation"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CreateImage - will create an image record
func CreateImage(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	var pb model.Passback
	var image presentation.RawImage

	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		logrus.WithFields(logrus.Fields{"body": string(body)}).Trace("Printing request body -- CreateImage")

		if err != nil {
			pb := model.BuildErrorPassback(http.StatusInternalServerError, err)
			logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("Error detected retrieving body")
			WriteHeaders(w, pb)
			return
		}

		err = json.Unmarshal(body, &image)
		if err != nil {
			pb = model.BuildErrorPassback(http.StatusBadRequest, err)
			logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("Unparseable json")
			WriteHeaders(w, pb)
			return
		}

		logrus.Debug("image:", image)
		pb = domain.CreateImage(image)
		if pb.IsError == false {
			imageID := pb.Obj.(storage.ImageID)
			location := "../images/" + imageID.ImageID.String()
			WriteHeadersWithLocation(w, pb, location)
		} else {
			WriteHeaders(w, pb)
		}
		return
	}
	err := errors.New("body cannot be empty")
	pb = model.BuildErrorPassback(http.StatusBadRequest, err)
	logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("empty body")
	return
}

// GetImages - will return all images
func GetImages(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	pb := domain.GetImages()
	WriteHeaders(w, pb)
}

// GetImage - will return an image
func GetImage(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	pb := GetUUIDFromVars("imageID", req)
	if pb.IsError {
		WriteHeaders(w, pb)
		return
	}
	imageID := pb.Obj.(uuid.UUID)
	pb = domain.GetImage(imageID)
	WriteHeaders(w, pb)
	return
}

// DeleteImage - will delete an image record
func DeleteImage(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	pb := GetUUIDFromVars("imageID", req)
	if pb.IsError {
		WriteHeaders(w, pb)
		return
	}
	imageID := pb.Obj.(uuid.UUID)
	pb = domain.DeleteImage(imageID)
	WriteHeaders(w, pb)
	return
}

// UpdateImage - will create or update an image record
func UpdateImage(w http.ResponseWriter, req *http.Request) {

	defer DrainAndCloseRequestBody(req)

	var image presentation.RawImage

	pb := GetUUIDFromVars("imageID", req)
	if pb.IsError {
		WriteHeaders(w, pb)
		return
	}
	imageID := pb.Obj.(uuid.UUID)

	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		logrus.WithFields(logrus.Fields{"body": string(body)}).Trace("Printing request body")

		if err != nil {
			pb := model.BuildErrorPassback(http.StatusInternalServerError, err)
			logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("Error detected retrieving body")
			WriteHeaders(w, pb)
			return
		}

		err = json.Unmarshal(body, &image)
		if err != nil {
			pb = model.BuildErrorPassback(http.StatusBadRequest, err)
			logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("Unparseable json")
			WriteHeaders(w, pb)
			return
		}

		logrus.Debug("image:", image)
		pb = domain.UpdateImage(image, imageID)
		WriteHeaders(w, pb)
		return
	} else {
		err := errors.New("body cannot be empty")
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("empty body")
		WriteHeaders(w, pb)
		return
	}
}
