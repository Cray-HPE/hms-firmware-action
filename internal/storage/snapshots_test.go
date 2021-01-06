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

package storage

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type Snapshot_TS struct {
	suite.Suite
}

//SNAPSHOT
func (suite *Snapshot_TS) Test_Snapshot_Equals_HappyPath_NonEmptyDevices() {
	s1 := HelperGetFilledSnapshot(3)
	s2 := HelperGetFilledSnapshot(3)

	s1.Name = s2.Name
	logrus.Trace(s1, s2)
	suite.True(s1.Equals(s2))
	suite.True(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))

	s2.Devices[1] = HelperGetFakeSnapshotDevices("BADXNAME", 0)
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
}

func (suite *Snapshot_TS) Test_Snapshot_Equals_HappyPath_EmptyDevices() {
	s1 := HelperGetStockSnapshotFixedDate()
	s2 := HelperGetStockSnapshotFixedDate()

	logrus.Trace(s1, s2)
	s1.Name = s2.Name
	suite.True(s1.Equals(s2))
	suite.True(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))
}

func (suite *Snapshot_TS) Test_Snapshot_Equals_NotEqual_MismatchNonSlice() {
	s1 := HelperGetStockSnapshot()
	// Time needs to be different
	time.Sleep(1 * time.Second)
	s2 := HelperGetStockSnapshot()

	logrus.Trace(s1, s2)
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))
}

func (suite *Snapshot_TS) Test_Snapshot_Equals_NotEqual_MismatchDeviceCount() {
	s1 := HelperGetFilledSnapshot(3)
	s2 := HelperGetFilledSnapshot(2)

	logrus.Trace(s1, s2)
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))
}

func (suite *Snapshot_TS) Test_Snapshot_Equals_NotEqual_MismatchDeviceCount_OneEmpty() {
	s1 := HelperGetFilledSnapshot(3)
	s2 := HelperGetFilledSnapshot(0)

	logrus.Trace(s1, s2)
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))
	// Change Order
	s2.Devices = append(s2.Devices, s1.Devices[1])
	s2.Devices = append(s2.Devices, s1.Devices[0])
	s2.Devices = append(s2.Devices, s1.Devices[2])
	suite.True(s1.Equals(s1))
	// Change Targets
	s2.Devices[1] = HelperGetFakeSnapshotDevices(s1.Devices[0].Xname, 3)
	logrus.Trace(s1, s2)
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
}

//SNAPSHOTDEVICES
func (suite *Snapshot_TS) Test_SnapshotDevices_Equals_NotEqual_Xname_DevicesSame() {
	s1 := HelperGetFakeSnapshotDevices("x0c0s2b0", 3)
	s2 := HelperGetFakeSnapshotDevices("x0c0s1b0", 3)

	logrus.Trace(s1, s2)
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))
}

func (suite *Snapshot_TS) Test_SnapshotDevices_Equals_NotEqual_Xname_DevicesDiff() {
	s1 := HelperGetFakeSnapshotDevices("x0c0s2b0", 3)
	s2 := HelperGetFakeSnapshotDevices("x0c0s2b0", 0)

	logrus.Trace(s1, s2)
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))
}

func (suite *Snapshot_TS) Test_SnapshotDevices_Equals_NotEqual_XnameEqual_DevicesDiff() {
	s1 := HelperGetFakeSnapshotDevices("x0c0s2b0", 3)
	s2 := HelperGetFakeSnapshotDevices("x0c0s2b0", 0)

	logrus.Trace(s1, s2)
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))

	// Change Order
	s2.Targets = append(s2.Targets, s1.Targets[1])
	s2.Targets = append(s2.Targets, s1.Targets[2])
	s2.Targets = append(s2.Targets, s1.Targets[0])
	suite.True(s1.Equals(s2))
	// Test Targets not equal
	s2.Targets[1].FirmwareVersion = "none"
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
	// Change Target
	s2.Targets[1] = HelperGetFakeSnapshotTargets("BAD")
	logrus.Trace(s1, s2)
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
}

func (suite *Snapshot_TS) Test_SnapshotDevices_Equals_HappyPath_EmptyTargets() {
	s1 := HelperGetFakeSnapshotDevices("x0c0s2b0", 0)
	s2 := HelperGetFakeSnapshotDevices("x0c0s2b0", 0)

	logrus.Trace(s1, s2)
	suite.True(s1.Equals(s2))
	suite.True(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))
}

func (suite *Snapshot_TS) Test_SnapshotDevices_Equals_HappyPath_NonEmptyTargets() {
	s1 := HelperGetFakeSnapshotDevices("x0c0s2b0", 4)
	s2 := HelperGetFakeSnapshotDevices("x0c0s2b0", 4)

	logrus.Trace(s1, s2)
	suite.True(s1.Equals(s2))
	suite.True(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))
}

//SNAPSHOTTARGETS

func (suite *Snapshot_TS) Test_SnapshotTargets_Equals_HappyPath() {
	s1 := HelperGetFakeSnapshotTargets("bar")
	s2 := HelperGetFakeSnapshotTargets("bar")

	logrus.Trace(s1, s2)
	suite.True(s1.Equals(s2))
	suite.True(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))
}

func (suite *Snapshot_TS) Test_SnapshotTargets_Equals_NotEqual_ID() {
	s1 := HelperGetFakeSnapshotTargets("bar")
	s1.Name = "FOO"
	s2 := HelperGetFakeSnapshotTargets("bar")

	logrus.Trace(s1, s2)
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))
}

func (suite *Snapshot_TS) Test_SnapshotTargets_Equals_NotEqual_Version() {
	s1 := HelperGetFakeSnapshotTargets("bar")
	s1.FirmwareVersion = "FOO"
	s2 := HelperGetFakeSnapshotTargets("bar")

	logrus.Trace(s1, s2)
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))
}

func (suite *Snapshot_TS) Test_SnapshotTargets_Equals_NotEqual_Both() {
	s1 := HelperGetFakeSnapshotTargets("bar")
	s1.Name = "FOO"
	s1.FirmwareVersion = "BAZ"
	s2 := HelperGetFakeSnapshotTargets("bar")

	logrus.Trace(s1, s2)
	suite.False(s1.Equals(s2))
	suite.False(s2.Equals(s1))
	suite.True(s2.Equals(s2))
	suite.True(s1.Equals(s1))
}

func (suite *Snapshot_TS) Test_SnapshotParameters_Equals() {
	sp1 := SnapshotParameters{
		Name: "n1",
	}
	sp1.TargetFilter = TargetFilter{[]string{"BMC", "BIOS"}}
	time.Sleep(1 * time.Second)
	sp2 := SnapshotParameters{
		Name: "n2",
	}
	sp2.TargetFilter = TargetFilter{[]string{"BMC", "BIOS", "RECOVER"}}
	suite.True(sp1.Equals(sp1))
	suite.True(sp2.Equals(sp2))
	suite.False(sp1.Equals(sp2))
	suite.False(sp2.Equals(sp1))
}

//TODO I think logic like this belongs in presentation pkg.... It cant live here,b/c of import cycle
////SNAPSHOTCOUNTS
//func (suite *Snapshot_TS) Test_SnapshotCounts_Equals_HappyPath() {
//	ss1 := HelperGetFakeSnapshots(4)
//	ss2 := HelperGetFakeSnapshots(4)
//	s1 := ss1
//	s2 := ss2
//
//	logrus.Trace(s1, s2)
//	suite.True(s1.Equals(s2))
//	suite.True(s2.Equals(s1))
//	suite.True(s2.Equals(s2))
//	suite.True(s1.Equals(s1))
//}
//
//func (suite *Snapshot_TS) Test_SnapshotCounts_Equals_NotEqual() {
//	ss1 := HelperGetFakeSnapshots(4)
//	ss2 := HelperGetFakeSnapshots(5)
//	s1 := ss1.ToSnapshotsCounts()
//	s2 := ss2.ToSnapshotsCounts()
//
//	logrus.Trace(s1, s2)
//	suite.False(s1.Equals(s2))
//	suite.False(s2.Equals(s1))
//	suite.True(s2.Equals(s2))
//	suite.True(s1.Equals(s1))
//}
//
//func (suite *Snapshot_TS) Test_ToSnapshotCounts_HappyPath() {
//	ss1 := HelperGetFakeSnapshots(4)
//	ss2 := HelperGetFakeSnapshots(4)
//	s1 := ss1.ToSnapshotsCounts()
//	s2 := ss2.ToSnapshotsCounts()
//
//	logrus.Trace(s1, s2)
//	suite.True(s1.Equals(s2))
//	suite.True(s2.Equals(s1))
//	suite.True(s2.Equals(s2))
//	suite.True(s1.Equals(s1))
//}
//
//
////SNAPSHOTCOUNT
//
//func (suite *Snapshot_TS) Test_SnapshotCount_Equals_HappyPath() {
//	ss1 := HelperGetFilledSnapshot(4)
//	ss2 := HelperGetFilledSnapshot(4)
//	s1 := ss1.ToSnapshotsCount()
//	s2 := ss2.ToSnapshotsCount()
//
//	logrus.Trace(s1, s2)
//	suite.True(s1.Equals(s2))
//	suite.True(s2.Equals(s1))
//	suite.True(s2.Equals(s2))
//	suite.True(s1.Equals(s1))
//}
//
//func (suite *Snapshot_TS) Test_SnapshotCount_Equals_NotEqual() {
//	ss1 := HelperGetFilledSnapshot(4)
//	ss2 := HelperGetFilledSnapshot(5)
//	s1 := ss1.ToSnapshotsCount()
//	s1.InProgress = false
//	s2 := ss2.ToSnapshotsCount()
//
//	logrus.Trace(s1, s2)
//	suite.False(s1.Equals(s2))
//	suite.False(s2.Equals(s1))
//	suite.True(s2.Equals(s2))
//	suite.True(s1.Equals(s1))
//}
//
//func (suite *Snapshot_TS) Test_ToSnapshotCount_HappyPath() {
//	ss1 := HelperGetFilledSnapshot(4)
//	ss2 := HelperGetStockSnapshotFixedDate()
//	s2 := presentation.SnapshotSummary{
//		Name:              ss2.Name,
//		DateTime:          ss2.CaptureTime,
//		Ready:             ss2.Ready,
//		UniqueDeviceCount: 4,
//	}
//	s1 := ss1.ToSnapshotsCount()
//
//	logrus.Trace(s1, s2)
//	suite.True(s1.Equals(s2))
//	suite.True(s2.Equals(s1))
//	suite.True(s2.Equals(s2))
//	suite.True(s1.Equals(s1))
//}

func Test_Snapshot_Suite(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	suite.Run(t, new(Snapshot_TS))
}
