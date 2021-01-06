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
	"github.com/google/uuid"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/presentation"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/storage"
)

type Snapshot_TS struct {
	suite.Suite
}

// SetupSuit is run ONCE
func (suite *Snapshot_TS) SetupSuite() {
}

// GET Snapshot(name)

func (suite *Snapshot_TS) Test_GET_SnapshotByName_NotFound() {
	pb := GetSnapshot("badName")
	suite.True(pb.IsError)
	suite.Equal(http.StatusNotFound, pb.StatusCode)
	logrus.Trace(pb)

}

func (suite *Snapshot_TS) Test_GetSnapshots() {
	ss1 := storage.HelperGetStockSnapshot()
	ss2 := storage.HelperGetStockSnapshot()
	ss2.Name = ss1.Name + ss2.Name

	err := DSP.StoreSnapshot(ss1)
	suite.True(err == nil)

	pb := GetSnapshots()
	suite.False(pb.IsError)
	sshotArr := pb.Obj.(presentation.SnapshotSummaries)
	suite.True(err == nil)
	count1 := 0
	count2 := 0
	for _, d := range sshotArr.Summaries {
		if d.Name == ss1.Name {
			count1++
		}
		if d.Name == ss2.Name {
			count2++
		}
	}
	suite.True(count1 == 1)
	suite.True(count2 == 0)

	err = DSP.StoreSnapshot(ss2)
	suite.True(err == nil)

	pb = GetSnapshots()
	suite.False(pb.IsError)
	sshotArr = pb.Obj.(presentation.SnapshotSummaries)
	suite.True(err == nil)
	count1 = 0
	count2 = 0
	for _, d := range sshotArr.Summaries {
		if d.Name == ss1.Name {
			count1++
		}
		if d.Name == ss2.Name {
			count2++
		}
	}
	suite.True(count1 == 1)
	suite.True(count2 == 1)

	err = DSP.DeleteSnapshot(ss1.Name)
	suite.True(err == nil)
	err = DSP.DeleteSnapshot(ss2.Name)
	suite.True(err == nil)
}

//func GetSnapshot(name string) (pb model.Passback) {
func (suite *Snapshot_TS) Test_GetSnapshot() {
}

func (suite *Snapshot_TS) Test_CreateSnapshots() {
	sp := storage.SnapshotParameters{
		Name: uuid.New().String(),
	}

	pb := CreateSnapshot(sp)
	suite.False(pb.IsError)
	pb = CreateSnapshot(sp)
	suite.True(pb.IsError)
	suite.Equal(http.StatusConflict, pb.StatusCode )
}

//func StartRestoreSnapshot(name string, dryrun bool, timeLimit int) (pb model.Passback) {
func (suite *Snapshot_TS) Test_StartRestoreSnapshot() {
	sp := storage.SnapshotParameters{
		Name: uuid.New().String(),
	}
	pb := CreateSnapshot(sp)
	suite.False(pb.IsError)

	pb = StartRestoreSnapshot(sp.Name, false, 8000)
	suite.False(pb.IsError)
}

//func DeleteSnapshot(name string) (pb model.Passback) {
func (suite *Snapshot_TS) Test_DeleteSnapshot() {
	sp := storage.SnapshotParameters{
		Name: uuid.New().String(),
	}
	pb := CreateSnapshot(sp)
	suite.False(pb.IsError)

	pb = DeleteSnapshot(sp.Name)
	suite.False(pb.IsError)

	pb = DeleteSnapshot("BadName")
	suite.True(pb.IsError)
}

//func RestoreSnapshot(action storage.Action, snapshot storage.Snapshot) {
func (suite *Snapshot_TS) Test_RestoreSnapshot() {
}

//func (suite *Snapshot_TS) Test_SnapshotByName_HappyPath() {
//	//Create new snapshot
//	newSnapshot := CreateSnapshot(strconv.Itoa(rand.Int()))
//	newSnapshotName := newSnapshot.Obj.(presentation.SnapshotName).Name
//
//	pb := GetSnapshot(newSnapshotName)
//	suite.Equal(http.StatusOK, pb.StatusCode)
//	expectedSnapshot := pb.Obj.(presentation.SnapshotMarshaled)
//	suite.Equal(expectedSnapshot.Name, newSnapshotName)
//
//	//Clean up
//	DeleteSnapshot(newSnapshotName)
//}

//func (suite *Snapshot_TS) Test_SnapshotDelete() {
//	//Create new snapshot
//	newSnapshot := CreateSnapshot(strconv.Itoa(rand.Int()))
//	newSnapshotName := newSnapshot.Obj.(presentation.SnapshotName).Name
//
//	pb := GetSnapshot(newSnapshotName)
//	suite.Equal(http.StatusOK, pb.StatusCode)
//	expectedSnapshot := pb.Obj.(presentation.SnapshotMarshaled)
//	suite.Equal(expectedSnapshot.Name, newSnapshotName)
//
//	//Clean up
//	pb = DeleteSnapshot(newSnapshotName)
//	suite.Equal(http.StatusNoContent, pb.StatusCode)
//	pb = GetSnapshot(newSnapshotName)
//	suite.Equal(http.StatusNotFound, pb.StatusCode)
//}

func (suite *Snapshot_TS) Test_SnapshotDelete_NotFound() {
	//Create new snapshot
	pb := DeleteSnapshot("NotFound")
	suite.Equal(http.StatusNotFound, pb.StatusCode)
}

func Test_DOMAIN_Snapshot(t *testing.T) {
	ConfigureSystemForUnitTesting()
	suite.Run(t, new(Snapshot_TS))
}
