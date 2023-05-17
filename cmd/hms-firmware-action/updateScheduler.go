/*
 * MIT License
 *
 * (C) Copyright [2020-2023] Hewlett Packard Enterprise Development LP
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

//TODO need to consider separating the DOMAIN  - (API stuff, from the Control Loop stuff)!
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Cray-HPE/hms-firmware-action/internal/domain"
	"github.com/Cray-HPE/hms-firmware-action/internal/hsm"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
	rf "github.com/Cray-HPE/hms-smd/pkg/redfish"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type VerifyStatus int

const (
	UpdateSuccess VerifyStatus = iota
	FailNoChange
	FailUnexpectedChange
)

var loopDelay = time.Duration(5) * time.Second

type PayloadCray struct {
	ImageURI         string   `json:"ImageURI"`
	TransferProtocol string   `json:"TransferProtocol"`
	Targets          []string `json:"Targets"`
}

type PayloadGigabyte struct {
	ImageURI         string `json:"ImageURI"`
	TransferProtocol string `json:"TransferProtocol"`
	UpdateComponent  string `json:"UpdateComponent"`
}

type PayloadHpe struct {
	ImageURI string `json:"ImageURI"`
}

//TODO -> currently the code only allows one action at a time.  We want to do one action at a time, and block if they
//  use the same xnames, but how do I check if another request has things locked? like another action?
//  how do I check if I already have the lock? How does 'restartability' play into this?
//TODO locks are only in memory right now, should they go to etcd? + if it dies inbetween doLaunch and doVerify it
//doesnt get the lock again

// controlLoop -> runs forever
// General flow: GET ALL ACTIVE ACTIONS (that is actions that are not completed, aborted or new (not yet configured)
// 		The code is in a particular order to make sure only one action is transitioned into running at a time,
//		thereby making sure only one thing can run at a time. DONT CHANGE IT!
//		ABORT -> if the action has signaled abort (via the API) then ABORT everything in it, and if this was
//			lastRunningAction then clear it
//		RUNNING -> the action is running, see what can be fired off (could doLaunch, or doVerify, or check if its unblocked).
//			Also check if it should be marked completed.  If it is that way and was lastRunningAction then clear it
//		CONFIGURED -> the action can be run.  Check if something else has the lastRunningAction title.  IF it does, then
//			block on that, else START and take the flag
// 		BLOCKED -> will either BLOCK or set to CONFIGURE.  It will look through BlockedBy and compare to the other action
//			states, if they are all aborted or completed then set this back to configure
// Regarding restartability -> the action states will be constant on a restart of FAS, so in theory the only thing lost is
// the indivdual operation in progress.  If its running, then check the RefreshTimer (which gets reset everytime its stored)
// if its been 10 mins since last refresh then restart last checkpoint
func controlLoop(domainGlobal *domain.DOMAIN_GLOBALS) {
	mainLogger.Debug("CONTROL LOOP - @BEGIN")
	var restart = true
	quitChannels := make(map[uuid.UUID]chan bool)

	// Check for unfinished snapshots
	checkExpiredItems := 0

	//If FAS dies while things are in doLaunch or doVerify, it will use the operation.RefreshTime to know when to try
	//again (10 mins after last refresh)
	for ; Running; time.Sleep(loopDelay) {

		// Check for expired snapshots and actions
		// Do not have to do this every time through the loop
		if checkExpiredItems <= 0 {
			domain.DeleteExpiredSnapshots()
			domain.DeleteExpiredActions(domainGlobal.DaysToKeepActions)
			checkExpiredItems = 500
		}
		checkExpiredItems--

		mainLogger.Debug("CONTROL LOOP - @TOP")
		//returns all "running" &  "configured"
		actions := domain.GetAllNonCompleteNonInitialActions()
		if len(actions) == 0 {
			restart = false
			continue
		}
		var lastRunningAction uuid.UUID
		lastRunningAction = uuid.Nil

		for k, action := range actions {

			//the action has signaled abort
			if action.State.Is("abortSignaled") {
				mainLogger.WithField("actionID", action.ActionID).Debug("CONTROL LOOP - @ABORTING")
				ops, err := domain.GetAllOperationsFromAction(action.ActionID)
				if err != nil {
					mainLogger.Error(err)
				}
				for _, op := range ops {
					//I will send a true on the quit channel, but I dont close the channel. Eventually I do delete it

					//We have a bit of a Pompeii problem.  There will be operations that in blocked/configured state, or
					//are in needsVerified (that haven't yet been verified) that b/c they have NO quit channel listening
					//cannot be aborted. Those operations wont ever run, b/c the action will get aborted, but they will stay
					//at whatever state they were in.  Which might be good, or might not be... Not sure. Probably doesnt matter.
					//I think the take away is we dont want to leave something 'in progress' or give that impression.
					if quitChannel, ok := quitChannels[op.OperationID]; ok {
						mainLogger.WithField("operationID", op.OperationID).Debug("signaling quit to operation")
						select {
						// If quitChannel queue is full, continue on
						case quitChannel <- true:
							mainLogger.WithFields(logrus.Fields{"operationID": op.OperationID}).Debug("TRUE Sent to QUIT CHANNEL")
						default:
							mainLogger.WithFields(logrus.Fields{"operationID": op.OperationID}).Debug("MESSAGE NOT SENT")
						}
						mainLogger.Debug("Signal Sent - Deleting Map Entry")

						//https://nanxiao.gitbooks.io/golang-101-hacks/content/posts/need-not-close-every-channel.html
						// according to ^ you dont need to close every chan when youve finished with it; garbage collection can take care of it
						delete(quitChannels, op.OperationID) //it doesnt hurt to delete from a map; but that shouldnt happen b/c we checked to see if it existed
						mainLogger.WithFields(logrus.Fields{"operationID": op.OperationID, "err": err}).Debug("deleted chan")
					} else {
						//There is no quit channel, so it wont hurt if we do this, b/c no one else is going to try to!
						op.State.Event("abort")
						op.EndTime.Scan(time.Now())
						op.StateHelper = "abort received from abort loop"
						mainLogger.WithFields(logrus.Fields{"operationID": op.OperationID}).Debug("aborted operation")
						domain.StoreOperation(op)
					}

					err := (*domainGlobal.HSM).ClearLock([]string{op.Xname})
					if err != nil {
						mainLogger.WithFields(logrus.Fields{"operationID": op.OperationID, "err": err}).Error("failed to unlock")
						op.Error = errors.New("Failed to unlock node")
						domain.StoreOperation(op)
					}
				}
				action.State.Event("abort")
				action.EndTime.Scan(time.Now())
				domain.StoreAction(action)

				if lastRunningAction == action.ActionID {
					lastRunningAction = uuid.Nil
				}

				//verify if the action is still blocked
			} else if action.State.Is("running") {
				if lastRunningAction == uuid.Nil {
					lastRunningAction = action.ActionID
				}
				mainLogger.WithField("actionID", action.ActionID).Debug("CONTROL LOOP - @RUNNING")

				operations := domain.GetAllActiveOperationsFromAction(action.ActionID)
				for opnum, operation := range operations {
					ToImagePB := domain.GetImageStorage(operation.ToImageID)
					if ToImagePB.IsError {
						mainLogger.Error(ToImagePB.Error.Detail)
						operation.Error = errors.New(ToImagePB.Error.Detail)
						operation.State.Event("fail")
						operation.StateHelper = "could not find the image"
						operation.EndTime.Scan(time.Now())
						domain.StoreOperation(operation)
						continue
					}

					FromImagePB := domain.GetImageStorage(operation.FromImageID)
					var FromImage storage.Image
					if !FromImagePB.IsError {
						FromImage = FromImagePB.Obj.(storage.Image)
					}

					var ToImage storage.Image
					ToImage = ToImagePB.Obj.(storage.Image)

					var quitChan chan bool
					if _, ok := quitChannels[operation.OperationID]; !ok {
						// Add a queue to channel to make non-blocking
						quitChan = make(chan bool, 3)
						quitChannels[operation.OperationID] = quitChan
					} else {
						quitChan = quitChannels[operation.OperationID]
					}

					tripper := operation.RefreshTime.Time.Add(time.Duration(10) * time.Minute)
					now := time.Now()
					hasTipped := now.After(tripper)
					//Launch or relaunch things
					if operation.State.Is("configured") {
						mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID}).Debug("starting doLaunch")
						go doLaunch(operation, ToImage, action.Command, domainGlobal, quitChan)
					} else if operation.State.Is("needsVerified") {
						mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID}).Debug("starting doVerify")
						go doVerify(operation, ToImage, FromImage, domainGlobal, quitChan)
					} else if operation.State.Is("inProgress") && (hasTipped || restart) {
						mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID}).Warn("restarting doLaunch, operation failed to refresh")
						go doLaunch(operation, ToImage, action.Command, domainGlobal, quitChan)
					} else if operation.State.Is("verifying") && (hasTipped || restart) {
						mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID}).Warn("restarting doVerify, operation failed to refresh")
						go doVerify(operation, ToImage, FromImage, domainGlobal, quitChan)
					}
					operations[opnum] = operation
				}

				//TODO note we might need a separate concept once @DependencyManagment has been re-introduced.
				totalOperations, _ := domain.GetAllOperationsFromAction(action.ActionID)

				//decided to make this sync (vs async) because we want to maintain the thread of control
				domain.CheckBlockage(&totalOperations)

				//Check if the whole thing is done!
				counts := domain.GetOperationSummaryFromAction(action.ActionID)
				if counts.Total == counts.Aborted+counts.NoSolution+counts.NoOperation+counts.Succeeded+counts.Failed {
					mainLogger.WithField("actionID", action.ActionID).Debug("operations complete, finishing action")
					action.State.Event("finish")
					action.EndTime.Scan(time.Now())
					for _, op := range totalOperations {
						err := (*domainGlobal.HSM).ClearLock([]string{op.Xname})
						if err != nil {
							mainLogger.WithFields(logrus.Fields{"operationID": op.OperationID, "err": err}).Error("failed to unlock")
							op.Error = errors.New("Failed to unlock node")
							domain.StoreOperation(op)
						}
					}
					if lastRunningAction == action.ActionID {
						lastRunningAction = uuid.Nil
					}
					domain.StoreAction(action)
				}

				//the action is configured, see if anythin else is running
			} else if action.State.Is("configured") && action.State.Can("start") {
				mainLogger.WithField("actionID", action.ActionID).Debug("CONTROL LOOP - @CONFIGURED")

				if lastRunningAction == uuid.Nil { //everthing else so far has been aborted or completed.
					action.State.Event("start")
					mainLogger.WithFields(logrus.Fields{"actionID": action.ActionID}).Debug("action is starting")
					lastRunningAction = action.ActionID

				} else {
					action.BlockedBy = append(action.BlockedBy, lastRunningAction)
					mainLogger.WithFields(logrus.Fields{"actionID": action.ActionID, "blockingAction": lastRunningAction}).Debug("action is blocked waiting for action to complete")
					action.State.Event("block")
				}
				actions[k] = action
				domain.StoreAction(action)

				//the action is blocked, see if we can unblock it
			} else if action.State.Is("blocked") {
				mainLogger.WithField("actionID", action.ActionID).Debug("CONTROL LOOP - @BLOCKER")
				blocked := false
				for _, v := range action.BlockedBy {
					blocker := domain.GetActionState(v)
					if blocker.ActionID == uuid.Nil {
						continue //b.c so far its still false
					}
					if !(blocker.State.Is("completed") || blocker.State.Is("aborted")) {
						blocked = true
						break
					}
				}
				if !blocked {
					//then unblock
					action.State.Event("unblock")
					domain.StoreAction(action)
					mainLogger.WithField("actionID", action.ActionID).Debug("action unblocked")
				}
			}
		}
		restart = false
	}
}

// doLaunch -> will check the file exists, lock the xname, perform the update.
// Parameters:
//		operation -> WHAT to do
//		image -> what to update to
//		command -> HOW the action was configured (override, dryrun)
//		globals -> connection to domain layer stuff (hsm, dsp)
//		quit -> a channel that we listen on so we know when to quit
// At each stage/transition it will re-store the operation back to persistent storage.  This may seem excessive, but it
// is vitally important so we know what has been done.
func doLaunch(operation storage.Operation, image storage.Image, command storage.Command, globals *domain.DOMAIN_GLOBALS, quit <-chan bool) {
	var err error

	//This COULD be a re-launch, in which case we need to restart
	if operation.State.Can("start") {
		err = operation.State.Event("start")
		if err != nil {
			mainLogger.Error(err)
		}
		operation.StateHelper = "preparing to launch"
		operation.Error = nil
		domain.StoreOperation(operation)
	} else if operation.State.Can("restart") {
		err = operation.State.Event("restart")
		if err != nil {
			mainLogger.Error(err)
		}
		operation.StateHelper = "preparing to re-launch"
		operation.Error = nil
		domain.StoreOperation(operation)
	} else {

		operation.Error = errors.New("invalid state, leaving doLaunch")
		mainLogger.WithField("operationID", operation.OperationID).Error(operation.Error)
		domain.StoreOperation(operation)
		return
	}

	var timer time.Duration
	timer = time.Now().AddDate(1, 0, 0).Sub(time.Now()) //ONE year in the future
	if operation.ExpirationTime.Valid {
		timer = operation.ExpirationTime.Time.Sub(operation.StartTime.Time)
	}
	timeout := time.After(timer)

	var pollingTime time.Time
	pollingSpeed := time.Duration(image.PollingSpeedSeconds) * time.Second
	pollingTime = time.Now().Add(pollingSpeed)
	//check if it is blacklistsed Only update compute nodes
	if strings.EqualFold(operation.HsmData.Type, "nodebmc") {
		blacklisted := false

		var role string
		role = operation.HsmData.Role
		//need to go to HSM and get component data for the n0 attached to this bmc.  Since its a nodeBMC it should
		//always have something under it, but if it doesnt, well we just ignore the role then.  ONLY nodes have roles;
		//so if it doesnt have a node, then it cannot have a role that we care about.
		var emptyArray []string
		xnames := []string{operation.HsmData.ID + "n0"}
		components, _ := HSM.GetStateComponents(xnames, emptyArray, emptyArray, emptyArray)
		//In this case we dont care about an error; because we are looking for 1 and ONLY 1 component;
		//if n0 doesnt exist we will probably get an empty components list with a 400 error; but we dont care; the ONLY
		//other plan would be to default blacklist if it fails these checks, but that doesnt seem right

		if len(components.Components) == 1 {
			role = components.Components[0].Role
			mainLogger.Debug("using child n0 role for parent")
		}

		for _, v := range nodeBlacklist {
			if strings.EqualFold(role, v) {
				blacklisted = true
			}
		}
		if blacklisted {

			operation.State.Event("nosol")
			operation.StateHelper = "Can not update node, black listed: " + role
			operation.Error = nil
			operation.EndTime.Scan(time.Now())
			err := (*globals.HSM).ClearLock([]string{operation.Xname})
			if err != nil {
				mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("failed to unlock")
				operation.Error = errors.New("Failed to unlock node")
			}
			domain.StoreOperation(operation)
			return
		}
	}
	//TODO in the future we need to consider a ROLLBACK possibility.
	//if its a dry run we want to check the file & powerState, but NOT lock the device
	var isFile, isLock, isPowerState bool

	if !command.OverrideDryrun { //casmhms-3642 -> not override; == DO A DRYRUN
		isLock = true
	}

	if len(image.AllowableDeviceStates) == 0 {
		isPowerState = true
	}

	var updateURL string
	for ; ; time.Sleep(time.Duration(1) * time.Second) {
		select {
		case <-quit: //signal stop
			mainLogger.WithField("operationID", operation.OperationID).Debug("operation aborted")
			operation.State.Event("abort")
			operation.EndTime.Scan(time.Now())
			operation.StateHelper = "abort received from quit in doLaunch"
			err := (*globals.HSM).ClearLock([]string{operation.Xname})
			if err != nil {
				mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("failed to unlock")
				operation.Error = errors.New("Failed to unlock node")
			}
			domain.StoreOperation(operation)
			return
		case <-timeout: //expiration time
			mainLogger.WithField("operationID", operation.OperationID).Debug("expiration time for  operation exceeded")
			operation.State.Event("fail")
			operation.StateHelper = "time expired; could not complete update"
			err := (*globals.HSM).ClearLock([]string{operation.Xname})
			if err != nil {
				mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("failed to unlock")
				operation.Error = errors.New("Failed to unlock node")
			}
			domain.StoreOperation(operation)
			return

		default:
			//if I try the file check and it has an err, will we ever find it? not likely, but it is possible that S3 is
			// not functioning correctly. The easiest thing to do is just keep trying.  The expiration time will eventually trip
			if !isFile {
				operation.StateHelper = "attempting to check file"
				mainLogger.WithField("operationID", operation.OperationID).Debug(operation.StateHelper)
				updateURL, err = fileCheck(image.S3URL)
				if err != nil {
					operation.Error = err
					operation.StateHelper = "failed to find file, trying again soon"
				} else {
					isFile = true
					operation.StateHelper = "found file"
					operation.Error = nil
				}
				mainLogger.WithField("operationID", operation.OperationID).Debug(operation.StateHelper)
				domain.StoreOperation(operation)

			} else if !isLock {
				operation.StateHelper = "attempting to lock"
				mainLogger.WithField("operationID", operation.OperationID).Debug(operation.StateHelper)
				lckErr := (*globals.HSM).SetLock([]string{operation.Xname})
				if lckErr != nil {
					mainLogger.WithFields(logrus.Fields{"xname": operation.Xname, "operationID": operation.OperationID, "lockMessage": lckErr}).Warn("could not lock component, trying again soon.")
					operation.Error = err
					operation.StateHelper = "failed to lock, trying again soon"
				} else {
					isLock = true
					operation.Error = nil
					operation.StateHelper = "got lock"
				}
				mainLogger.WithField("operationID", operation.OperationID).Debug(operation.StateHelper)
				domain.StoreOperation(operation)

			} else if !isPowerState {
				if time.Now().After(pollingTime) {
					pollingTime = time.Now().Add(pollingSpeed) //reset the poller
					powerState, err := getPowerState(&operation.HsmData)
					if err != nil {
						mainLogger.Error(err)
						operation.Error = err
						operation.StateHelper = "could not get power state"
						isPowerState = false
					}

					for _, v := range image.AllowableDeviceStates {
						if strings.ToLower(v) == strings.ToLower(powerState) {
							operation.StateHelper = "power state satisfied: " + powerState
							isPowerState = true
							break
						}
					}
					if isPowerState == false {
						//We assume Off or rebooting
						operation.StateHelper = "reboot not satisfied, powerstate: " + powerState
					}
					domain.StoreOperation(operation)
				}
			} else if isLock && isFile && isPowerState {

				if operation.FromImageID == uuid.Nil && !command.RestoreNotPossibleOverride {
					operation.State.Event("nosol")
					operation.StateHelper = "cannot perform the update as the override was not enabled and there is no image to go back to."
					operation.EndTime.Scan(time.Now())
					operation.Error = nil
					mainLogger.Debug(operation.StateHelper)
					err := (*globals.HSM).ClearLock([]string{operation.Xname})
					if err != nil {
						mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("failed to unlock")
						operation.Error = errors.New("Failed to unlock node")
					}
					domain.StoreOperation(operation)
					return
				}

				//OK, now that we have verified the power, lock and file -> time to update.  You get ONE CHANCE to do this
				// If manufacturer is blank, then we will copy manufacturer from image record.
				// We can do this because we made it this far and want to flash the image we selected.
				if operation.HsmData.Manufacturer == "" {
					mainLogger.Debug("Opearation Manufacturer is blank, setting to: " + strings.ToLower(image.Manufacturer))
					operation.HsmData.Manufacturer = strings.ToLower(image.Manufacturer)
					operation.Manufacturer = strings.ToLower(image.Manufacturer)
					domain.StoreOperation(operation)
				}
				var passback model.Passback
				passback = model.BuildErrorPassback(http.StatusTeapot, errors.New("by default, this has failed"))
				if command.OverrideDryrun {
					if strings.EqualFold(operation.HsmData.Manufacturer, manufacturerIntel) {
						path := operation.HsmData.InventoryURI + "/" + operation.Target + "/Actions/Oem/Intel.Oem.Update" + operation.Target
						file := "images/" + updateURL
						operation.StateHelper = "sending intel payload"
						operation.Error = nil
						mainLogger.Debug(operation.StateHelper)
						domain.StoreOperation(operation)

						passback = SendSecureRedfishFileUpload(globals, operation.HsmData.FQDN, path, "upload", file,
							operation.HsmData.User, operation.HsmData.Password)
					} else if strings.EqualFold(operation.HsmData.Manufacturer, manufacturerCray) {
						operation.StateHelper = "sending cray payload"
						operation.Error = nil
						mainLogger.Debug(operation.StateHelper)
						domain.StoreOperation(operation)

						pc := PayloadCray{
							ImageURI:         updateURL,
							TransferProtocol: "HTTP",
							Targets:          []string{operation.HsmData.InventoryURI + "/" + operation.Target},
						}

						pcm, _ := json.Marshal(pc)
						pcs := string(pcm)
						passback = SendSecureRedfish(globals, operation.HsmData.FQDN, operation.HsmData.UpdateURI,
							pcs, operation.HsmData.User, operation.HsmData.Password, "POST")
					} else if strings.EqualFold(operation.HsmData.Manufacturer, manufacturerGigabyte) {

						// Need to replace hostname with IP address because gigabyte does not have a DNS server
						updateImageURI := updateURL
						u, err := url.Parse(updateImageURI)
						if err == nil {
							host, _, _ := net.SplitHostPort(u.Host)
							if len(host) == 0 {
								host = u.Host
							}
							addr, err := net.LookupIP(host)
							if err == nil && len(addr) > 0 {
								log.Println("Replacing: ", host, " with ", addr[0].String())
								updateImageURI = strings.Replace(updateImageURI, host, addr[0].String(), 1)
								operation.StateHelper = "sending gigabyte payload"
								operation.Error = nil
								mainLogger.Debug(operation.StateHelper)
								domain.StoreOperation(operation)

								pg := PayloadGigabyte{
									ImageURI:         updateImageURI,
									TransferProtocol: "HTTP",
									UpdateComponent:  operation.Target,
								}
								pgm, _ := json.Marshal(pg)
								pgs := string(pgm)
								passback = SendSecureRedfish(globals, operation.HsmData.FQDN, operation.HsmData.UpdateURI,
									pgs, operation.HsmData.User, operation.HsmData.Password, "POST")
								// Gigabyte provide update status from the UpdateService
								operation.UpdateInfoLink = "/redfish/v1/UpdateService"
							} else {
								mainLogger.Errorf("Could not replace hostname: %s", host)
							}
						} else {
							mainLogger.Errorf("Could not parse: %s", updateImageURI)
						}
					} else if strings.EqualFold(operation.HsmData.Manufacturer, manufacturerHPE) {
						operation.StateHelper = "sending hpe payload"
						operation.Error = nil
						mainLogger.Debug(operation.StateHelper)
						domain.StoreOperation(operation)

						pc := PayloadHpe{
							ImageURI: updateURL,
						}

						pcm, _ := json.Marshal(pc)
						pcs := string(pcm)
						passback = SendSecureRedfish(globals, operation.HsmData.FQDN, operation.HsmData.UpdateURI,
							pcs, operation.HsmData.User, operation.HsmData.Password, "POST")
						// iLO return a link to a task which we can monitor for update progress
						tasklink := new(model.TaskLink)
						err = json.Unmarshal(passback.Obj.([]byte), &tasklink)
						operation.TaskLink = tasklink.Link
						mainLogger.WithFields(logrus.Fields{"TaskLink": operation.TaskLink, "err": err}).Info("TASKLINK")
					} else {
						_ = operation.State.Event("fail")
						operation.Error = errors.New("unsupported manufacturer")
						mainLogger.Debug("Unspported Manufacturer - Can not send payload")
						passback = model.BuildErrorPassback(http.StatusBadRequest, operation.Error)
					}
				} else {
					operation.StateHelper = "dry run completed: Updated Image: " + image.FirmwareVersion
					_ = operation.State.Event("success")
					_ = operation.EndTime.Scan(time.Now())
					operation.Error = nil
					mainLogger.Debug(operation.StateHelper)
					domain.StoreOperation(operation)
					return
				}

				if passback.IsError || passback.StatusCode >= 400 { //if we HAVE an error; or if the status code is the error range 4XX, 5XX
					operation.Error = errors.New(passback.Error.Detail)
					operation.State.Event("fail")
					operation.StateHelper = "failed to update target - status code: " + strconv.Itoa(passback.StatusCode) + " - See operation for any error message"
					operation.EndTime.Scan(time.Now())

					err := (*globals.HSM).ClearLock([]string{operation.Xname})
					if err != nil {
						mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("failed to unlock")
						operation.Error = errors.New("Failed to unlock node")
					}
					domain.StoreOperation(operation)
					return
				} else {

					//needs verified! unless its dryrun
					if operation.State.Can("needsVerify") {
						operation.State.Event("needsVerify")
						operation.StateHelper = "update complete, needs verification"
						domain.StoreOperation(operation)
						return
					}
				}
			}
		}
	}
}

// doVerify -> will handle the reboot and then verify the firmware version
// Parameters:
//		operation -> WHAT to do
//		ToImage -> what to update to
//		FromImage -> what to update from
//		globals -> connection to domain layer stuff (hsm, dsp)
//		quit -> a channel that we listen on so we know when to quit
// At each stage/transition it will re-store the operation back to persistent storage.  This may seem excessive, but it
// is vitally important so we know what has been done.
// TODO review the automatic vs manual reboot satisfaction criteria and the tries criteria
func doVerify(operation storage.Operation, ToImage storage.Image, FromImage storage.Image, globals *domain.DOMAIN_GLOBALS, quit <-chan bool) {
	var err error
	err = nil

	//it is possible this is a re launch of doVerify
	if operation.State.Can("verifying") {
		err = operation.State.Event("verifying")
		if err != nil {
			mainLogger.Error(err)
		}
		operation.StateHelper = "verifying potential success"
		operation.Error = nil
		domain.StoreOperation(operation)
	} else if operation.State.Can("reverifying") {
		err = operation.State.Event("reverifying")
		if err != nil {
			mainLogger.Error(err)
		}
		operation.StateHelper = "preparing to re-attempt verifying"
		operation.Error = nil
		domain.StoreOperation(operation)
	} else {
		operation.Error = errors.New("invalid state, leaving doVerify")
		mainLogger.WithField("operationID", operation.OperationID).Error(operation.Error)
		err := (*globals.HSM).ClearLock([]string{operation.Xname})
		if err != nil {
			mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("failed to unlock")
			operation.Error = errors.New("Failed to unlock node")
		}
		domain.StoreOperation(operation)
		return
	}

	var timer time.Duration
	timer = time.Now().AddDate(1, 0, 0).Sub(time.Now()) //ONE year in the future
	if operation.ExpirationTime.Valid {
		timer = operation.ExpirationTime.Time.Sub(operation.StartTime.Time) //from start time, not NOW
	}
	timeout := time.After(timer)

	var pollingTime time.Time
	pollingSpeed := time.Duration(ToImage.PollingSpeedSeconds) * time.Second
	pollingTime = time.Now().Add(pollingSpeed)

	var manualRebootSatisfied bool
	manualRebootSatisfied = !(ToImage.NeedManualReboot) // the reboot is satisfied if it DOESNT need a reboot
	var automaticRebootSatisfied bool
	if ToImage.NeedManualReboot == false {
		automaticRebootSatisfied = false
	} else {
		automaticRebootSatisfied = true
	}

	defaultTimeToWait := time.Duration(2) * time.Minute

	allowedTries := 20

	entryTime := time.Now()
	var verifySatisfied bool

	var rebootStarted bool
	var rebootTime time.Time
	for ; ; time.Sleep(time.Duration(1) * time.Second) {
		select {
		case <-quit: //signal stop
			mainLogger.WithField("operationID", operation.OperationID).Debug("operation aborted")
			operation.State.Event("abort")
			operation.EndTime.Scan(time.Now())
			operation.StateHelper = "abort received from quit in doVerify"
			err := (*globals.HSM).ClearLock([]string{operation.Xname})
			if err != nil {
				mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("failed to unlock")
				operation.Error = errors.New("Failed to unlock node")
			}
			domain.StoreOperation(operation)
			return
		case <-timeout: //expiration time
			mainLogger.WithField("operationID", operation.OperationID).Debug("expiration time for  operation exceeded")
			operation.State.Event("fail")
			operation.StateHelper = "time expired; could not verify"
			err := (*globals.HSM).ClearLock([]string{operation.Xname})
			if err != nil {
				mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("failed to unlock")
				operation.Error = errors.New("Failed to unlock node")
			}
			domain.StoreOperation(operation)
			return
		default:
			if !automaticRebootSatisfied {
				if time.Now().After(entryTime.Add(defaultTimeToWait)) {
					automaticRebootSatisfied = true
				}
			} else if !manualRebootSatisfied {
				if time.Now().After(entryTime.Add(time.Duration(ToImage.WaitTimeBeforeManualRebootSeconds)*time.Second)) && !rebootStarted {
					//see https://cray.slack.com/archives/GJUBRT8US/p1588276620304200 for notes
					path := operation.HsmData.ActionReset.Target //"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"
					// check LOCK
					lckErr := (*globals.HSM).SetLock([]string{operation.Xname})
					if lckErr != nil {
						mainLogger.WithFields(logrus.Fields{"xname": operation.Xname, "operationID": operation.OperationID, "lockMessage": lckErr}).Warn("could not lock component, trying again soon.")
						operation.Error = err
						operation.StateHelper = "failed to lock for reset, trying again soon"
						domain.StoreOperation(operation)
					} else {
						passback := SendSecureRedfish(globals, operation.HsmData.FQDN, path, "{\"ResetType\":\""+ToImage.ForceResetType+"\"}", operation.HsmData.User, operation.HsmData.Password, "POST")
						//its possible we could get an error code, but we are really close to being done, should we ignore it? or FAIL the whole thing?

						if passback.IsError {
							mainLogger.WithField("err", passback.Error.Detail).Errorf("error encountered rebooting xname: %s", operation.Xname)
						} else {
							mainLogger.WithFields(logrus.Fields{"response": passback.Obj, "statusCode": passback.StatusCode}).Debugf("issued restart to xname: %s", operation.Xname)
						}
						rebootStarted = true
						rebootTime = time.Now()
						operation.StateHelper = "reboot command issued"
						domain.StoreOperation(operation)
					}
				} else {
					operation.StateHelper = "waiting to reboot"
					domain.StoreOperation(operation)
				}

				if time.Now().After(rebootTime.Add(time.Duration(ToImage.WaitTimeAfterRebootSeconds)*time.Second)) && rebootStarted {
					//Consider what happens if we get NO response, because its still rebooting?! That will probably be
					//an error
					if time.Now().After(pollingTime) {
						pollingTime = time.Now().Add(pollingSpeed) // reset it
						powerState, err := getPowerState(&operation.HsmData)
						if err != nil {
							mainLogger.Error(err)
							operation.Error = err
							operation.StateHelper = "could not get power state"
						} else if powerState == rf.POWER_STATE_ON {
							operation.StateHelper = "reboot satisfied"
							manualRebootSatisfied = true
						} else {
							//We assume Off or rebooting
							operation.StateHelper = "reboot not satisfied, powerstate: " + powerState
							manualRebootSatisfied = false
						}
						domain.StoreOperation(operation)
						//unfortuneately we cannot use the status/health of the FirmwareInventory/{endpoint} to determine health
						// of a update.  According to the RF spec, only OK, warning, and critical are supported. Gigabyte doesnt even do this,
						//and cray does 'updating'? but it auto reboots.  So best thing to do it make WHOMEVER creates an ToImage tell us timings.
					}
				}
			} else if !verifySatisfied && manualRebootSatisfied && automaticRebootSatisfied {
				if time.Now().After(pollingTime) {
					allowedTries-- //Decrease the count
					if allowedTries < 1 {
						//operation.Error = err // add a more meaningful error!
						operation.State.Event("fail")
						operation.StateHelper = "Firmware update failed verification"
						err := (*globals.HSM).ClearLock([]string{operation.Xname})
						if err != nil {
							mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("failed to unlock")
							operation.Error = errors.New("Failed to unlock node")
						}
						operation.EndTime.Scan(time.Now())
						mainLogger.Debug(operation.StateHelper)
						domain.StoreOperation(operation)
						return
					}
					pollingTime = time.Now().Add(pollingSpeed) // reset it
					firmwareVersion, err := domain.RetrieveFirmwareVersion(&operation.HsmData, operation.Target)
					if err != nil {
						mainLogger.WithFields(logrus.Fields{"err": err, "operationID": operation.OperationID}).Error("failed to retrieve firmware version")
						continue
					} else {
						var stat VerifyStatus
						if ToImage.FirmwareVersion == firmwareVersion {
							//THEN the version matches!
							stat = UpdateSuccess
							operation.StateHelper = "Update Successful to version: " + ToImage.FirmwareVersion
							//ITS on the old version still!!! FAIL After Timeout
						} else if operation.FromFirmwareVersion == firmwareVersion {
							stat = FailNoChange
							operation.StateHelper = "no change detected in firmware version"
						} else {
							//the version has changed but it was unexpected - So FAIL
							stat = FailUnexpectedChange
							operation.StateHelper = "unexpected change detected in firmware version. Expected " +
								ToImage.FirmwareVersion + " got: " + firmwareVersion
						}
						if stat == UpdateSuccess {
							verifySatisfied = true
							//SET SUCCESS
							operation.State.Event("success")
							operation.Error = nil
							operation.EndTime.Scan(time.Now())
							err := (*globals.HSM).ClearLock([]string{operation.Xname})
							if err != nil {
								mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("failed to unlock")
								operation.Error = errors.New("Failed to unlock node")
							}
							domain.StoreOperation(operation)
							return
						}
						// We dont just quit on a FailNoChange... b/c we give it time to rectify
						// but if its changed and its unexpected then we do fail
						if stat == FailUnexpectedChange {
							verifySatisfied = true
							//SET FAIL
							operation.State.Event("fail")
							operation.Error = nil
							operation.EndTime.Scan(time.Now())
							err := (*globals.HSM).ClearLock([]string{operation.Xname})
							if err != nil {
								mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("failed to unlock")
								operation.Error = errors.New("Failed to unlock node")
							}
							domain.StoreOperation(operation)
							return
						}
					}
					// UpdateInfoLink is currently only available on Gigabyte
					if operation.UpdateInfoLink != "" {
						updateInfo, err := domain.RetrieveUpdateInfo(&operation.HsmData, operation.UpdateInfoLink)
						if err != nil {
							mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("Update Info Check")
						} else {
							if updateInfo.UpdateTarget == operation.Target {
								if updateInfo.UpdateStatus == "Preparing" || updateInfo.UpdateStatus == "VerifyingFirmware" || updateInfo.UpdateStatus == "Downloading" {
									operation.StateHelper = "Firmware Update Information Returned " + updateInfo.UpdateStatus
									domain.StoreOperation(operation)
								} else if updateInfo.UpdateStatus == "Flashing" {
									operation.StateHelper = "Firmware Update Information Returned " + updateInfo.UpdateStatus + " " + updateInfo.FlashPercentage
									domain.StoreOperation(operation)
								} else if updateInfo.UpdateStatus == "" {
									operation.StateHelper = "Firmware Update Information Unavailable"
									domain.StoreOperation(operation)
								} else if updateInfo.UpdateStatus == "Completed" {
									operation.State.Event("success")
									operation.StateHelper = "Firmware Update Information Returned " + updateInfo.UpdateStatus + " " + updateInfo.FlashPercentage + " -- Reboot of node may be required"
									domain.StoreOperation(operation)
									return
								} else {
									operation.State.Event("fail")
									operation.StateHelper = "Firmware Update Information Returned " + updateInfo.UpdateStatus + " " + updateInfo.FlashPercentage + " -- See " + operation.UpdateInfoLink
									operation.Error = errors.New("See " + operation.UpdateInfoLink)
									domain.StoreOperation(operation)
									return
								}
							} else {
								mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "Update Info": updateInfo}).Error("Update Info Check - Targets don't match")
							}
						}
					}
					// TaskLink is currently only available on iLO
					if operation.TaskLink != "" {
						taskStatus, err := domain.RetrieveTaskStatus(&operation.HsmData, operation.TaskLink)
						if err != nil {
							mainLogger.WithFields(logrus.Fields{"operationID": operation.OperationID, "err": err}).Error("Task Status Check")
						} else {
							if taskStatus.TaskState == "Running" {
								operation.StateHelper = "Firmware Task Returned Running"
								domain.StoreOperation(operation)
							} else if taskStatus.TaskState == "Completed" && taskStatus.TaskStatus == "OK" {
								operation.State.Event("success")
								operation.StateHelper = "Firmware Task Returned " + taskStatus.TaskState + " with Status " + taskStatus.TaskStatus + " -- Reboot of node may be required"
								domain.StoreOperation(operation)
								return
							} else {
								operation.State.Event("fail")
								operation.StateHelper = "Firmware Task Returned " + taskStatus.TaskState + " with Status " + taskStatus.TaskStatus + " -- See " + operation.TaskLink
								operation.Error = errors.New("See " + operation.TaskLink)
								domain.StoreOperation(operation)
								return
							}
						}
					}
					domain.StoreOperation(operation) // Update RefreshTime
				}
			}
		}
	}
}

func fileCheck(fileLocation string) (returnLocation string, err error) {
	returnLocation = fileLocation
	URL, err := url.Parse(fileLocation)
	if err != nil {
		return returnLocation, err
	}

	if strings.ToLower(URL.Scheme) == "s3" {

		bucket := URL.Host //this helps us capture the bucket name // fw-update in s3://fw-update/f1.1123.24.xz.iso

		s3endpoint, err := url.Parse(S3_ENDPOINT)
		if err != nil {
			return returnLocation, err
		}

		URL.Host = s3endpoint.Host //ex: http://rgw.local:8080
		URL.Scheme = s3endpoint.Scheme
		URL.Path = bucket + URL.Path

		returnLocation = URL.String()
	}
	//else the scheme is http

	mainLogger.WithFields(logrus.Fields{"URL": returnLocation}).Debug("GETTING HEAD of FILE")
	response, err := http.Head(returnLocation)
	if err != nil {
		mainLogger.Error(err)
		return returnLocation, err
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New("unexpected status code; could not find file via HEAD")
		return returnLocation, err
	}

	return returnLocation, nil
}

func SendSecureRedfish(globals *domain.DOMAIN_GLOBALS, server string, path string, bodyStr string, authUser string,
	authPass string, method string) (pb model.Passback) {

	tmpURL, _ := url.Parse("https://" + server + path)
	req, err := http.NewRequest(method, tmpURL.String(), bytes.NewBuffer([]byte(bodyStr)))
	if err != nil {
		mainLogger.Error(err)
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		return
	}

	if !(authUser == "" && authPass == "") {
		req.SetBasicAuth(authUser, authPass)
	}
	reqContext, _ := context.WithTimeout(context.Background(), time.Second*40)
	req = req.WithContext(reqContext)
	if err != nil {
		mainLogger.Error(err)
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("cache-control", "no-cache")

	mainLogger.WithFields(logrus.Fields{"URL": tmpURL.String(), "body": bodyStr}).Debug("SENDING COMMAND")

	globals.RFClientLock.RLock()
	resp, err := globals.RFHttpClient.Do(req)
	globals.RFClientLock.RUnlock()
	if err != nil {
		mainLogger.Error(err)
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	mainLogger.WithFields(logrus.Fields{"response": string(body), "status": resp.StatusCode}).Debug("RECEIVED RESPONSE")
	if err != nil {
		mainLogger.Error(err)
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		pb.Error.Detail = string(body)
	} else {
		pb = model.BuildSuccessPassback(resp.StatusCode, body)
		pb.Error.Detail = string(body)
	}
	return
}

// Creates a new file upload http request with optional extra params
func SendSecureRedfishFileUpload(globals *domain.DOMAIN_GLOBALS, server string, path string, paramName string,
	filename string, authUser string, authPass string) (pb model.Passback) {

	file, err := os.Open(filename)
	if err != nil {
		mainLogger.Error(err)
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	part, err := writer.CreateFormFile(paramName, filepath.Base(filename))
	if err != nil {
		mainLogger.Error(err)
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		return
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		mainLogger.Error(err)
		pb = model.BuildErrorPassback(http.StatusBadRequest, err)
		return
	}

	tmpURL, _ := url.Parse("https://" + server + path)

	req, err := http.NewRequest("POST", tmpURL.String(), payload)
	if err != nil {
		mainLogger.Error(err)
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		return
	}
	if !(authUser == "" && authPass == "") {
		req.SetBasicAuth(authUser, authPass)
	}
	reqContext, _ := context.WithTimeout(context.Background(), time.Second*40)
	req = req.WithContext(reqContext)
	if err != nil {
		mainLogger.Error(err)
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		return
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	globals.RFClientLock.RLock()
	resp, err := globals.RFHttpClient.Do(req)
	globals.RFClientLock.RUnlock()
	if err != nil {
		mainLogger.Error(err)
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		mainLogger.Error(err)
		pb = model.BuildErrorPassback(http.StatusInternalServerError, err)
	} else {
		pb = model.BuildSuccessPassback(resp.StatusCode, body)
	}
	return
}

// TODO: Causing FAS to crash -- FIX MJB 20200603
func getPowerState(hd *hsm.HsmData) (powerState string, err error) {
	powerState = rf.POWER_STATE_ON
	/*
		passback := SendSecureRedfish(hd.FQDN, hd.BmcPath, "", hd.User, hd.Password, "GET")
		switch hd.RfType {
		case rf.ChassisType:
			var info rf.Chassis
			err = json.Unmarshal(passback.Obj.([]byte), &info)
			if err == nil {
				if info.PowerState != "" {
					powerState = info.PowerState
				} else {
					mainLogger.Debugf("no power state for (%s/%s) ; assuming 'On'",
						hd.ID, hd.RfType)
					powerState = rf.POWER_STATE_ON
				}
			}
		case rf.ComputerSystemType:
			var info rf.ComputerSystem
			err = json.Unmarshal(passback.Obj.([]byte), &info)
			if err == nil {
				powerState = info.PowerState
			}
		case rf.ManagerType:
			var info rf.Manager
			err = json.Unmarshal(passback.Obj.([]byte), &info)
			if err == nil {
				mainLogger.Debugf("Status: %v\n", info.Status)

				// Managers don't really have a power state so
				// we'll assume any reponse means 'On'
				powerState = rf.POWER_STATE_ON
			}
		default:
			mainLogger.Errorf("%s unknown Redfish Type\n", hd.RfType)
			// punt
			powerState = "UNKNOWN"
		}

		if powerState == rf.POWER_STATE_ON {
			//then its on
		}
	*/
	return
}
