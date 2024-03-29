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
	"bufio"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/presentation"
)

var LoaderRunning bool = false
var LOADERLOGSDIR string = "/fw/loaderlogs/"

func GetLoaderStatus() (pb model.Passback) {
	var lStatus presentation.LoaderStatus
	pb.StatusCode = http.StatusOK
	if LoaderRunning {
		lStatus.Status = "busy"
	} else {
		lStatus.Status = "ready"
	}
	files, err := ioutil.ReadDir(LOADERLOGSDIR)
	if err == nil {
		for _, file := range files {
			filename := file.Name()
			var llist presentation.LoaderList
			llist.LoaderRunID = filename
			lStatus.LoaderRunList = append(lStatus.LoaderRunList, llist)
		}
	}
	pb = model.BuildSuccessPassback(pb.StatusCode, lStatus)
	return pb
}

func GetLoaderStatusID(id uuid.UUID) (pb model.Passback) {
	var lOutput presentation.LoaderOutput
	filename := LOADERLOGSDIR + id.String()
	file, err := os.Open(filename)
	if err != nil {
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
		return pb
	}
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		lOutput.Output = append(lOutput.Output, scan.Text())
	}
	pb = model.BuildSuccessPassback(http.StatusOK, lOutput)
	return pb
}

func DoLoader(localFile string, id uuid.UUID) {
	var cmd *exec.Cmd
	LoaderRunning = true // incase it was not set before
	os.Mkdir(LOADERLOGSDIR, 0777)
	logfilename := LOADERLOGSDIR + id.String()
	logfile, err := os.Create(logfilename)
	if err != nil {
		logrus.Error(err)
		LoaderRunning = false
		return
	}
	defer logfile.Close()

	if localFile == "" {
		cmd = exec.Command("/fw-loader", "--log-level", "info")
	} else {
		cmd = exec.Command("/fw-loader", "--log-level", "info", "--local-file", localFile)
	}

	cmd.Stdout = logfile
	cmd.Stderr = logfile

	err = cmd.Run()
	if err != nil {
		logrus.Error("LOADER ERROR: ", err)
	}
	logfile.Close()
	output, err := ioutil.ReadFile(logfilename)
	if err == nil {
		logrus.Info(string(output))
	}

	LoaderRunning = false
}

func DeleteLoaderRun(id uuid.UUID) (pb model.Passback) {
	logfilename := LOADERLOGSDIR + id.String()
	logrus.Debug("Delete file: ", logfilename)
	_, err := os.Stat(logfilename)
	if os.IsNotExist(err) {
		pb = model.BuildErrorPassback(http.StatusNotFound, err)
		return pb
	}
	err = os.Remove(logfilename)
	if err != nil {
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		return pb
	}
	pb = model.BuildSuccessPassback(http.StatusNoContent, nil)
	return pb
}

// Loads firmware from Nexus - Called when FAS Starts
// Continues runnning until images are in FAS
func DoLoadFromNexus(sleeptimeMinutes int) {
	var id uuid.UUID
	sleeptime := time.Duration(sleeptimeMinutes) * time.Minute
	imageCount, _ := NumImages()
	for imageCount == 0 {
		if LoaderRunning {
			sleeptime = 30 * time.Second
		} else {
			oldid := id
			id = uuid.New()
			logrus.Info("Auto Load from Nexus")
			DoLoader("", id)
			DeleteLoaderRun(oldid)
			sleeptime = time.Duration(sleeptimeMinutes) * time.Minute
		}
		time.Sleep(sleeptime)
		imageCount, _ = NumImages()
	}
}
