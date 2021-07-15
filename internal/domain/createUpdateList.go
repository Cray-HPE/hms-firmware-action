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

package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/hsm"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/model"
	"stash.us.cray.com/HMS/hms-firmware-action/internal/storage"
)

//Create Update List for this Update ActionID and add to Master Update List
func GenerateOperations(actionID uuid.UUID) {
	action, err := (*GLOB.DSP).GetAction(actionID)
	if err != nil {
		logrus.WithFields(logrus.Fields{"ERROR": err}).Error("cannot retrieve action, cannot generate operations")
		return
	}

	//Ok this is a bit fluid; The general flow will be:
	// Generate initial xname list -> by using stateComponent filter
	// GetHSMData
	// Get targets
	//  Filter  by maufacturer/model
	//  Filter  by targets
	//  Filter by image->
	//    there  can only be ONE image specified; kick out any temp ops that dont make sense;
	//	        like if the image could NEVER be applicable based on target, device type, manaufacturer, model
	// THEN foreach operation:
	//    Fill out the image stuff
	//    find out if the image has depenencies
	//    create any dependent operations and queue them to be filled and have depenedencies created
	// store ops and action

	//STEP 1 -> filter for xnames | if the struct is empty it will get ALL xnames
	hsmDataMap, errs := (*GLOB.HSM).FillHSMData(action.Parameters.StateComponentFilter.Xnames,
		action.Parameters.StateComponentFilter.Partitions,
		action.Parameters.StateComponentFilter.Groups,
		action.Parameters.StateComponentFilter.DeviceTypes)

	if len(errs) > 0 {
		for _, value := range errs {
			action.Errors = append(action.Errors, value.Error())
		}
	}
	err = (*GLOB.DSP).StoreAction(action)
	if err != nil {
		logrus.Error(err)
	}
	//STEP 2 -> Get the target data based on the reduced hsmDataMap; and filter it accordingly
	_, MatchedXnameTargets, _ := FilterTargets(&hsmDataMap, action.Parameters.TargetFilter)

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

	//STEP 3 -> Filter on Models/Manufactuer -> reduce the number of entries in the hsmdata map
	FilterModelManufacturer(&XnameTargetHSMMap, action.Parameters.InventoryHardwareFilter)

	//STEP 4 -> generate candidate operations
	candidateOperations := make(map[uuid.UUID]storage.Operation)
	for XT, _ := range XnameTargetHSMMap {
		hsmdata := XnameTargetHSMMap[XT]
		op := storage.NewOperation()
		op.ActionID = action.ActionID
		op.Target = XT.Target
		op.TargetName = XT.TargetName
		op.Xname = XT.Xname
		op.Model = hsmdata.Model
		op.Manufacturer = hsmdata.Manufacturer
		op.HsmData = hsmdata
		op.DeviceType = hsmdata.Type
		op.AutomaticallyGenerated = false
		candidateOperations[op.OperationID] = *op
	}

	//STEP 5 -> filter candidates by Image; set the EXPLICT ToImageID if applicable
	FilterImage(&candidateOperations, action.Parameters)

	//STEP 6a -> load all the FromFirmwareVerions! into the operation by looping through the deviceMap
	// I am intentionally getting this for ALL Targets b/c of the recursion needed for operations may need this data.
	// I think this is the lesser of two evils, to get a bit more data, that I may need, then to do a very expensive query MANY times!

	deviceMap, errlist := GetCurrentFirmwareVersionsFromHsmDataAndTargets(XnameTargetHSMMap)
	action.Errors = append(action.Errors, errlist...)
	err = (*GLOB.DSP).StoreAction(action)
	if err != nil {
		logrus.Error(err)
	}
	//6b -> get all images
	imageMap := GetImageMap()

	buildOperations := true
	for buildOperations == true {
		buildOperations = false //immediately set to false, so we dont just loop infinite
		//I
		for operationID, _ := range candidateOperations {
			operation := candidateOperations[operationID]
			if operation.State.Can("configure") {
				//set up a few times
				operation.StartTime.Scan(time.Now()) //The clock starts NOW, set the expiration time!
				if action.Command.TimeLimit_Seconds > 0 {
					operation.ExpirationTime.Scan(time.Now().Add(time.Duration(action.Command.TimeLimit_Seconds) * time.Second))
				}
				//STEP 6b -> load the FromFirmwareVersion -> so I can look up the FromFirmwareID
				SetFirmwareVersion(&operation, &deviceMap)
				//if candidateOperation
				//STEP 7 -> try to lookup the actual images IDs!
				FillInImageId(&operation, &imageMap, action.Parameters)

				//STEP 8 -> SetNoSolutionOperations!
				//  At this point, every candidate operation should have a ToImageID; if it doesnt then END IT!
				SetNoSolOp(&operation)

				//STEP 9 -> SetNoOperationOperations!
				SetNoOpOp(&operation, action.Command.OverwriteSameImage)

				//Not a NoSOl nor a NoOP
				if operation.State.Can("configure") { //it has been configured; it is now READY to be 'started'
					//TODO here is a good place to add in @DependencyManagment -> see develop @ tag: with-dependency to see what we had.
					operation.State.Event("configure")

				}

				err := (*GLOB.DSP).StoreOperation(operation)
				if err != nil {
					logrus.Error(err)
				}
			}
			candidateOperations[operationID] = operation
		}
	}

	// Clean up Error List - Only have one of each error string
	action.Errors = model.RemoveDuplicateStrings(action.Errors)
	//Start or Finish the Action!
	if len(candidateOperations) == 0 {
		action.EndTime.Scan(time.Now())
		action.State.Event("finish")
	} else {
		if action.State.Can("configure") { //if it cant start its because it got kicked out!
			action.State.Event("configure")
		}

		//Figure out if there are any sibling blockers (xname == xname)
		//store the operations and load the OperationIDs into the action
		xnameOps := make(map[string][]uuid.UUID)
		for k, v := range candidateOperations {
			if v.State.Is("configured") {
				if xnameOp, ok := xnameOps[v.Xname]; ok {
					//there is at least one other entry!
					lastOp := xnameOp[len(xnameOp)-1]

					v.BlockedBy = append(v.BlockedBy, lastOp)
					v.State.Event("block")
					v.StateHelper = "blocked by sibling"
					candidateOperations[k] = v

				} else { //else, we are the first one here, so create it
					xnameOps[v.Xname] = append(xnameOps[v.Xname], v.OperationID)
				}
			}

			//regardless of that state, save it to the action
			action.OperationIDs = append(action.OperationIDs, k)
			err := (*GLOB.DSP).StoreOperation(v)
			if err != nil {
				logrus.Error(err)
			}
		}
	}
	//Store the action
	(*GLOB.DSP).StoreAction(action)
}

func FillInImageId(operation *storage.Operation, imageMap *map[uuid.UUID]storage.Image, parameters storage.ActionParameters) (err error) {
	for _, image := range *imageMap {
		_, found := model.Find(image.Models, operation.Model)
		_, softwareIdFound := model.Find(image.SoftwareIds, operation.SoftwareId)
		// if a software id is found on the node and the image, but does not match, do not use image
		if (!softwareIdFound) && (len(image.SoftwareIds) > 0 && len(operation.SoftwareId) > 0) {
			continue
		}
		// If a software id is found and matches the image, use this image no need to check other fields
		// Otherwise Model, DeviceType, Target, and Manufacturer must be the same
		if (softwareIdFound && (image.Target == operation.Target || image.Target == operation.TargetName)) ||
			(found &&
				strings.EqualFold(image.DeviceType, operation.DeviceType) &&
				(image.Target == operation.Target || image.Target == operation.TargetName) &&
				strings.EqualFold(image.Manufacturer, operation.Manufacturer)) { //if the image could be on. or could be applied
			if image.FirmwareVersion == operation.FromFirmwareVersion { //We found the FROM IMAGE!!
				//TODO problem: The tag thing gets hard here... new rule: the firmware version must be unique for the devicetype/manf/model;
				operation.FromImageID = image.ImageID
			}
			if parameters.Command.Version != "explicit" && operation.AutomaticallyGenerated == false { //then try to figure out what to set it to!
				_, found := model.Find(image.Tags, parameters.Command.Tag) //This satisfies CASMHMS-3169
				if parameters.Command.Version == "latest" {                //check if there is something later!
					if found {
						if operation.ToImageID == uuid.Nil {
							operation.ToImageID = image.ImageID
						}
						if (*imageMap)[operation.ToImageID].SemanticFirmwareVersion.LessThan(image.SemanticFirmwareVersion) {
							operation.ToImageID = image.ImageID
						}
					}
				} else if parameters.Command.Version == "earliest" { //check if there is something earlier!
					if (operation.ToImageID == uuid.Nil ||
						(*imageMap)[operation.ToImageID].SemanticFirmwareVersion.GreaterThan(image.SemanticFirmwareVersion)) &&
						found { // you can ONLY get default image, unless you used an EXPLICIT image filter! {
						operation.ToImageID = image.ImageID
					}
				}
			}
		}
	}
	return nil
}

func FilterImage(candidateOperations *map[uuid.UUID]storage.Operation, parameters storage.ActionParameters) (err error) {
	//Filter on Image filter. Need to have all the operation data to see if the explicit image would fit from a Generic TYPE perspective
	logrus.WithFields(logrus.Fields{"Parameters": parameters}).Trace("IN FilterImage")
	if parameters.ImageFilter.ImageID != uuid.Nil && parameters.Command.Version == "explicit" { //start to apply the filter
		image, err := (*GLOB.DSP).GetImage(parameters.ImageFilter.ImageID)
		if err != nil {
			// an image should exist, but doesnt, then somehow we got a bad request through...
			//which means someone JUST deleted it, or something went wrong.  So dump the operations
			logrus.WithFields(logrus.Fields{"image NOT FOUND": image}).Error("Removed ALL candidate operation FilterImage")
			candidateOperations = nil
		} else { // the image exists
			logrus.WithFields(logrus.Fields{"image FOUND": image}).Trace("FilterImage")
			for k, v := range *candidateOperations {
				_, found := model.Find(image.Models, v.Model)

				if !parameters.ImageFilter.OverrideImage && // If overrideImage True, we update alsways
					(!found ||
						image.Target != v.Target ||
						!strings.EqualFold(image.Manufacturer, v.Manufacturer) ||
						!strings.EqualFold(image.DeviceType, v.DeviceType)) {
					delete(*candidateOperations, k)
					logrus.WithFields(logrus.Fields{"candidate operations": v, "image": image}).Trace("Removed candidate operation FilterImage")
				} else { //the image is a possibility, save it!
					logrus.WithFields(logrus.Fields{"candidate operations": v, "image": image}).Trace("Setting image FilterImage")
					v.ToImageID = image.ImageID
					(*candidateOperations)[k] = v
				}
			}
		}
	}
	return nil
}

func FilterModelManufacturer(dataMap *map[hsm.XnameTarget]hsm.HsmData, parameters storage.InventoryHardwareFilter) (err error) {
	if parameters.Empty() == false {
		for xnameTarget, hsmdata := range *dataMap {
			if parameters.Manufacturer != "" && hsmdata.Manufacturer != parameters.Manufacturer {
				delete(*dataMap, xnameTarget)
				logrus.WithFields(logrus.Fields{"xnameTarget": xnameTarget, "manufacturer filter": parameters.Manufacturer}).Trace("removing device as candidate; manufacturer is not equal. ")
				continue
			}
			if parameters.Model != "" && hsmdata.Model != parameters.Model {
				delete(*dataMap, xnameTarget)
				logrus.WithFields(logrus.Fields{"xnameTarget": xnameTarget, "Models filter": parameters.Model}).Trace("removing device as candidate; Models is not equal. ")
				continue
			}
		}
	}
	return
}

func FilterTargets(hsmDataMap *map[string]hsm.HsmData, parameters storage.TargetFilter) (XnameTargets []hsm.XnameTarget, MatchedXnameTargets []hsm.XnameTarget, UnMatchedXnameTargets []hsm.XnameTarget) {
	XnameTargets, errs := (*GLOB.HSM).GetTargetsRF(hsmDataMap)
	if len(errs) != 0 {
		logrus.Error(errs)
	}

	//Filter on Targets
	if len(parameters.Targets) > 0 {
		for _, XT := range XnameTargets {
			matched := false
			//Filter on Targets
			for _, target := range parameters.Targets {
				if (XT.Target == target) || (XT.TargetName == target) {
					MatchedXnameTargets = append(MatchedXnameTargets, XT)
					matched = true
				}
			}
			if !matched {
				UnMatchedXnameTargets = append(UnMatchedXnameTargets, XT)
			}
		}
	} else {
		//if there was no filter; than take everything
		MatchedXnameTargets = append(MatchedXnameTargets, XnameTargets...)
	}
	return
}

func SetNoSolOp(candidateOperation *storage.Operation) {
	if candidateOperation.ToImageID == uuid.Nil && candidateOperation.State.Can("nosol") {
		candidateOperation.State.Event("nosol")
		candidateOperation.EndTime.Scan(time.Now())
		candidateOperation.StateHelper = "No Image available"
	}
	return
}

func SetNoOpOp(candidateOperation *storage.Operation, overwriteSameImage bool) {
	if !overwriteSameImage {
		if candidateOperation.ToImageID == candidateOperation.FromImageID && candidateOperation.State.Can("noop") {
			candidateOperation.State.Event("noop")
			candidateOperation.EndTime.Scan(time.Now())
			candidateOperation.StateHelper = "Firmware at requested version"
		}
	}
	return
}

func GetImageMap() (images map[uuid.UUID]storage.Image) {
	imagelist, _ := (*GLOB.DSP).GetImages()
	images = make(map[uuid.UUID]storage.Image)
	for _, image := range imagelist {
		images[image.ImageID] = image
	}
	return
}

func SetFirmwareVersion(candidateOperation *storage.Operation, deviceMap *map[string]storage.Device) {
	if device, ok := (*deviceMap)[candidateOperation.Xname]; ok {
		for _, target := range device.Targets {
			if target.Name == candidateOperation.Target && target.FirmwareVersion != "" {
				candidateOperation.FromFirmwareVersion = target.FirmwareVersion
				candidateOperation.SoftwareId = target.SoftwareId
				candidateOperation.TargetName = target.TargetName
			}
		}
	}
}
