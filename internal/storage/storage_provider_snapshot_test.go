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

import "github.com/sirupsen/logrus"

func (suite *Storage_Provider_TS) Test_Storage_Provider_StoreSnapshot_HappyPath() {
	sshot := HelperGetStockSnapshot()
	err := MS.StoreSnapshot(sshot)
	suite.True(err == nil)

	returnSShot, err := MS.GetSnapshot(sshot.Name)
	suite.True(err == nil)
	suite.True(returnSShot.Equals(sshot))

	err = MS.DeleteSnapshot(sshot.Name)
	suite.True(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_DeleteSnapshot_NotFound() {
	err := MS.DeleteSnapshot("BADNAME")
	suite.False(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_DeleteSnapshot_Happy() {
	snapshot := HelperGetStockSnapshot()
	err := MS.StoreSnapshot(snapshot)
	suite.True(err == nil)

	err = MS.DeleteSnapshot(snapshot.Name)
	suite.True(err == nil)

	// Make sure deleted
	_, err = MS.GetSnapshot(snapshot.Name)
	suite.False(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetSnapshot_NotFound() {
	_, err := MS.GetSnapshot("BADNAME")
	suite.False(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetSnapshot_Happy() {
	logrus.SetReportCaller(true)
	logrus.SetLevel(logrus.TraceLevel)
	sshot := HelperGetStockSnapshot()
	err := MS.StoreSnapshot(sshot)
	suite.True(err == nil)

	returnSShot, err := MS.GetSnapshot(sshot.Name)
	suite.True(err == nil)
	logrus.Trace(sshot)
	logrus.Trace(returnSShot)
	suite.True(returnSShot.Equals(sshot))

	err = MS.DeleteSnapshot(sshot.Name)
	suite.True(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetSnapshots() {
	ss1 := HelperGetStockSnapshot()
	ss2 := HelperGetStockSnapshot()
	ss2.Name = ss1.Name + ss2.Name

	err := MS.StoreSnapshot(ss1)
	suite.True(err == nil)

	sshotArr, err := MS.GetSnapshots()
	suite.True(err == nil)
	count1 := 0
	count2 := 0
	for _, d := range sshotArr {
		if d.Name == ss1.Name {
			suite.True(d.Equals(ss1))
			count1++
		}
		if d.Name == ss2.Name {
			suite.True(d.Equals(ss2))
			count2++
		}
	}
	suite.True(count1 == 1)
	suite.True(count2 == 0)

err = MS.StoreSnapshot(ss2)
	suite.True(err == nil)

	sshotArr, err = MS.GetSnapshots()
	suite.True(err == nil)
	count1 = 0
	count2 = 0
	for _, d := range sshotArr {
		if d.Name == ss1.Name {
			suite.True(d.Equals(ss1))
			count1++
		}
		if d.Name == ss2.Name {
			suite.True(d.Equals(ss2))
			count2++
		}
	}
	suite.True(count1 == 1)
	suite.True(count2 == 1)

	err = MS.DeleteSnapshot(ss1.Name)
	suite.True(err == nil)
	err = MS.DeleteSnapshot(ss2.Name)
	suite.True(err == nil)
}
