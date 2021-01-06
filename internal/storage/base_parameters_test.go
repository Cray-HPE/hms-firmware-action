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

	"github.com/stretchr/testify/suite"
)

type Base_Parameters_TS struct {
	suite.Suite
}

// SetupSuit is run ONCE
func (suite *Base_Parameters_TS) SetupSuite() {
}

func (suite *Base_Parameters_TS) Test_StateComponentFilter_Equals() {
	sf1 := StateComponentFilter{}
	sf2 := StateComponentFilter{}
	sf1.Groups = HelperRandStringSlice(3)
	sf2.Partitions = HelperRandStringSlice(3)
	suite.False(sf1.Equals(sf2))
	suite.False(sf2.Equals(sf1))
	sf1.Groups = HelperRandStringSlice(1)
	sf2.Groups = HelperRandStringSlice(1)
	suite.False(sf1.Equals(sf2))
	suite.False(sf2.Equals(sf1))
	sf2.Groups[0] = sf1.Groups[0]
	suite.False(sf1.Equals(sf2))
	suite.False(sf2.Equals(sf1))
	sf2.Partitions = HelperRandStringSlice(1)
	sf1.Partitions = HelperRandStringSlice(1)
	suite.False(sf1.Equals(sf2))
	suite.False(sf2.Equals(sf1))
	sf1.Partitions[0] = sf2.Partitions[0]
	suite.True(sf1.Equals(sf2))
	suite.True(sf2.Equals(sf1))
}

func (suite *Base_Parameters_TS) Test_InventoryHardwareFilter_Equals() {
	if1 := InventoryHardwareFilter{
		Manufacturer: "cray",
		Model:        "f43",
	}
	if2 := InventoryHardwareFilter{
		Manufacturer: "cray",
		Model:        "f46",
	}
	suite.True(if1.Equals(if1))
	suite.True(if2.Equals(if2))
	suite.False(if1.Equals(if2))
	suite.False(if2.Equals(if1))
}

func (suite *Base_Parameters_TS) Test_InventoryHardwareFilter_Empty() {
	if1 := InventoryHardwareFilter{
		Manufacturer: "cray",
		Model:        "f43",
	}
	suite.False(if1.Empty())
	if1.Manufacturer = ""
	suite.False(if1.Empty())
	if1.Model = ""
	suite.True(if1.Empty())
}

func (suite *Base_Parameters_TS) Test_TargetFilter_Equals() {
	t1 := TargetFilter{}
	t2 := TargetFilter{}
	t1.Targets = HelperRandStringSlice(3)
	t2.Targets = HelperRandStringSlice(2)
	suite.False(t1.Equals(t2))
	suite.False(t2.Equals(t1))
	t2.Targets = HelperRandStringSlice(3)
	suite.False(t1.Equals(t2))
	suite.False(t2.Equals(t1))
	t2.Targets[0] = t1.Targets[1]
	t2.Targets[1] = t1.Targets[2]
	t2.Targets[2] = t1.Targets[0]
	suite.True(t1.Equals(t2))
	suite.True(t2.Equals(t1))
}

func Test_Storage_Base_Parameters(t *testing.T) {
	//This setups the production routs and handler
	suite.Run(t, new(Base_Parameters_TS))
}
