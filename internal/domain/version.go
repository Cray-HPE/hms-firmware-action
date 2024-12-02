/*
 * MIT License
 *
 * (C) Copyright [2020-2024] Hewlett Packard Enterprise Development LP
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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Cray-HPE/hms-firmware-action/internal/hsm"
	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
	rf "github.com/Cray-HPE/hms-smd/pkg/redfish"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func drainAndCloseBody(resp *http.Response) {
	// Must always drain and close response bodies
	if resp != nil || resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body) // ok even if already drained
		resp.Body.Close()
	}
}

func contains(s []string, str string) bool {
	for _, v := range s{
		if v == str {
			return true
		}
	}
	return false
}

//passing in a copy!
func GetCurrentFirmwareVersionsFromHsmDataAndTargets(hd map[hsm.XnameTarget]hsm.HsmData) (deviceMap map[string]storage.Device, errors []string) {
	badDevices, errors := PruneXnameTargetList(&hd)
	goodDevices, err := RetrieveFirmwareVersionFromTargets(&hd)
	if err != nil {
		errors = append(errors, err.Error())
	}
	deviceMap = FullJoinDeviceMap(badDevices, goodDevices)
	return deviceMap, errors
}

func GetCurrentFirmwareVersionsFromParams(params storage.SnapshotParameters) (devices []storage.Device, errlist []string) {
	hsmDataMap := make(map[string]hsm.HsmData)
	//first pass -> filter for xnames | if the struct is empty it will get ALL xnames
	hsmDataMap, errs := (*GLOB.HSM).FillHSMData(params.StateComponentFilter.Xnames,
		params.StateComponentFilter.Partitions,
		params.StateComponentFilter.Groups,
		params.StateComponentFilter.DeviceTypes)

	if len(errs) > 0 {
		for _, value := range errs {
			errlist = append(errlist, value.Error())
		}
	}

	//Get the target data
	_, MatchedXnameTargets, _ := FilterTargets(&hsmDataMap, params.TargetFilter)

	XnameTargetHSMMap := make(map[hsm.XnameTarget]hsm.HsmData)
	for key, value := range MatchedXnameTargets {
		XnameTargetHSMMap[MatchedXnameTargets[key]] = hsmDataMap[value.Xname]
	}

	//TODO Perhaps move THIS to a global? Not going to do it In June of 2020 b.c it works and
	//  I dont want to spend the time monkeying around with it!
	specialTargets := make(map[string]string)
	specialTargets["node0.bios"] = "/redfish/v1/Systems/Node0"
	specialTargets["node1.bios"] = "/redfish/v1/Systems/Node1"

	(*GLOB.HSM).RefillModelRF(&XnameTargetHSMMap, specialTargets)

	FilterModelManufacturer(&XnameTargetHSMMap, params.InventoryHardwareFilter)
	devicesThatareNOTDiscoveredOK, errr := PruneXnameTargetList(&XnameTargetHSMMap)
	if len(errr) > 0 {
		errlist = append(errlist, errr...)
	}

	goodDevices, err := RetrieveFirmwareVersionFromTargets(&XnameTargetHSMMap)
	if err != nil {
		errlist = append(errlist, err.Error())
	}
	imageMap := GetImageMap()

	//fill in the ImageID on the target if possible; this will help us if we need to restore!
	FillInImageIDForDevices(&goodDevices, &XnameTargetHSMMap, &imageMap)

	finalDeviceMap := FullJoinDeviceMap(devicesThatareNOTDiscoveredOK, goodDevices)
	devices = FlattenDeviceMap(finalDeviceMap)
	//pb = model.BuildSuccessPassback(http.StatusOK, devices)
	return
}

func FillInImageIDForDevices(deviceMap *map[string]storage.Device, hsmData *map[hsm.XnameTarget]hsm.HsmData, imageMap *map[uuid.UUID]storage.Image) {
	for _, device := range *deviceMap {
		for targetnum, target := range device.Targets {
			//create the xnametarget to do the lookup
			xnametarget := hsm.XnameTarget{
				Xname:  device.Xname,
				Target: target.Name,
			}
			if devData, ok := (*hsmData)[xnametarget]; ok {
				for _, image := range *imageMap {
					var found bool
					if (len(image.SoftwareIds) > 0) && contains(image.SoftwareIds,target.SoftwareId) {
						found = true
					} else {
						_, found = model.Find(image.Models, devData.Model)
					}
					if found &&
						strings.EqualFold(image.DeviceType, devData.Type) &&
						strings.EqualFold(image.Manufacturer, devData.Manufacturer) &&
						image.Target == target.Name &&
						image.FirmwareVersion == target.FirmwareVersion {
						device.Targets[targetnum].ImageID = image.ImageID
					}
				}
				if target.ImageID == uuid.Nil && target.Error == nil {
					target.Error = errors.New("could not find a suitable image for target")
				}
			}
		}
	}
	return
}

func FullJoinDeviceMap(A map[string]storage.Device, B map[string]storage.Device) (C map[string]storage.Device) {
	C = make(map[string]storage.Device)
	for key, value := range A {
		if device, ok := C[key]; ok { // it exists
			for _, src_target := range value.Targets {
				var found bool
				for _, dst_target := range device.Targets {
					if src_target.Name == dst_target.Name {
						found = true
					}
				}
				if !found {
					device.Targets = append(device.Targets)
				}
			}
			if device.Error == nil && value.Error != nil {
				device.Error = value.Error
			}
			C[key] = device
		} else { // it does not YET exist
			C[key] = value
		}
	}

	for key, value := range B {
		if device, ok := C[key]; ok { // it exists
			for _, src_target := range value.Targets {
				var found bool
				for _, dst_target := range device.Targets {
					if src_target.Name == dst_target.Name {
						found = true
					}
				}
				if !found {
					device.Targets = append(device.Targets)
				}
			}
			if device.Error == nil && value.Error != nil {
				device.Error = value.Error
			}
			C[key] = device
		} else { // it does not YET exist
			C[key] = value
		}
	}
	return
}

func FlattenDeviceMap(A map[string]storage.Device) (B []storage.Device) {
	for _, val := range A {
		B = append(B, val)
	}
	return
}

// PruneXnameTargetList -> if there is an xnametarget, whose hsmdata isnt DiscoverOK, then we CANT get
// the fw version, so kick it out.
func PruneXnameTargetList(hd *map[hsm.XnameTarget]hsm.HsmData) (badDeviceMap map[string]storage.Device, errors []string) {
	//prune the BAD !DiscoveredOk ones
	badDeviceMap = make(map[string]storage.Device)

	badHsmDataMap := make(map[string]hsm.HsmData)

	//Kickout !DiscoveredOK
	for xnameTarget, datum := range *hd {
		if datum.DiscInfo.LastStatus != rf.DiscoverOK { //TODO, should we do this?  shouldnt the device be DiscoveredOK?
			datum.Error = fmt.Errorf("%s discovery status: %s", datum.ID, datum.DiscInfo.LastStatus)
			errors = append(errors, datum.Error.Error())
			//PUT in bad, take out of general population
			badHsmDataMap[xnameTarget.Xname] = datum
			delete(*hd, xnameTarget)
		}
	}

	//load bad hsmData into Bad Device!
	for xname, datum := range badHsmDataMap {
		if device, ok := badDeviceMap[xname]; ok {
			device.Error = datum.Error
			badDeviceMap[xname] = device
		} else { // cannot find the device in the map yet
			device := storage.Device{
				Xname: xname,
				Error: datum.Error,
			}
			badDeviceMap[xname] = device
		}
	}
	return
}

func RetrieveFirmwareVersionFromTargets(hd *map[hsm.XnameTarget]hsm.HsmData) (deviceMap map[string]storage.Device, err error) {
	if len(*hd) == 0 {
		err = errors.New("No Viable Targets")
		logrus.Error(err)
		return
	}
	deviceMap = make(map[string]storage.Device)
	taskMap := make(map[uuid.UUID]hsm.XnameTarget)
	taskList := (*GLOB.RFTloc).CreateTaskList(GLOB.BaseTRSTask, len(*hd))

	counter := 0
	for xnameTarget, _ := range *hd {
		if xnameTarget.Version != "" {
			var theErr error
			updateVer := model.DeviceFirmwareVersion{
				Version: xnameTarget.Version,
				Name:    xnameTarget.TargetName,
			}
			updateDeviceMap(deviceMap, updateVer, xnameTarget, theErr)
			// We have already found the version, so we are not going to add to tasklist
			// Reduce tasklist length by one to avoid looking for a blank record
			if len(taskList) > 0 {
				taskList = taskList[:len(taskList)-1]
			}
			continue
		}
		hsmdata := (*hd)[xnameTarget]
		taskMap[taskList[counter].GetID()] = xnameTarget
		urlStr, _ := GetFirmwareVersionURL(hsmdata, xnameTarget.Target)
		taskList[counter].Request.URL, _ = url.Parse(urlStr)
		taskList[counter].Timeout = time.Second * 40
		taskList[counter].CPolicy.Retry.Retries = 3

		if !(hsmdata.User == "" && hsmdata.Password == "") {
			taskList[counter].Request.SetBasicAuth(hsmdata.User, hsmdata.Password)
		}
		counter++
	}

	// Only execute tasklist if we have items, otherwise we get errors back
	if len(taskList) > 0 {
		(*GLOB.RFClientLock).RLock()
		defer (*GLOB.RFClientLock).RUnlock()
		rchan, err := (*GLOB.RFTloc).Launch(&taskList)
		if err != nil {
			logrus.Error(err)
		}

		for _, _ = range taskList {
			tdone := <-rchan
			var theErr error
			var body []byte
			var updateVer model.DeviceFirmwareVersion
			xnameTarget := taskMap[tdone.GetID()]

			for i := 0; i < 1; i++ { //artificial scope -> DO NOT DELETE THIS; IM NOT KIDDING!
				// I am doing this because I want to BREAK out and handle storing the 'error' into a Target 1 time instead of Copying the 20 lines of code 5 times.
				// the alternative design was a GOTO; with a continue in the happy case to NOT rewrite the success with an error; this is simpler and easier to read.
				// FOR REAL though, if you delete this, may you be haunted by cobol programmers & may your next job involve writing software on windows 2000

				if *tdone.Err != nil {
					theErr = *tdone.Err
					logrus.Error(theErr)
					break
				}
				if tdone.Request.Response.StatusCode < 200 && tdone.Request.Response.StatusCode >= 300 {
					theErr = errors.New("bad status code: " + strconv.Itoa(tdone.Request.Response.StatusCode))
					logrus.Error(theErr)
					break
				}
				if tdone.Request.Response.Body == nil {
					theErr = errors.New("empty body")
					logrus.Error(theErr)
					break
				}
				body, err = ioutil.ReadAll(tdone.Request.Response.Body)
				if err != nil {
					theErr = err
					logrus.Error(theErr)
					break
				}
				err = json.Unmarshal(body, &updateVer)
				if err != nil {
					theErr = err
					logrus.Error(theErr)
					break
				}
				// FINALLY!!!! ok; it should be good data!
				//Its possible that OLD cray bmc code may exist that corrupts that makes this struct empty...
				// its because a wrapping set of {} may be missing...
				// im taking the logic out that checks for that, b/c its too confusing!  we think this is no longer an issue;
				//so if this fails we know we have to put it back!
				if updateVer.Version == "" {
					if updateVer.BiosVersion != "" {
						updateVer.Version = updateVer.BiosVersion
					} else if updateVer.FirmwareVersion != "" {
						updateVer.Version = updateVer.FirmwareVersion
					}
				}
			} // END OF ARTIFICAL SCOPE  -> Still not kidding about deleting this.
			updateDeviceMap(deviceMap, updateVer, xnameTarget, theErr)

			drainAndCloseBody(tdone.Request.Response)
		}
		(*GLOB.RFTloc).Close(&taskList)
		close(rchan)
	} else {
		// Guard against possible leaks
		(*GLOB.RFTloc).Close(&taskList)
	}
	return
}

func updateDeviceMap(deviceMap map[string]storage.Device, updateVer model.DeviceFirmwareVersion, xnameTarget hsm.XnameTarget, theErr error) {
	target := storage.Target{
		Name: xnameTarget.Target,
	}
	if device, ok := deviceMap[xnameTarget.Xname]; ok {
		var foundTarget bool
		foundTarget = false
		for k, v := range device.Targets {
			if v.Name == xnameTarget.Target { //foundTarget the target!
				if theErr != nil {
					target.Error = theErr
				} else {
					target.FirmwareVersion = updateVer.Version
					target.SoftwareId = updateVer.SoftwareId
					target.TargetName = updateVer.Name
				}
				device.Targets[k] = target
				foundTarget = true
			}
		}
		if !foundTarget { // cannot find THIS target in targets of device
			if theErr != nil {
				target.Error = theErr
				logrus.Error(theErr)
			} else {
				target.FirmwareVersion = updateVer.Version
				target.SoftwareId = updateVer.SoftwareId
				target.TargetName = updateVer.Name
			}
			device.Targets = append(device.Targets, target)
		}
		deviceMap[xnameTarget.Xname] = device
	} else { // cannot find the device in the map yet
		device := storage.Device{
			Xname:   xnameTarget.Xname,
			Targets: nil,
		}
		if theErr != nil {
			target.Error = theErr
			logrus.Error(theErr)
		} else {
			target.FirmwareVersion = updateVer.Version
			target.SoftwareId = updateVer.SoftwareId
			target.TargetName = updateVer.Name
		}
		device.Targets = append(device.Targets, target)
		deviceMap[xnameTarget.Xname] = device
	}
}

func GetTaskLinkURL(data hsm.HsmData, tasklink string) (retURL string, err error) {
	err = nil
	retURL = "https://" + data.FQDN + tasklink
	return retURL, err
}

func GetUpdateInfoURL(data hsm.HsmData, updateinfolink string) (retURL string, err error) {
	err = nil
	retURL = "https://" + data.FQDN + updateinfolink
	return retURL, err
}

func GetFirmwareVersionURL(data hsm.HsmData, target string) (retURL string, err error) {
	rfEndpt := data.InventoryURI + "/" + target
	if data.InventoryURI == "" {
		err = fmt.Errorf("Could not recognize device/target: %s/%s", data.ID, target)
	}
	retURL = "https://" + data.FQDN + rfEndpt
	return retURL, err
}

func RetrieveUpdateInfo(hd *hsm.HsmData, updateinfolink string) (updateInfo model.UpdateInfo, err error) {
	var updateInfoRaw model.UpdateInformation
	urlStr, _ := GetUpdateInfoURL(*hd, updateinfolink)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logrus.Error(err)
		return
	}

	if !(hd.User == "" && hd.Password == "") {
		req.SetBasicAuth(hd.User, hd.Password)
	}

	reqContext, _ := context.WithTimeout(context.Background(), time.Second*40)
	req = req.WithContext(reqContext)
	if err != nil {
		logrus.Error(err)
		return
	}

	(*GLOB).RFClientLock.RLock()	// TODO: Do we really need locks?
	resp, err := (*GLOB).RFHttpClient.Do(req)
	(*GLOB).RFClientLock.RUnlock()
	defer drainAndCloseBody(resp)
	if err != nil {
		logrus.Error(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		return
	}

	err = json.Unmarshal(body, &updateInfoRaw)
	if err != nil {
		logrus.Error(err)
		return
	}

	updateInfo.FlashPercentage = updateInfoRaw.Oem.AMIUpdateService.UpdateInformation.FlashPercentage
	updateInfo.UpdateStatus = updateInfoRaw.Oem.AMIUpdateService.UpdateInformation.UpdateStatus
	updateInfo.UpdateTarget = updateInfoRaw.Oem.AMIUpdateService.UpdateInformation.UpdateTarget
	return
}

func RetrieveTaskStatus(hd *hsm.HsmData, tasklink string) (stateStatus model.TaskStateStatus, err error) {
	urlStr, _ := GetTaskLinkURL(*hd, tasklink)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logrus.Error(err)
		return
	}

	if !(hd.User == "" && hd.Password == "") {
		req.SetBasicAuth(hd.User, hd.Password)
	}

	reqContext, _ := context.WithTimeout(context.Background(), time.Second*40)
	req = req.WithContext(reqContext)
	if err != nil {
		logrus.Error(err)
		return
	}

	(*GLOB).RFClientLock.RLock()	// TODO: Do we really need locks?
	resp, err := (*GLOB).RFHttpClient.Do(req)
	(*GLOB).RFClientLock.RUnlock()
	defer drainAndCloseBody(resp)
	if err != nil {
		logrus.Error(err)
		return
	}
	if resp.StatusCode >= 400 {
		err = fmt.Errorf("Task Status Code: %d", resp.StatusCode)
		logrus.Error(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		return
	}

	err = json.Unmarshal(body, &stateStatus)
	if err != nil {
		logrus.Error(err)
		return
	}
	return
}

func RetrieveFirmwareVersion(hd *hsm.HsmData, target string) (firmwareVersion string, err error) {
	var updateVer model.DeviceFirmwareVersion
	urlStr, _ := GetFirmwareVersionURL(*hd, target)

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logrus.Error(err)
		return
	}

	if !(hd.User == "" && hd.Password == "") {
		req.SetBasicAuth(hd.User, hd.Password)
	}

	reqContext, _ := context.WithTimeout(context.Background(), time.Second*40)
	req = req.WithContext(reqContext)
	if err != nil {
		logrus.Error(err)
		return
	}

	(*GLOB).RFClientLock.RLock()
	resp, err := (*GLOB).RFHttpClient.Do(req)
	(*GLOB).RFClientLock.RUnlock()
	defer drainAndCloseBody(resp)
	if err != nil {
		logrus.Error(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err)
		return
	}

	err = json.Unmarshal(body, &updateVer)
	if err != nil {
		logrus.Error(err)
		return
	}

	if updateVer.Version == "" {
		if updateVer.BiosVersion != "" {
			updateVer.Version = updateVer.BiosVersion
		} else if updateVer.FirmwareVersion != "" {
			updateVer.Version = updateVer.FirmwareVersion
		}
	}
	firmwareVersion = updateVer.Version
	logrus.Trace(firmwareVersion)
	return firmwareVersion, err
}
