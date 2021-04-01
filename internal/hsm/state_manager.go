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
package hsm

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"strconv"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"stash.us.cray.com/HMS/hms-base"
	rf "stash.us.cray.com/HMS/hms-smd/pkg/redfish"
	"stash.us.cray.com/HMS/hms-smd/pkg/sm"
)

const crayModelRedfishPath = "/redfish/v1/Chassis/Enclosure"
const crayModelBIOS0RedfishPath = "/redfish/v1/Chassis/Enclosure"
const crayModelBIOS1RedfishPath = "/redfish/v1/Chassis/Enclosure"
const intelModelRedfishPath = "/redfish/v1/Chassis/RackMount"
const gigabyteModelRedfishPath = "/redfish/v1/Chassis/Self"
const hpeModelRedfishPath = "/redfish/v1/Chassis/1"
const manufacturerCray = "cray"
const manufacturerGigabyte = "gigabyte"
const manufacturerIntel = "intel"
const manufacturerHPE = "hpe"

const hsmRedfishEndpointsPath = "/hsm/v2/Inventory/RedfishEndpoints"
const hsmRedfishUpdateServicePath = "/hsm/v2/Inventory/ServiceEndpoints/UpdateService/RedfishEndpoints"
const hsmStateComponentsPath = "/hsm/v2/State/Components"
const hsmComponentEndpointsPath = "/hsm/v2/Inventory/ComponentEndpoints"
const hsmInventoryHardwarePath = "/hsm/v2/Inventory/Hardware"
const defaultSMSServer = "https://api-gw-service-nmn/apis/smd"

type RedfishModel struct {
	Model string `json:"Model"`
}

type XnameTarget struct {
	Xname  string
	Target string
	TargetName string
	Version string
}

// RefillModelRF -> will take a listing of xnameTargets / hsmdata  + a list of special targets/ rf paths and perform an
// operation to reset the hsmdata.model. It will use the rf path to query the device and pull out the model
func (b *HSMv0) RefillModelRF(XnameTargetHsmData *map[XnameTarget]HsmData, specialTargets map[string]string) (errs []error) {

	if len(specialTargets) == 0 {
		return
	}

	//taskMap remembers the UUID of the task, to the xnameTarget it belongs to.
	taskMap := make(map[uuid.UUID]XnameTarget)

	//targetMap is a lookup from xnameTarget to path (for the query)
	targetMap := make(map[XnameTarget]string)

	//For each xnametarget; if the target is int hte specialTargets list; then store the rfpath according to xnameTarget
	for xnameTarget, _ := range *XnameTargetHsmData {
		for target, rfpath := range specialTargets {
			if strings.EqualFold(xnameTarget.Target, target) {
				targetMap[xnameTarget] = rfpath
			}
		}
	}

	if len(targetMap) == 0 {
		return
	}

	taskList := (*b.HSMGlobals.RFTloc).CreateTaskList(b.HSMGlobals.BaseTRSTask, len(targetMap))

	//for every stored xnametarget rfpath; build a task; be sure to store the uuid of the task, for later retrieval
	counter := 0
	for xnameTarget, rfpath := range targetMap {

		hsmdata := (*XnameTargetHsmData)[xnameTarget]
		ID := taskList[counter].GetID()
		taskMap[ID] = xnameTarget
		taskList[counter].Request.URL, _ = url.Parse("https://" + path.Join(hsmdata.FQDN, rfpath))
		taskList[counter].Timeout = time.Second * 40
		taskList[counter].RetryPolicy.Retries = 3

		if !(hsmdata.User == "" && hsmdata.Password == "") {
			taskList[counter].Request.SetBasicAuth(hsmdata.User, hsmdata.Password)
		}
		counter++
	}

	(*b.HSMGlobals.RFClientLock).RLock()
	defer (*b.HSMGlobals.RFClientLock).RUnlock()
	rchan, err := (*b.HSMGlobals.RFTloc).Launch(&taskList)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		errs = append(errs, err)
	}

	for _, _ = range targetMap {
		tdone := <-rchan
		if *tdone.Err != nil {
			b.HSMGlobals.Logger.Error(*tdone.Err)
			errs = append(errs, *tdone.Err)
			continue
		}

		body, err := ioutil.ReadAll(tdone.Request.Response.Body)
		var data NodeInfo
		err = json.Unmarshal(body, &data)
		if err != nil {
			b.HSMGlobals.Logger.Error(err)
			errs = append(errs, err)
		} else {

			xnameTarget := taskMap[tdone.GetID()]
			hsmdata := (*XnameTargetHsmData)[xnameTarget]
			hsmdata.Model = data.Model
			(*XnameTargetHsmData)[xnameTarget] = hsmdata
		}
	}
	return
}

func (b *HSMv0) GetTargetsRF(hd *map[string]HsmData) (tuples []XnameTarget, errs []error) {

	taskMap := make(map[uuid.UUID]*HsmData)

	var HsmDataWithSetInventoryURI []*HsmData
	for xname, v := range *hd {
		if len(v.InventoryURI) > 0 {
			val := (*hd)[xname]
			HsmDataWithSetInventoryURI = append(HsmDataWithSetInventoryURI, &val)
		} else {
			b.HSMGlobals.Logger.WithFields(logrus.Fields{"xname": xname}).Warn("No InventoryURI available to query")
		}
	}

	if len(HsmDataWithSetInventoryURI) == 0 {
		return
	}

	taskList := (*b.HSMGlobals.RFTloc).CreateTaskList(b.HSMGlobals.BaseTRSTask, len(HsmDataWithSetInventoryURI))

	for l, data := range HsmDataWithSetInventoryURI {
		taskMap[taskList[l].GetID()] = data
		taskList[l].Request.URL, _ = url.Parse("https://" + path.Join(data.FQDN, data.InventoryURI))
		if data.Manufacturer == "hpe" {
		  taskList[l].Request.URL, _ = url.Parse("https://" + path.Join(data.FQDN, data.InventoryURI + "?$expand=."))
		}
		taskList[l].Timeout = time.Second * 40
		taskList[l].RetryPolicy.Retries = 3

		if !(data.User == "" && data.Password == "") {
			taskList[l].Request.SetBasicAuth(data.User, data.Password)
		}
	}

	(*b.HSMGlobals.RFClientLock).RLock()
	defer (*b.HSMGlobals.RFClientLock).RUnlock()
	rchan, err := (*b.HSMGlobals.RFTloc).Launch(&taskList)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		errs = append(errs, err)
	}

	for _, _ = range HsmDataWithSetInventoryURI {
		tdone := <-rchan
		if *tdone.Err != nil {
			b.HSMGlobals.Logger.Error(*tdone.Err)
			errs = append(errs, *tdone.Err)
			continue
		}

		body, err := ioutil.ReadAll(tdone.Request.Response.Body)
		var data TargetedMembers
		err = json.Unmarshal(body, &data)
		if err != nil {
			b.HSMGlobals.Logger.Error(err)
			errs = append(errs, err)
		} else {

			xhd := taskMap[tdone.GetID()]
			for k, _ := range data.InventoriedMembers {
				tuples = append(tuples, XnameTarget{
					Xname:  xhd.ID,
					Target: filepath.Base(data.InventoriedMembers[k].Path),
					TargetName: data.InventoriedMembers[k].TargetName,
					Version: data.InventoriedMembers[k].Version,
				})
			}
		}
	}
	return
}

func (b *HSMv0) FillUpdateServiceData(hd *map[string]HsmData) (errs []error) {

	taskList := (*b.HSMGlobals.SVCTloc).CreateTaskList(b.HSMGlobals.BaseTRSTask, len(*hd))

	taskMap := make(map[uuid.UUID]string)

	counter := 0
	for xname, datum := range *hd {
		taskMap[taskList[counter].GetID()] = xname
		taskList[counter].Request.URL, _ = url.Parse(b.HSMGlobals.StateManagerServer + hsmRedfishUpdateServicePath + "/" + datum.ID)
		taskList[counter].RetryPolicy.Retries = 3
		counter++
	}

	rchan, err := (*b.HSMGlobals.SVCTloc).Launch(&taskList)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		errs = append(errs, err)
	}

	for _, _ = range *hd {
		tdone := <-rchan
		xname := taskMap[tdone.GetID()]
		datum := (*hd)[xname]

		if *tdone.Err != nil {
			datum.Error = *tdone.Err
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(*tdone.Err)
			errs = append(errs, *tdone.Err)
			continue
		}

		if tdone.Request.Response.StatusCode < 200 && tdone.Request.Response.StatusCode >= 300 {
			datum.Error = errors.New("bad status code from UpdateService: " + strconv.Itoa(tdone.Request.Response.StatusCode))
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(datum.Error)
			errs = append(errs, datum.Error)
			continue
		}

		if tdone.Request.Response.Body == nil {
			datum.Error = errors.New("empty body")
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(datum.Error)
			errs = append(errs, datum.Error)
			continue
		}

		body, err := ioutil.ReadAll(tdone.Request.Response.Body)
		if err != nil {
			datum.Error = err
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(err)
			errs = append(errs, err)
			continue
		}

		var data UpdateService
		err = json.Unmarshal(body, &data)
		if err != nil {
			datum.Error = err
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(err)
			errs = append(errs, err)
			continue
		}
		datum.UpdateURI = data.ServiceInfo.Actions.Update.Path
		if len(data.ServiceInfo.FirmwareInventory.Path) > 0 {
			datum.InventoryURI = data.ServiceInfo.FirmwareInventory.Path
		} else if len(data.ServiceInfo.SoftwareInventory.Path) > 0 {
			datum.InventoryURI = data.ServiceInfo.SoftwareInventory.Path
		}
		(*hd)[xname] = datum
	}

	return
}

func (b *HSMv0) FillComponentEndpointData(hd *map[string]HsmData) (errs []error) {
	taskMap := make(map[uuid.UUID]string) //xname!
	taskList := (*b.HSMGlobals.SVCTloc).CreateTaskList(b.HSMGlobals.BaseTRSTask, len(*hd))
	counter := 0
	for xname, _ := range *hd {
		taskList[counter].Request.URL, _ = url.Parse(b.HSMGlobals.StateManagerServer + hsmComponentEndpointsPath + "/" + xname)
		taskList[counter].RetryPolicy.Retries = 3
		b.HSMGlobals.Logger.WithField("xname", xname).Trace(hsmComponentEndpointsPath)
		taskMap[taskList[counter].GetID()] = xname
		counter++
	}

	rchan, err := (*b.HSMGlobals.SVCTloc).Launch(&taskList)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		errs = append(errs, err)
	}

	for _, _ = range taskList {
		tdone := <-rchan
		xname := taskMap[tdone.GetID()]
		datum := (*hd)[xname]

		if *tdone.Err != nil {
			datum.Error = *tdone.Err
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(*tdone.Err)
			errs = append(errs, *tdone.Err)
			continue
		}
		b.HSMGlobals.Logger.Tracef("tdone: Get ComponentEndpoint data: %+v", tdone.Request.Response.StatusCode)
		if tdone.Request.Response.StatusCode != http.StatusOK {
			datum.Error = errors.New("bad status code from ComponentEndpoint data: " + strconv.Itoa(tdone.Request.Response.StatusCode) + " -- " + tdone.Request.URL.String())
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(datum.Error)
			errs = append(errs, datum.Error)
			continue
		}

		if tdone.Request.Response.Body == nil {
			datum.Error = errors.New("empty body")
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(datum.Error)
			errs = append(errs, datum.Error)
			continue
		}

		body, err := ioutil.ReadAll(tdone.Request.Response.Body)
		if err != nil {
			datum.Error = err
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(err)
			errs = append(errs, err)
			continue
		}

		var componentEndpoint sm.ComponentEndpoint
		err = json.Unmarshal(body, &componentEndpoint)
		if err != nil {
			datum.Error = err
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(err)
			errs = append(errs, err)
			continue
		}

		if componentEndpoint.RedfishChassisInfo != nil &&
			componentEndpoint.RedfishChassisInfo.Actions != nil {
			datum.ActionReset = componentEndpoint.RedfishChassisInfo.Actions.ChassisReset
		}

		if componentEndpoint.RedfishSystemInfo != nil &&
			componentEndpoint.RedfishSystemInfo.Actions != nil {
			datum.ActionReset = componentEndpoint.RedfishSystemInfo.Actions.ComputerSystemReset
		}

		if componentEndpoint.RedfishManagerInfo != nil &&
			componentEndpoint.RedfishManagerInfo.Actions != nil {
			datum.ActionReset = componentEndpoint.RedfishManagerInfo.Actions.ManagerReset
		}

		datum.BmcPath = componentEndpoint.OdataID //This logic if from CAPMC (hsmapi.go:~436)
		datum.RfType = componentEndpoint.RedfishType

		(*hd)[xname] = datum

	}
	return
}

func (b *HSMv0) GetStateComponents(xnames []string, partitions []string, groups []string, types []string) (data base.ComponentArray, err error) {

	var all bool
	all = true
	var xnameString string
	if len(xnames) > 0 {
		for _, v := range xnames {
			tmp := "&id=" + v
			xnameString += tmp
		}
		all = false
	}

	var parString string
	if len(partitions) > 0 {
		for _, v := range partitions {
			tmp := "&partition=" + v
			parString += tmp
		}
		all = false
	}

	var groupString string
	if len(groups) > 0 {
		for _, v := range groups {
			tmp := "&group=" + v
			groupString += tmp
		}
		all = false
	}

	var typeString string
	if len(types) > 0 {
		for _, v := range types {
			tmp := "&type=" + v
			typeString += tmp
		}
		all = false
	}

	baseURL, _ := url.Parse(b.HSMGlobals.StateManagerServer + hsmStateComponentsPath)
	finalURL := baseURL
	if all == false {
		queryString := xnameString + parString + groupString + typeString
		queryString = trimLeftChars(queryString, 1)
		finalURL, _ = url.Parse(baseURL.String() + "/?" + queryString)
		//finalURL, _ = url.Parse(strings.Replace(baseURL.String(), "/?&", "/?", 1))
	}

	b.HSMGlobals.Logger.WithField("url", finalURL.String()).Trace("preparing to get state componets")

	req, err := http.NewRequest("GET", finalURL.String(), nil)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		return
	}

	reqContext, _ := context.WithTimeout(context.Background(), time.Second*40)
	req = req.WithContext(reqContext)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		return
	}

	resp, err := b.HSMGlobals.SVCHttpClient.Do(req)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		return
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		return
	}
	logrus.Debug(data)
	return
}

func (b *HSMv0) FillRedfishEndpointData(hd *map[string]HsmData) (errs []error) {
	taskMap := make(map[uuid.UUID]string) //xname!
	taskList := (*b.HSMGlobals.SVCTloc).CreateTaskList(b.HSMGlobals.BaseTRSTask, len(*hd))
	counter := 0
	for xname, _ := range *hd {
		taskList[counter].Request.URL, _ = url.Parse(b.HSMGlobals.StateManagerServer + hsmRedfishEndpointsPath + "/" + xname)
		taskList[counter].RetryPolicy.Retries = 3
		taskMap[taskList[counter].GetID()] = xname
		counter++
	}

	rchan, err := (*b.HSMGlobals.SVCTloc).Launch(&taskList)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		errs = append(errs, err)
	}

	for _, _ = range taskList {
		tdone := <-rchan
		xname := taskMap[tdone.GetID()]
		datum := (*hd)[xname]

		if *tdone.Err != nil {
			datum.Error = *tdone.Err
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(*tdone.Err)
			errs = append(errs, *tdone.Err)
			continue
		}
		b.HSMGlobals.Logger.Tracef("tdone: GetHSMData: %+v", tdone.Request.Response)
		if tdone.Request.Response.StatusCode != http.StatusOK {
			datum.Error = errors.New("bad status code from Inventory/RedfishEndpoints: " + strconv.Itoa(tdone.Request.Response.StatusCode))
			//DELETE it from the listing, b/c if it doesnt have a RF endpoint, we cannot talk to it!
			delete(*hd, xname)
			b.HSMGlobals.Logger.Error(datum.Error)
			errs = append(errs, datum.Error)
			continue
		}

		if tdone.Request.Response.Body == nil {
			datum.Error = errors.New("empty body")
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(datum.Error)
			errs = append(errs, datum.Error)
			continue
		}

		body, err := ioutil.ReadAll(tdone.Request.Response.Body)
		if err != nil {
			datum.Error = err
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(err)
			errs = append(errs, err)
			continue
		}

		var data rf.RedfishEPDescription
		err = json.Unmarshal(body, &data)
		if err != nil {
			datum.Error = err
			(*hd)[xname] = datum
			b.HSMGlobals.Logger.Error(err)
			errs = append(errs, err)
			continue
		}

		tmpDatum := ToHsmDataFromRFEndpoint(&data)
		datum.CopyFrom(&tmpDatum)
		(*hd)[xname] = datum

	}
	return
}

func (b *HSMv0) Init(globals *HSM_GLOBALS) (err error) {
	b.HSMGlobals = HSM_GLOBALS{}
	b.HSMGlobals = *globals

	if globals.VaultEnabled {
		b.HSMGlobals.Credentials, err = setupVault(b)
		if err != nil {
			b.HSMGlobals.Logger.Error(err)
		}
	}

	if globals.LockEnabled {

		logy := logrus.New()
		logLevel := ""
		envstr := os.Getenv("SERVICE_RESERVATION_VERBOSITY")
		if envstr != "" {
			logLevel = strings.ToUpper(envstr)
		}

		switch logLevel {
		case "TRACE":
			logy.SetLevel(logrus.TraceLevel)
		case "DEBUG":
			logy.SetLevel(logrus.DebugLevel)
		case "INFO":
			logy.SetLevel(logrus.InfoLevel)
		case "WARN":
			logy.SetLevel(logrus.WarnLevel)
		case "ERROR":
			logy.SetLevel(logrus.ErrorLevel)
		case "FATAL":
			logy.SetLevel(logrus.FatalLevel)
		case "PANIC":
			logy.SetLevel(logrus.PanicLevel)
		default:
			logy.SetLevel(logrus.ErrorLevel)
		}

		Formatter := new(logrus.TextFormatter)
		Formatter.TimestampFormat = "2006-01-02T15:04:05.999999999Z07:00"
		Formatter.FullTimestamp = true
		Formatter.ForceColors = true
		logy.SetFormatter(Formatter)

		b.HSMGlobals.Reservation.Init(b.HSMGlobals.StateManagerServer, "", 1, logy)
	}

	return
}

func (b *HSMv0) Ping() (err error) {
	//_, err = b.GetStateComponents([]string{}, []string{}, []string{}, []string{})
	finalURL, _ := url.Parse(b.HSMGlobals.StateManagerServer + "/service/values/class")

	req, err := http.NewRequest("GET", finalURL.String(), nil)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		return
	}

	reqContext, _ := context.WithTimeout(context.Background(), time.Second*5)
	req = req.WithContext(reqContext)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		return
	}

	_, err = b.HSMGlobals.SVCHttpClient.Do(req)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		return
	}
	return
}

func (b *HSMv0) FillHSMData(xnames []string, partitions []string, groups []string, types []string) (hd map[string]HsmData, errs []error) {
	// Get xnames
	//   fill role
	//   get/fill endpoints
	//   fill credentials
	//   fill model/manf
	//   fill updateService uri

	hd = make(map[string]HsmData)
	// get StateCompnent
	filteredComponents, _ := b.GetStateComponents(xnames,
		partitions,
		groups,
		types)

	logrus.WithField("filteredComponents", filteredComponents).Trace("GET STATE COMPONENTS")

	//Fill Xname / Role
	for _, v := range filteredComponents.Components {
		tmpHSMData := HsmData{
			ID:   v.ID,
			Role: v.Role}
		hd[v.ID] = tmpHSMData

	}
	b.HSMGlobals.Logger.Trace(filteredComponents.Components)

	//Get MOST of the data
	errs = b.FillComponentEndpointData(&hd)
	if len(errs) > 0 {
		logrus.Error(errs)
	}

	//Get MOST of the data
	errs = b.FillRedfishEndpointData(&hd)
	if len(errs) > 0 {
		logrus.Error(errs)
	}

	//get credentials
	if b.HSMGlobals.VaultEnabled {
		// Lookup the credentials if we have Vault enabled.
		for k, v := range hd {
			err := updateHsmDataWithCredentials(b, &v)

			if err != nil {
				b.HSMGlobals.Logger.Error(err)
				errs = append(errs, err)
			} else {
				b.HSMGlobals.Logger.Debugf("Updated endpoint %s credentials with those retrieved from Vault", k)
				hd[k] = v
			}
		}
	}

	//get model/manufacturer
	errs = b.FillModelManufacturerRF(&hd)
	if len(errs) > 0 {
		logrus.Info("ANDREW")
		logrus.Error(errs)
	}

	//get updateService uri
	errs = b.FillUpdateServiceData(&hd)
	if len(errs) > 0 {
		logrus.Error(errs)
	}

	return
}

func (b *HSMv0) RestoreCredentials(hd *HsmData) (err error) {
	if b.HSMGlobals.VaultEnabled {
		// Lookup the credentials if we have Vault enabled.
		err := updateHsmDataWithCredentials(b, hd)

		if err != nil {
			b.HSMGlobals.Logger.Error(err)

		} else {
			b.HSMGlobals.Logger.Tracef("Restored endpoint %s credentials with those retrieved from Vault", hd.Hostname)
		}
	}
	return
}

func (b *HSMv0) FillModelManufacturerRF(hd *map[string]HsmData) (errs []error) {
	//get model/manufacturer

	URIs := []string{}
	URIs = append(URIs, crayModelRedfishPath)
	URIs = append(URIs, intelModelRedfishPath)
	URIs = append(URIs, gigabyteModelRedfishPath)
	URIs = append(URIs, hpeModelRedfishPath)
	//taskList = nil
	taskList := (*b.HSMGlobals.RFTloc).CreateTaskList(b.HSMGlobals.BaseTRSTask, len(*hd)*len(URIs))

	type XnameURI struct {
		Xname string
		URI   string
	}

	taskMap := make(map[uuid.UUID]XnameURI) //xname/URI!

	counter := 0
	for _, datum := range *hd {
		for _, uri := range URIs {
			tmpXnameURI := XnameURI{
				Xname: datum.ID,
				URI:   uri,
			}
			taskMap[taskList[counter].GetID()] = tmpXnameURI
			taskList[counter].Request.URL, _ = url.Parse("https://" + path.Join(datum.FQDN, uri))
			taskList[counter].Timeout = time.Second * 20
			taskList[counter].RetryPolicy.Retries = 3

			if !(datum.User == "" && datum.Password == "") {
				taskList[counter].Request.SetBasicAuth(datum.User, datum.Password)
			}
			counter++
		}
	}

	(*b.HSMGlobals.RFClientLock).RLock()
	defer (*b.HSMGlobals.RFClientLock).RUnlock()
	rchan, err := (*b.HSMGlobals.RFTloc).Launch(&taskList)
	if err != nil {
		b.HSMGlobals.Logger.Error(err)
		errs = append(errs, err)
	}

	for _, _ = range taskList {
		tdone := <-rchan
		if *tdone.Err != nil {
			b.HSMGlobals.Logger.Error(*tdone.Err)
			errs = append(errs, *tdone.Err)
			continue
		}
		tmpXnameURI := taskMap[tdone.GetID()]
		if tdone.Request.Response.StatusCode == http.StatusOK {
			//try to get the body
			if tdone.Request.Response.Body != nil {
				body, err := ioutil.ReadAll(tdone.Request.Response.Body)
				var device RedfishModel
				err = json.Unmarshal(body, &device)
				if err != nil {
					b.HSMGlobals.Logger.Error(err)
					errs = append(errs, err)
				} else {
					tmpHSMData := (*hd)[tmpXnameURI.Xname]

					if tmpXnameURI.URI == crayModelRedfishPath {
						tmpHSMData.Model = device.Model
						tmpHSMData.Manufacturer = manufacturerCray
					} else if tmpXnameURI.URI == intelModelRedfishPath {
						tmpHSMData.Model = device.Model
						tmpHSMData.Manufacturer = manufacturerIntel
					} else if tmpXnameURI.URI == gigabyteModelRedfishPath {
						tmpHSMData.Model = device.Model
						tmpHSMData.Manufacturer = manufacturerGigabyte
					} else if tmpXnameURI.URI == hpeModelRedfishPath {
						tmpHSMData.Model = device.Model
						tmpHSMData.Manufacturer = manufacturerHPE
					} else {
						tmpHSMData.Manufacturer = "unknown"
						tmpHSMData.Model = "unknown"
					}
					//flush it back
					b.HSMGlobals.Logger.Debugf("MODEL: %s -- MANUFACTURER: %s", tmpHSMData.Model, tmpHSMData.Manufacturer)
					(*hd)[tmpXnameURI.Xname] = tmpHSMData
				}
			}
		}
	}
	return
}

func trimLeftChars(s string, n int) string {
	m := 0
	for i := range s {
		if m >= n {
			return s[i:]
		}
		m++
	}
	return s[:0]
}

// We expect FAS to be reserving individual items.  If we get a list, we will
// reserve anything in the list. Already being reserved is an error, and reservations do not nest,
// so only call ClearLock call once!
func (b *HSMv0) SetLock(xnames []string) (error error) {
	if !b.HSMGlobals.LockEnabled {
		return nil
	}

	var aquireList []string
	//If we already have the lock we dont need to re-aquire it.
	for _, xname := range xnames {
		if b.HSMGlobals.Reservation.Check([]string{xname}) == false {
			aquireList = append(aquireList, xname)
		}
	}

	error = b.HSMGlobals.Reservation.Aquire(aquireList)
	return error
}

func (b *HSMv0) ClearLock(xnames []string) (error error) {
	if !b.HSMGlobals.LockEnabled {
		return nil
	}

	var clearList []string
	for _, xname := range xnames {
		if b.HSMGlobals.Reservation.Check([]string{xname}) == true {
			clearList = append(clearList, xname)
		}
	}
	error = b.HSMGlobals.Reservation.Release(clearList)
	return error
}
