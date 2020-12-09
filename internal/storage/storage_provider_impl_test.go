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
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

//var MS MemStorage
var MS StorageProvider

func (suite *Storage_Provider_TS) SetupSuite() {
}

type Storage_Provider_TS struct {
	suite.Suite
}

func (suite *Storage_Provider_TS) Test_Storage_Provider_Ping() {
	err := MS.Ping()
	suite.True(err == nil)
}

func Test_Mem_Suite(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)

	storage := os.Getenv("STORAGE")
	if storage == "MEMORY" {
		var mem MemStorage
		MS = &mem
		err := MS.Init(logger)
		if err != nil {
			logrus.Panic("failed to init mem")
		}

		suite.Run(t, new(Storage_Provider_TS))
	} else if storage == "ETCD" {

		var etc ETCDStorage
		MS = &etc
		err := MS.Init(logger)
		if err != nil {
			logrus.Panic("failed to init etcd")
		}
		suite.Run(t, new(Storage_Provider_TS))

	} else if storage == "BOTH" {

		var etc ETCDStorage
		MS = &etc
		err := MS.Init(logger)
		if err != nil {
			logrus.Panic("failed to init etcd")
		}
		suite.Run(t, new(Storage_Provider_TS))

		var mem MemStorage
		MS = &mem
		err = MS.Init(logger)
		if err != nil {
			logrus.Panic("failed to init mem")
		}

		suite.Run(t, new(Storage_Provider_TS))

	} else {
		var mem MemStorage
		MS = &mem
		err := MS.Init(logger)
		if err != nil {
			logrus.Panic("failed to init mem")
		}

		suite.Run(t, new(Storage_Provider_TS))
	}
}

//func Test_Playground(t *testing.T) {
//
//	os.Setenv("ETCD_HOST", "localhost")
//	os.Setenv("ETCD_PORT", "2379")
//	logger := logrus.New()
//	logger.SetLevel(logrus.TraceLevel)
//
//	tmpStorageImplementation := &ETCDStorage{
//		Logger: logger,
//	}
//	MS = tmpStorageImplementation
//
//	err := MS.Init(logger)
//	if err != nil {
//		logrus.Panic("failed to init etcd")
//	}
//
//	a := HelperGetStockAction()
//	err = MS.StoreAction(a)
//	x, err := MS.GetAction(a.ActionID)
//
//	logrus.Info(x, err)
//}
