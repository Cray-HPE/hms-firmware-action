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

package hsm

import (
	"stash.us.cray.com/HMS/hms-base"
)

type HSMv0 struct {
	HSMGlobals HSM_GLOBALS
}

type HSMProvider interface {
	RefillModelRF(XtHd *map[XnameTarget]HsmData, specialTargets map[string]string) (errs []error)
	GetTargetsRF(hd *map[string]HsmData) (tuples []XnameTarget, errs []error)
	FillRedfishEndpointData(hd *map[string]HsmData) (errs []error)
	FillUpdateServiceData(hd *map[string]HsmData) (errs []error)
	FillModelManufacturerRF(hd *map[string]HsmData) (errs []error)
	FillComponentEndpointData(hd *map[string]HsmData) (errs []error)
	GetStateComponents(xnames []string, partitions []string, groups []string, types []string) (data base.ComponentArray, err error)
	FillHSMData(xnames []string, partitions []string, groups []string, types []string) (hd map[string]HsmData, errs []error)
	RestoreCredentials(hd *HsmData) (err error)
	//OtherStuff -> GOOD
	ClearLock(xnames []string) error
	SetLock(xnames []string) error
	Ping() (err error)
	Init(globals *HSM_GLOBALS) (err error)
}
