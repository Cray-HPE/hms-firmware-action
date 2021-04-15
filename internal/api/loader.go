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
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/domain"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/model"
)

// CreateAction - creates an action and will trigger an 'update'
func LoaderStatus(w http.ResponseWriter, req *http.Request) {
	pb := domain.GetLoaderStatus()
	WriteHeaders(w, pb)
	return
}

func LoaderLoad(w http.ResponseWriter, req *http.Request) {
	var err error
	var pb model.Passback
	if domain.LoaderRunning {
		err = errors.New("Loader busy, try again later")
		pb = model.BuildErrorPassback(http.StatusTooManyRequests, err)
		logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error("Loader busy")
		WriteHeaders(w, pb)
		return
	}
	domain.LoaderRunning = true
	//	_, err = ioutil.ReadAll(req.Body)
	//if err == io.EOF {
	if req.Body == http.NoBody {
		pb = model.BuildSuccessPassback(http.StatusOK, "")
		WriteHeaders(w, pb)
		go domain.DoLoader("")
		return
	}
	fmt.Println(req.Body)
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
	fmt.Println(filename + " --- " + ext)
	if (ext != ".rpm") && (ext != ".RPM") && (ext != ".zip") && (ext != ".ZIP") {
		err = errors.New(filename + " not .rpm or .zip --> " + ext)
		pb = model.BuildErrorPassback(http.StatusUnsupportedMediaType, err)
		logrus.WithFields(logrus.Fields{"ERROR": err, "HttpStatusCode": pb.StatusCode}).Error(filename + " must be .rpm " + ext)
		WriteHeaders(w, pb)
		domain.LoaderRunning = false
		return
	}
	filename = "/" + filename
	fmt.Println(filename)
	fmt.Println("Filename " + filename)
	savefile, err := os.Create(filename)
	if err != nil {
		log.Fatal("can not open file "+filename+" ", err)
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
	pb = model.BuildSuccessPassback(http.StatusOK, "")
	WriteHeaders(w, pb)
	go domain.DoLoader(filename)
	return
}
