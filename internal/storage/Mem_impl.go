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
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type MemStorage struct {
	Logger     *logrus.Logger
	mutex      sync.Mutex
	Actions    map[uuid.UUID]Action
	Operations map[uuid.UUID]Operation
	Images     map[uuid.UUID]Image
	Snapshots  map[string]Snapshot
}

func (b *MemStorage) Init(Logger *logrus.Logger) (err error) {
	b.Logger = Logger

	b.Actions = make(map[uuid.UUID]Action)
	b.Operations = make(map[uuid.UUID]Operation)
	b.Images = make(map[uuid.UUID]Image)
	b.Snapshots = make(map[string]Snapshot)

	return err
}
func (b *MemStorage) Ping() (err error) {
	b.Logger.Debug("MEMORY PING")
	return err
}

// err always nil
func (b *MemStorage) StoreSnapshot(s Snapshot) (err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.Snapshots[s.Name] = s
	return err
}

func (b *MemStorage) DeleteSnapshot(name string) (err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if _, ok := b.Snapshots[name]; ok {
		delete(b.Snapshots, name)
	} else {
		err = errors.New("could not find key")
		b.Logger.WithField("snapshot", name).Error(err)
	}
	return err
}

func (b *MemStorage) GetSnapshot(name string) (s Snapshot, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if s, ok := b.Snapshots[name]; ok {
		return s, nil
	} else {
		err = errors.New("could not find key")
		b.Logger.WithField("snapshotID", name).Error(err)
	}
	return s, err
}

func (b *MemStorage) GetSnapshots() (s []Snapshot, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, val := range b.Snapshots {
		s = append(s, val)
	}
	return s, err
}

// err is always nil
func (b *MemStorage) StoreAction(a Action) (err error) {
	// Get the current state of the stored action to see if
	// it has been signaled to stop
	curStAction, curExists := b.GetAction(a.ActionID)
	if curExists == nil {
		if curStAction.State.Is("abortSignaled") {
			// Change to signal abort if possible
			if a.State.Can("signalAbort") == true {
				logrus.Info("Changed State from " + a.State.Current() + " to abortSignaled")
				a.State.Event("signalAbort")
			}
		}
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.Actions[a.ActionID] = a
	a.RefreshTime.Scan(time.Now()) //need to make sure we always update refresh time
	return err
}

func (b *MemStorage) DeleteAction(actionID uuid.UUID) (err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if _, ok := b.Actions[actionID]; ok {
		delete(b.Actions, actionID)
	} else {
		err = errors.New("could not find key")
		b.Logger.WithField("actionID", actionID.String()).Error(err)

	}
	return err
}

func (b *MemStorage) GetAction(actionID uuid.UUID) (a Action, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if a, ok := b.Actions[actionID]; ok {
		return a, nil
	} else {
		err = errors.New("could not find key")
		b.Logger.WithField("actionID", actionID.String()).Error(err)
	}
	return a, err
}

// err always nil
func (b *MemStorage) GetActions() (a []Action, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, val := range b.Actions {
		a = append(a, val)
	}
	return a, err
}

// err always nil
func (b *MemStorage) StoreOperation(o Operation) (err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	//Reset the refresh time, its vital to know if something died in progress
	o.RefreshTime.Scan(time.Now())
	b.Operations[o.OperationID] = o
	return err
}

func (b *MemStorage) DeleteOperation(operationID uuid.UUID) (err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if _, ok := b.Operations[operationID]; ok {
		delete(b.Operations, operationID)
	} else {
		err = errors.New("could not find key")
		b.Logger.WithField("operationID", operationID.String()).Error(err)
	}
	return err
}

func (b *MemStorage) GetOperation(operationID uuid.UUID) (o Operation, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if o, ok := b.Operations[operationID]; ok {
		return o, nil
	} else {
		err = errors.New("could not find key")
		b.Logger.WithField("operationID", operationID.String()).Error(err)
	}
	return o, err
}

func (b *MemStorage) GetOperations(actionID uuid.UUID) (o []Operation, err error) {
	action, err := b.GetAction(actionID)
	if err != nil {
		b.Logger.Error(err)
		return
	}
	for _, opid := range action.OperationIDs {
		op, err := b.GetOperation(opid)
		if err != nil {
			b.Logger.Error(err)
		} else {
			o = append(o, op)
		}
	}
	return o, err
}

// err is always nil
func (b *MemStorage) StoreImage(i Image) (err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.Images[i.ImageID] = i
	return err
}

func (b *MemStorage) DeleteImage(imageID uuid.UUID) (err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if _, ok := b.Images[imageID]; ok {
		delete(b.Images, imageID)
	} else {
		err = errors.New("could not find key")
		b.Logger.WithField("imageID", imageID.String()).Error(err)
	}
	return err
}

func (b *MemStorage) GetImage(imageID uuid.UUID) (i Image, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if i, ok := b.Images[imageID]; ok {
		return i, nil
	} else {
		err = errors.New("could not find key")
		b.Logger.WithField("imageID", imageID.String()).Error(err)
	}
	return i, err
}

// err always nil
func (b *MemStorage) GetImages() (i []Image, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, val := range b.Images {
		i = append(i, val)
	}
	return i, err
}
