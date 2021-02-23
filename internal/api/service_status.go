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
	"fmt"
	"net/http"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/domain"
)

func ServiceStatus(w http.ResponseWriter, req *http.Request) {
	var check domain.CheckServiceStatus
	check.Status = true
	pb := domain.ServiceStatusDetails(check)
	WriteHeaders(w, pb)
}

func ServiceStatusVersion(w http.ResponseWriter, req *http.Request) {
	var check domain.CheckServiceStatus
	check.Version = true
	pb := domain.ServiceStatusDetails(check)
	WriteHeaders(w, pb)
}

func ServiceStatusDetails(w http.ResponseWriter, req *http.Request) {
	var check domain.CheckServiceStatus
	check.Version = true
	check.HSMStatus = true
	check.Status = true
	check.StorageStatus = true
	pb := domain.ServiceStatusDetails(check)
	WriteHeaders(w, pb)
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hms-firmware-action")
	w.WriteHeader(http.StatusOK)
}
