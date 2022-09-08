/*
 * MIT License
 *
 * (C) Copyright [2020-2022] Hewlett Packard Enterprise Development LP
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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	hmetcd "github.com/Cray-HPE/hms-hmetcd"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	//kvUrlDefault = "https://localhost:2379"
	//kvUrlDefault = "http://localhost:2379"
	kvUrlDefault     = "mem:" // Use in memory KV store for the time being.
	kvRetriesDefault = 5
	keyPrefix        = "/fas/"
	keyMin           = " "
	keyMax           = "~"
)

type ETCDStorage struct {
	Logger   *logrus.Logger
	mutex    sync.Mutex
	kvHandle hmetcd.Kvi
}

func (e *ETCDStorage) fixUpKey(k string) string {
	key := k
	if !strings.HasPrefix(k, keyPrefix) {
		key = keyPrefix
		if strings.HasPrefix(k, "/") {
			key += k[1:]
		} else {
			key += k
		}
	}
	return key
}

func (e *ETCDStorage) kvStore(key string, val interface{}) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	data, err := json.Marshal(val)
	if err == nil {
		realKey := e.fixUpKey(key)
		err = e.kvHandle.Store(realKey, string(data))
	}
	return err
}

func (e *ETCDStorage) kvGet(key string, val interface{}) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	realKey := e.fixUpKey(key)
	v, exists, err := e.kvHandle.Get(realKey)
	if exists {
		// We have a key, so val is valid.
		err = json.Unmarshal([]byte(v), &val)
	} else if err == nil {
		// No key and no error.  We will return this condition as an error
		err = fmt.Errorf("Key %s does not exist", key)
	}
	return err
}

//if a key doesnt exist, etcd doesn't return an error
func (e *ETCDStorage) kvDelete(key string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	realKey := e.fixUpKey(key)
	e.Logger.Trace("delete" + realKey)
	return e.kvHandle.Delete(e.fixUpKey(key))
}

func (e *ETCDStorage) Init(Logger *logrus.Logger) (err error) {
	e.Logger = Logger
	var kverr error

	retries := kvRetriesDefault
	host, hostExists := os.LookupEnv("ETCD_HOST")
	if !hostExists {
		e.kvHandle = nil
		err = fmt.Errorf("No ETCD HOST specified, can't open ETCD.")
		return
	}
	port, portExists := os.LookupEnv("ETCD_PORT")
	if !portExists {
		e.kvHandle = nil
		err = fmt.Errorf("No ETCD PORT specified, can't open ETCD.")
		return
	}

	kvURL := fmt.Sprintf("http://%s:%s", host, port)
	e.Logger.Info(kvURL)

	etcOK := false
	for ix := 1; ix <= retries; ix++ {
		e.kvHandle, kverr = hmetcd.Open(kvURL, "")
		if kverr != nil {
			e.Logger.Error("ERROR opening connection to ETCD (attempt ", ix, "):", kverr)
		} else {
			etcOK = true
			e.Logger.Info("ETCD connection succeeded.")
			break
		}
	}
	if !etcOK {
		e.kvHandle = nil
		err = fmt.Errorf("ETCD connection attempts exhausted, can't connect.")
	}
	return err
}

func (e *ETCDStorage) Ping() (err error) {
	e.Logger.Debug("ETCD PING")
	key := fmt.Sprintf("/ping/%s", uuid.New().String())
	err = e.kvStore(key, "")
	if err == nil {
		err = e.kvDelete(key)
	}
	return
}

func (e *ETCDStorage) StoreAction(a Action) (err error) {
	key := fmt.Sprintf("/actions/%s", a.ActionID.String())
	storable := ToActionStorable(a)
	err = e.kvStore(key, storable)
	if err != nil {
		e.Logger.Error(err)
	}
	return
}

func (e *ETCDStorage) DeleteAction(actionID uuid.UUID) (err error) {
	_, err = e.GetAction(actionID)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("/actions/%s", actionID.String())
	err = e.kvDelete(key)
	if err != nil {
		e.Logger.Error(err)
	}
	return
}

func (e *ETCDStorage) GetAction(actionID uuid.UUID) (a Action, err error) {
	key := fmt.Sprintf("/actions/%s", actionID.String())

	var retrieveable ActionStorable
	err = e.kvGet(key, &retrieveable)
	if err != nil {
		e.Logger.Error(err)
	}
	a = ToActionFromStorable(retrieveable)
	return a, err
}

func (e *ETCDStorage) GetActions() (a []Action, err error) {
	k := e.fixUpKey("/actions/")
	kvl, err := e.kvHandle.GetRange(k+keyMin, k+keyMax)
	if err == nil {
		for _, kv := range kvl {
			var act ActionStorable
			err = json.Unmarshal([]byte(kv.Value), &act)
			if err != nil {
				e.Logger.Error(err)
			} else {
				newAct := ToActionFromStorable(act)
				a = append(a, newAct)
			}
		}
	} else {
		e.Logger.Error(err)
	}
	return a, err
}

// StoreOperation -> as part of this process we will delete the hsmdata; creds!
func (e *ETCDStorage) StoreOperation(o Operation) (err error) {
	//Reset the refresh time, its vital to know if something died in progress
	o.RefreshTime.Scan(time.Now())
	storable := ToOperationStorable(o)
	key := fmt.Sprintf("/operations/%s", o.OperationID.String())
	err = e.kvStore(key, storable)
	if err != nil {
		e.Logger.Error(err)
	}
	return
}
func (e *ETCDStorage) DeleteOperation(operationID uuid.UUID) (err error) {
	_, err = e.GetOperation(operationID)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("/operations/%s", operationID.String())
	err = e.kvDelete(key)
	if err != nil {
		e.Logger.Error(err)
	}
	return
}

func (e *ETCDStorage) GetOperation(operationID uuid.UUID) (o Operation, err error) {
	key := fmt.Sprintf("/operations/%s", operationID.String())
	var retrieveable OperationStorable
	err = e.kvGet(key, &retrieveable)
	if err != nil {
		e.Logger.Error(err)
	}
	o = ToOperationFromStorable(retrieveable)

	return
}

func (e *ETCDStorage) GetOperations(actionID uuid.UUID) (o []Operation, err error) {
	action, err := e.GetAction(actionID)
	if err != nil {
		e.Logger.Error(err)
		return
	}
	for _, opid := range action.OperationIDs {
		op, err := e.GetOperation(opid)
		if err != nil {
			e.Logger.Error(err)
		} else {
			o = append(o, op)
		}
	}
	return o, err
}

func (e *ETCDStorage) StoreSnapshot(s Snapshot) (err error) {
	key := fmt.Sprintf("/snapshots/%s", s.Name)

	storable := ToSnapshotStorable(s)
	logrus.Info(storable)
	err = e.kvStore(key, storable)
	if err != nil {
		e.Logger.Error(err)
	}
	return
}
func (e *ETCDStorage) GetSnapshot(name string) (ss Snapshot, err error) {
	key := fmt.Sprintf("/snapshots/%s", name)
	var retrieveable SnapshotStorable
	err = e.kvGet(key, &retrieveable)
	if err != nil {
		e.Logger.Error(err)
		return ss, err
	} else {
		ss = ToSnapshotFromStorable(retrieveable)
		return ss, err
	}

	return
}
func (e *ETCDStorage) GetSnapshots() (s []Snapshot, err error) {
	k := e.fixUpKey("/snapshots/")
	kvl, err := e.kvHandle.GetRange(k+keyMin, k+keyMax)
	if err == nil {
		for _, kv := range kvl {
			var ss SnapshotStorable
			err = json.Unmarshal([]byte(kv.Value), &ss)
			if err != nil {
				e.Logger.Error(err)
			} else {
				snp := ToSnapshotFromStorable(ss)
				s = append(s, snp)
			}
		}
	} else {
		e.Logger.Error(err)
	}
	return
}
func (e *ETCDStorage) DeleteSnapshot(name string) (err error) {
	_, err = e.GetSnapshot(name)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("/snapshots/%s", name)
	err = e.kvDelete(key)
	if err != nil {
		e.Logger.Error(err)
	}
	return
}

func (e *ETCDStorage) GetImages() (i []Image, err error) {
	k := e.fixUpKey("/images/")
	kvl, err := e.kvHandle.GetRange(k+keyMin, k+keyMax)
	if err == nil {
		for _, kv := range kvl {
			var img Image
			err = json.Unmarshal([]byte(kv.Value), &img)
			if err != nil {
				e.Logger.Error(err)
			} else {
				i = append(i, img)
			}
		}
	} else {
		e.Logger.Error(err)
	}
	return
}
func (e *ETCDStorage) GetImage(imageID uuid.UUID) (i Image, err error) {
	key := fmt.Sprintf("/images/%s", imageID.String())
	err = e.kvGet(key, &i)
	if err != nil {
		e.Logger.Error(err)
	}
	return
}
func (e *ETCDStorage) StoreImage(i Image) (err error) {
	key := fmt.Sprintf("/images/%s", i.ImageID.String())
	err = e.kvStore(key, i)
	if err != nil {
		e.Logger.Error(err)
	}
	return
}
func (e *ETCDStorage) DeleteImage(imageID uuid.UUID) (err error) {
	_, err = e.GetImage(imageID)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("/images/%s", imageID.String())
	err = e.kvDelete(key)
	if err != nil {
		e.Logger.Error(err)
	}
	return
}
