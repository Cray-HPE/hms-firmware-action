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

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"stash.us.cray.com/HMS/hms-firmware-action/internal/model"
)

var LoaderRunning bool = false
var LOADERLOGSDIR string = "/loaderlogs/"

type LoaderList struct {
	LoaderID string `json:"loaderID,omitempty"`
}

type LoaderStatus struct {
	Status        string   `json:"loaderStatus,omitempty"`
	LoaderRunList []string `json:"loaderRunList,omitempty"`
}

type LoaderOutput struct {
	Output []string `json:"loaderRunOutput,omitempty"`
}

func GetLoaderStatus() (pb model.Passback) {
	var lStatus LoaderStatus
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
			lStatus.LoaderRunList = append(lStatus.LoaderRunList, filename)
		}
	}
	pb = model.BuildSuccessPassback(pb.StatusCode, lStatus)
	return pb
}

func GetLoaderStatusID(id uuid.UUID) (pb model.Passback) {
	var lOutput LoaderOutput
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
