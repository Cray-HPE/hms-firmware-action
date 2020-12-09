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

package storage

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/google/uuid"
)

func HelperGetStockImage() (i Image) {
	i = Image{
		ImageID:               uuid.New(),
		DeviceType:            "fake",
		Manufacturer:          "cray",
		Models:                []string{"mock"},
		Target:                "TEST",
		Tags:                  []string{"default"},
		FirmwareVersion:       "1.0.0",
		UpdateURI:             "path/to/URL",
		NeedManualReboot:      false,
		PollingSpeedSeconds:   30,
		S3URL:                 "www.sample.com",
		AllowableDeviceStates: []string{},
	}
	i.CreateTime.Scan(time.Now())
	i.SemanticFirmwareVersion, _ = semver.NewVersion("v1.0.0")
	return i
}

func HelperGetStockAction() (a Action) {
	parameters := ActionParameters{
		Command: Command{
			OverrideDryrun: false,
		},
	}
	a = *NewAction(parameters)
	return
}

func HelperGetStockOperation() (o Operation) {
	o = *NewOperation()

	return
}

func HelperGetStockSnapshotFixedDate() (s Snapshot) {
	s = HelperGetStockSnapshot()
	s.CaptureTime.Scan(time.Date(2020, 01, 14, 14, 50, 20, 333, time.UTC))
	return s
}

func HelperGetStockSnapshot() (s Snapshot) {
	s = Snapshot{
		Name:    uuid.New().String(),
		Ready:   true,
		Devices: nil,
	}
	s.CaptureTime.Scan(time.Now())
	return s
}

func HelperGetFilledSnapshot(numDevice int) (s Snapshot) {
	s = HelperGetStockSnapshotFixedDate()

	for i := 0; i < numDevice; i++ {
		s.Devices = append(s.Devices, HelperGetFakeSnapshotDevices("x2c0r0b"+strconv.Itoa(i), i))
	}

	return s
}

func HelperGetFakeSnapshotDevices(fakeXname string, numDevices int) (sd Device) {

	sd = Device{
		Xname: fakeXname,
	}
	for i := 0; i < numDevices; i++ {
		sd.Targets = append(sd.Targets, HelperGetFakeSnapshotTargets(strconv.Itoa(i)))
	}
	return sd
}

func HelperGetFakeSnapshotTargets(fakeID string) (st Target) {

	st = Target{
		Name:            fakeID,
		FirmwareVersion: "ver." + fakeID + ".foo.123",
	}
	return st
}

/*
func HelperGetFakeSnapshots(numSnapshots int) (ss Snapshots) {
	for i := 0; i <= numSnapshots; i++ {
		s := HelperGetFilledSnapshot(i)
		ss.Snapshots = append(ss.Snapshots, s)
	}
	return ss
}
*/

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func randString(length int) (ret string) {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	ret = b.String() // E.g. "ExcbsVQs"
	return
}

func HelperRandStringSlice(num int) (ret []string) {
	for i := 0; i < num; i++ {
		ret = append(ret, randString(8))
	}
	return
}
