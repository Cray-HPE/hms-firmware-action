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
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type Actions_TS struct {
	suite.Suite
}

// SetupSuit is run ONCE
func (suite *Actions_TS) SetupSuite() {
}

func (suite *Actions_TS) Test_Action_enterState() {
	// TODO
}

func (suite *Actions_TS) Test_Action_restoreState() {
	a := HelperGetStockAction()
	suite.True(a.State.Current() == "new")
	a.restoreState("running")
	suite.True(a.State.Current() == "running")
	a.restoreState("bad")
	suite.True(a.State.Current() == "failed")
}

func (suite *Actions_TS) Test_Action_NewAction() {
	ap1 := ActionParameters{}
	ap1.StateComponentFilter.Groups = HelperRandStringSlice(3)
	a := NewAction(ap1)
	a.restoreState("running")
}

func (suite *Actions_TS) Test_Operation_enterState() {
	// TODO
}

func (suite *Actions_TS) Test_Operation_restoreState() {
	a := HelperGetStockOperation()
	suite.True(a.State.Current() == "initial")
	a.restoreState("blocked")
	suite.True(a.State.Current() == "blocked")
	a.restoreState("bad")
	suite.True(a.State.Current() == "failed")
}

func (suite *Actions_TS) Test_ImageFilter_Equals() {
	if1 := ImageFilter{}
	if2 := ImageFilter{}
	if1.ImageID = uuid.New()
	if2.ImageID = if1.ImageID
	suite.True(if1.Equals(if2))
	suite.True(if2.Equals(if1))
	if2.ImageID = uuid.New()
	suite.False(if1.Equals(if2))
	suite.False(if2.Equals(if1))
}

func (suite *Actions_TS) Test_Command_Equals() {
	c1 := Command{}
	c1.OverrideDryrun = false
	c1.Version = "1.2.3"
	c1.Description = "Description"
	c2 := Command{}
	c2.OverrideDryrun = false
	c2.Version = "1.2.3"
	c2.Description = "Description"
	suite.True(c1.Equals(c2))
	c2.Version = "1.4.5"
	suite.False(c1.Equals(c2))
}

func (suite *Actions_TS) Test_ActionParameter_Equals() {
	ap1 := ActionParameters{}
	ap2 := ActionParameters{}
	ap1.StateComponentFilter.Groups = HelperRandStringSlice(3)
	ap2.StateComponentFilter.Groups = ap1.StateComponentFilter.Groups
	suite.True(ap1.Equals(ap2))
	suite.True(ap2.Equals(ap1))
	ap2.StateComponentFilter.Groups = HelperRandStringSlice(3)
	suite.False(ap1.Equals(ap2))
	suite.False(ap2.Equals(ap1))
}

func (suite *Actions_TS) Test_NewOperation() {
	o := *NewOperation()
	suite.True(o.OperationID != uuid.Nil)
}

/* CopyOperation was removed
func (suite *Images_TS) Test_CopyOperation() {
	o1 := HelperGetStockOperation()
	o2 := *CopyOperation(&o1)

	// o2 is a partial copy of o1, they are not identical
	suite.False(o1.Equals(o2))
}
*/

func (suite *Actions_TS) Test_Action_Equals() {
	a1 := HelperGetStockAction()
	a2 := HelperGetStockAction()

	suite.True(a1.Equals(a1))
	suite.True(a2.Equals(a2))
	suite.False(a1.Equals(a2))
	suite.False(a2.Equals(a1))
}

func (suite *Actions_TS) Test_Operation_Equals() {
	o1 := HelperGetStockOperation()
	o2 := HelperGetStockOperation()

	suite.True(o1.Equals(o1))
	suite.True(o2.Equals(o2))
	suite.False(o1.Equals(o2))
	suite.False(o2.Equals(o1))
}

func (suite *Actions_TS) Test_NewActionID() {
	a := NewActionID()
	aId1 := a.ActionID
	suite.True(aId1 != uuid.Nil)
}

func Test_Storage_Actions(t *testing.T) {
	//This setups the production routs and handler
	suite.Run(t, new(Actions_TS))
}
