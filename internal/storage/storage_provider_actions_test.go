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
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetAction() {
	a := HelperGetStockAction()
	err := MS.StoreAction(a)
	suite.True(err == nil)

	aRet, err := MS.GetAction(a.ActionID)
	suite.True(err == nil)
	suite.True(a.Equals(aRet))

	err = MS.DeleteAction(a.ActionID)
	suite.True(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_DeleteAction_Error() {
	err := MS.DeleteAction(uuid.New())
	suite.True(err != nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_DeleteAction_Happy() {
	a := HelperGetStockAction()
	err := MS.StoreAction(a)
	suite.True(err == nil)

	err = MS.DeleteAction(a.ActionID)
	suite.True(err == nil)

	_, err = MS.GetAction(a.ActionID)
	suite.False(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetAction_Error() {
	_, err := MS.GetAction(uuid.New())
	suite.True(err != nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetAction_Happy() {
	a := HelperGetStockAction()
	err := MS.StoreAction(a)
	suite.True(err == nil)

	aRet, err := MS.GetAction(a.ActionID)
	logrus.Info(aRet)
	logrus.Info(a)
	suite.True(err == nil)
	suite.True(aRet.Equals(a))

	err = MS.DeleteAction(a.ActionID)
	suite.True(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetActions() {
	a1 := HelperGetStockAction()
	err := MS.StoreAction(a1)
	suite.True(err == nil)
	a2 := HelperGetStockAction()

	actionArr, err := MS.GetActions()
	suite.True(err == nil)
	count1 := 0
	count2 := 0
	for _, d := range actionArr {
		if d.ActionID == a1.ActionID {
			suite.True(d.Equals(a1))
			count1++
		}
		if d.ActionID == a2.ActionID {
			suite.True(d.Equals(a2))
			count2++
		}
	}
	suite.True(count1 == 1)
	suite.True(count2 == 0)

	err = MS.StoreAction(a2)
	suite.True(err == nil)

	actionArr, err = MS.GetActions()
	suite.True(err == nil)
	count1 = 0
	count2 = 0
	for _, d := range actionArr {
		if d.ActionID == a1.ActionID {
			suite.True(d.Equals(a1))
			count1++
		}
		if d.ActionID == a2.ActionID {
			suite.True(d.Equals(a2))
			count2++
		}
	}
	suite.True(count1 == 1)
	suite.True(count2 == 1)

	err = MS.DeleteAction(a1.ActionID)
	suite.True(err == nil)

	err = MS.DeleteAction(a2.ActionID)
	suite.True(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_DeleteOperation_Error() {
	err := MS.DeleteOperation(uuid.New())
	suite.True(err != nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_DeleteOperation_Happy() {
	o := HelperGetStockOperation()
	err := MS.StoreOperation(o)
	suite.True(err == nil)

	err = MS.DeleteOperation(o.OperationID)
	suite.True(err == nil)

	_, err = MS.GetOperation(o.OperationID)
	suite.False(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetOperation_Error() {
	_, err := MS.GetOperation(uuid.New())
	suite.True(err != nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetOperation_Happy() {
	o := HelperGetStockOperation()
	err := MS.StoreOperation(o)
	suite.True(err == nil)

	oRet, err := MS.GetOperation(o.OperationID)
	suite.True(err == nil)
	// Set Time so we can check equality
	o.StartTime = oRet.StartTime
	o.RefreshTime = oRet.RefreshTime
	logrus.SetLevel(logrus.TraceLevel)
	logrus.Trace(oRet)
	logrus.Trace(o)
	suite.True(oRet.Equals(o))

	//err = MS.DeleteOperation(o.OperationID)
	//suite.True(err == nil)
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_GetOperations() {
	a := HelperGetStockAction()
	err := MS.StoreAction(a)
	suite.True(err == nil)

	o1 := HelperGetStockOperation()
	o1.ActionID = a.ActionID
	err = MS.StoreOperation(o1)
	suite.True(err == nil)
	o2 := HelperGetStockOperation()
	o2.ActionID = a.ActionID
	err = MS.StoreOperation(o2)
	suite.True(err == nil)

	operationArr, err := MS.GetOperations(a.ActionID)
	suite.True(err == nil)
	count1 := 0
	count2 := 0
	for _, d := range operationArr {
		if d.OperationID == o1.OperationID {
			// Set Time so we can check equality
			d.RefreshTime = o1.RefreshTime
			suite.True(d.Equals(o1))
			count1++
		}
		if d.OperationID == o2.OperationID {
			// Set Time so we can check equality
			d.RefreshTime = o2.RefreshTime
			suite.True(d.Equals(o2))
			count2++
		}
	}
	suite.True(count1 == 1)
	suite.True(count2 == 1)

	err = MS.DeleteOperation(o1.OperationID)
	suite.True(err == nil)

	err = MS.DeleteOperation(o2.OperationID)
	suite.True(err == nil)

	err = MS.DeleteAction(a.ActionID)
	suite.True(err == nil)

	a2 := HelperGetStockAction()
	err = MS.StoreAction(a)
	suite.True(err == nil)
	operationArr, err = MS.GetOperations(a2.ActionID)
	suite.True(err == nil)
}
