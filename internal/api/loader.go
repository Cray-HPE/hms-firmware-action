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
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/sirupsen/logrus"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/domain"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/model"
)

type LoaderID struct {
	LoaderRunID uuid.UUID `json:"loaderRunID,omitempty"`
}

func LoaderStatusID(w http.ResponseWriter, req *http.Request) {
	logrus.Debug("In LoaderStatusID")
	pb := GetUUIDFromVars("loaderID", req)
	if pb.IsError {
		WriteHeaders(w, pb)
		return
	}
	loaderRunID := pb.Obj.(uuid.UUID)
	pb = domain.GetLoaderStatusID(loaderRunID)
	WriteHeaders(w, pb)
	return
}

func LoaderStatus(w http.ResponseWriter, req *http.Request) {
	logrus.Debug("In LoaderStatus")
	pb := domain.GetLoaderStatus()
	WriteHeaders(w, pb)
	return
}

func LoaderLoadNexus(w http.ResponseWriter, req *http.Request) {
	logrus.Debug("In LoaderLoadNexus")
	var err error
	var pb model.Passback
	var loaderID LoaderID
	if domain.LoaderRunning {
		err = errors.New("Loader busy, try again later")
		pb = model.BuildErrorPassback(http.StatusTooManyRequests, err)
		logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("Loader busy")
		WriteHeaders(w, pb)
		return
	}
	loaderID.LoaderRunID = uuid.New()
	pb = model.BuildSuccessPassback(http.StatusOK, loaderID)
	WriteHeaders(w, pb)
	go domain.DoLoader("", loaderID.LoaderRunID)
	return
}

func LoaderLoad(w http.ResponseWriter, req *http.Request) {
	logrus.Debug("In LoaderLoad")
	var err error
	var pb model.Passback
	var loaderID LoaderID
	if domain.LoaderRunning {
		err = errors.New("Loader busy, try again later")
		pb = model.BuildErrorPassback(http.StatusTooManyRequests, err)
		logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("Loader busy")
		WriteHeaders(w, pb)
		return
	}
	domain.LoaderRunning = true
	req.ParseMultipartForm(32 << 20)
	uploadfile, header, err := req.FormFile("file")
	if err != nil {
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("Error no file found")
		WriteHeaders(w, pb)
		domain.LoaderRunning = false
		return
	}
	defer uploadfile.Close()
	filename := header.Filename
	ext := filepath.Ext(filename)
	if (strings.ToLower(ext) != ".rpm") && (strings.ToLower(ext) != ".zip") {
		err = errors.New(filename + " not .rpm or .zip --> " + ext)
		pb = model.BuildErrorPassback(http.StatusUnsupportedMediaType, err)
		logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error(filename + " must be .rpm " + ext)
		WriteHeaders(w, pb)
		domain.LoaderRunning = false
		return
	}
	filename = "/" + filename
	savefile, err := os.Create(filename)
	if err != nil {
		logrus.Error("can not open file "+filename+" ", err)
		domain.LoaderRunning = false
		err = errors.New("ERROR Creating File")
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error()
		WriteHeaders(w, pb)
		domain.LoaderRunning = false
		return
	} else {
		defer savefile.Close()
		io.Copy(savefile, uploadfile)
	}
	loaderID.LoaderRunID = uuid.New()
	pb = model.BuildSuccessPassback(http.StatusOK, loaderID)
	WriteHeaders(w, pb)
	req.MultipartForm.RemoveAll()
	go domain.DoLoader(filename, loaderID.LoaderRunID)
	return
}

func LoaderDeleteID(w http.ResponseWriter, req *http.Request) {
	logrus.Debug("In LoaderDeleteID")
	pb := GetUUIDFromVars("loaderID", req)
	if pb.IsError {
		WriteHeaders(w, pb)
		return
	}
	loaderRunID := pb.Obj.(uuid.UUID)
	pb = domain.DeleteLoaderRun(loaderRunID)
	WriteHeaders(w, pb)
}
