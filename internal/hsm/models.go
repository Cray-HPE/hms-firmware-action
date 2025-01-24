/*
 * MIT License
 *
 * (C) Copyright [2020-2025] Hewlett Packard Enterprise Development LP
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
	"sync"

	"github.com/Cray-HPE/hms-certs/pkg/hms_certs"
	compcredentials "github.com/Cray-HPE/hms-compcredentials"
	rf "github.com/Cray-HPE/hms-smd/v2/pkg/redfish"
	reservation "github.com/Cray-HPE/hms-smd/v2/pkg/service-reservations"
	trsapi "github.com/Cray-HPE/hms-trs-app-api/v3/pkg/trs_http_api"
	"github.com/sirupsen/logrus"
)

type HsmData struct {
	ID           string           `json:"id,omitempty"`
	Type         string           `json:"type"`
	Hostname     string           `json:"hostname,omitempty"`
	Domain       string           `json:"domain,omitempty"`
	FQDN         string           `json:"FQDN"`
	Password     string           `json:"-"`
	User         string           `json:"-"`
	UpdateURI    string           `json:"updateURI"`
	InventoryURI string           `json:"inventoryURI"`
	Role         string           `json:"role"`
	Model        string           `json:"model"`
	Manufacturer string           `json:"manufacturer"`
	Error        error            `json:"error"`
	BmcPath      string           `json:"bmdPath"'`
	RfType       string           `json:"rfType"`
	DiscInfo     rf.DiscoveryInfo `json:"DiscoveryInfo,omitempty"`
	ActionReset  rf.ActionReset   `json:"actionReset,omitempty"`
}

func ToHsmDataFromRFEndpoint(src *rf.RedfishEPDescription) (dst HsmData) {
	dst = HsmData{
		ID:       src.ID,
		Type:     src.Type,
		Hostname: src.Hostname,
		Domain:   src.Domain,
		FQDN:     src.FQDN,
		Password: src.Password,
		User:     src.User,
		DiscInfo: src.DiscInfo,
	}
	return
}

func (src *HsmData) CopyFrom(ref *HsmData) {
	if ref.ID != "" {
		src.ID = ref.ID
	}
	if ref.Type != "" {
		src.Type = ref.Type
	}
	if ref.Hostname != "" {
		src.Hostname = ref.Hostname
	}
	if ref.Domain != "" {
		src.Domain = ref.Domain
	}
	if ref.FQDN != "" {
		src.FQDN = ref.FQDN
	}
	if ref.Password != "" {
		src.Password = ref.Password
	}
	if ref.User != "" {
		src.User = ref.User
	}
	if ref.UpdateURI != "" {
		src.UpdateURI = ref.UpdateURI
	}
	if ref.InventoryURI != "" {
		src.InventoryURI = ref.InventoryURI
	}
	if ref.Role != "" {
		src.Role = ref.Role
	}
	if ref.Model != "" {
		src.Model = ref.Model
	}
	if ref.Manufacturer != "" {
		src.Manufacturer = ref.Manufacturer
	}
	if ref.Error != nil {
		src.Error = ref.Error
	}
	src.DiscInfo = ref.DiscInfo
	src.ActionReset = ref.ActionReset

}

func (obj *HsmData) Equals(other HsmData) bool {
	if obj.ID != other.ID ||
		obj.Type != other.Type ||
		obj.Domain != other.Domain ||
		obj.Hostname != other.Hostname ||
		obj.FQDN != other.FQDN ||
		obj.Password != other.Password ||
		obj.UpdateURI != other.UpdateURI ||
		obj.InventoryURI != other.InventoryURI ||
		obj.Role != other.Role ||
		obj.Model != other.Model ||
		obj.Manufacturer != other.Manufacturer ||
		obj.Error != other.Error ||
		DiscoveryInfoEquals(obj.DiscInfo, other.DiscInfo) == false {
		return false
	}

	return true
}

func DiscoveryInfoEquals(obj rf.DiscoveryInfo, other rf.DiscoveryInfo) bool {
	if obj.LastAttempt != other.LastAttempt ||
		obj.LastStatus != other.LastStatus ||
		obj.RedfishVersion != other.RedfishVersion {
		return false
	}
	return true
}

type RedfishEndpoints struct {
	RedfishEndpoints []HsmData `json:"RedfishEndpoints"`
}

type UpdateService struct {
	ServiceInfo struct {
		FirmwareInventory struct {
			Path string `json:"@odata.id"`
		} `json:"FirmwareInventory"`
		SoftwareInventory struct {
			Path string `json:"@odata.id"`
		} `json:"SoftwareInventory"`
		Actions struct {
			Update struct {
				Path string `json:"target"`
			} `json:"#UpdateService.SimpleUpdate"`
		}
	}
}
type RedfishEndpointIDs struct {
	RedfishEndpoints []struct {
		Id string `json:"id"`
	}
}

type TargetedMembers struct {
	InventoriedMembers []struct {
		Path       string `json:"@odata.id"`
		TargetId   string `json:"Id"`
		TargetName string `json:"Name"`
		Version    string `json:"Version"`
	} `json:"Members"`
}

type NodeInfo struct {
	Model string `json:"Model"`
}

type StateComponents struct {
	Role string `json:"Role"`
}

type InventoryHardware struct {
	PopulatedFRU struct {
		NodeFRUInfo struct {
			Model string `json:"Models"`
		}
	}
}

type HSM_GLOBALS struct {
	Logger             *logrus.Logger
	BaseTRSTask        *trsapi.HttpTask
	RFTloc             *trsapi.TrsAPI
	SVCTloc            *trsapi.TrsAPI
	RFClientLock       *sync.RWMutex
	StateManagerServer string
	VaultEnabled       bool
	VaultKeypath       string
	Credentials        *compcredentials.CompCredStore
	Running            *bool
	LockEnabled        bool
	RFHttpClient       *hms_certs.HTTPClientPair
	SVCHttpClient      *hms_certs.HTTPClientPair
	Reservation        reservation.Production
}

func (g *HSM_GLOBALS) NewGlobals(Logger *logrus.Logger, baseTRSRequest *trsapi.HttpTask,
	tlocRF *trsapi.TrsAPI, tlocSVC *trsapi.TrsAPI,
	clientRF *hms_certs.HTTPClientPair,
	clientSVC *hms_certs.HTTPClientPair,
	rfClientLock *sync.RWMutex,
	sms string, vault bool, keypath string, running *bool,
	lockEnabled bool) {
	g.Logger = Logger
	g.BaseTRSTask = baseTRSRequest
	g.RFTloc = tlocRF
	g.SVCTloc = tlocSVC
	g.RFClientLock = rfClientLock
	g.RFHttpClient = clientRF
	g.SVCHttpClient = clientSVC
	g.StateManagerServer = sms
	g.VaultKeypath = keypath
	g.VaultEnabled = vault
	g.Running = running
	g.LockEnabled = lockEnabled
}
