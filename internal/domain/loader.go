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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"

	"stash.us.cray.com/HMS/hms-firmware-action/internal/model"
)

var LoaderRunning bool = false
var LOADEROUTFILE string = "/loader_out.txt"

type LoaderStatus struct {
	Status string `json:"loaderStatus,omitempty"`
	Output string `json:"lastRunOutput,omitempty"`
}

func GetLoaderStatus() (pb model.Passback) {
	var lStatus LoaderStatus
	pb.StatusCode = http.StatusOK
	if LoaderRunning {
		lStatus.Status = "busy"
	} else {
		lStatus.Status = "ready"
	}
	output, err := ioutil.ReadFile(LOADEROUTFILE)
	if err == nil {
		lStatus.Output = string(output)
	}
	pb = model.BuildSuccessPassback(pb.StatusCode, lStatus)
	return pb
}

func DoLoader(localFile string) {
	var cmd *exec.Cmd
	LoaderRunning = true // incase it was not set before
	outfile, err := os.Create(LOADEROUTFILE)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer outfile.Close()

	if localFile == "" {
		cmd = exec.Command("/fw-loader")
	} else {
		cmd = exec.Command("/fw-loader", "--localFile", localFile)
	}

	cmd.Stdout = outfile
	cmd.Stderr = outfile

	err = cmd.Run()
	if err != nil {
		fmt.Println("LOADER ERROR: ", err)
	}
	outfile.Close()
	output, err := ioutil.ReadFile(LOADEROUTFILE)
	if err == nil {
		logrus.Info(string(output))
	}

	LoaderRunning = false
}
