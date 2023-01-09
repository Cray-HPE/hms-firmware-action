/*
 * MIT License
 *
 * (C) Copyright [2022-2023] Hewlett Packard Enterprise Development LP
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
	"net/http"

	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
	"github.com/sirupsen/logrus"
)

func DumpDB() (pb model.Passback) {
	logrus.Info("Dumping DB")
	var database storage.Db

	actions, _ := GetStoredActions()
	for _, action := range actions {
		database.Actions = append(database.Actions, storage.ToActionStorable(action))
	}
	operations, _ := GetAllOperations()
	for _, operation := range operations {
		database.Operations = append(database.Operations, storage.ToOperationStorable(operation))
	}
	database.Images, _ = GetStoredImages()
	snapshots, _ := GetStoredSnapshots()
	for _, snapshot := range snapshots {
		database.Snapshots = append(database.Snapshots, storage.ToSnapshotStorable(snapshot))
	}
	var err error
	err = nil
	if err == nil {
		pb = model.BuildSuccessPassback(http.StatusOK, database)
	} else {
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
	}
	return
}

func LoadDB(db storage.Db) (pb model.Passback) {
	logrus.Info("Loading DB")
	for i := 0; i < len(db.Actions); i++ {
		logrus.Info("** action " + db.Actions[i].ActionID.String())
		action := storage.ToActionFromStorable(db.Actions[i])
		StoreAction(action)
	}
	for i := 0; i < len(db.Operations); i++ {
		logrus.Info("** operation " + db.Operations[i].OperationID.String())
		operation := storage.ToOperationFromStorable(db.Operations[i])
		StoreOperation(operation)
	}
	for i := 0; i < len(db.Images); i++ {
		logrus.Info("** image " + db.Images[i].ImageID.String())
		StoreImage(db.Images[i])
	}
	for i := 0; i < len(db.Snapshots); i++ {
		snapshot := storage.ToSnapshotFromStorable(db.Snapshots[i])
		StoreSnapshot(snapshot)
	}

	pb = model.BuildSuccessPassback(http.StatusNoContent, nil)
	return
}
