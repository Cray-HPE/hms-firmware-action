// Copyright (c) 2019 Cray Inc. All Rights Reserved.
package sm

import (
	"strings"
)

type SMPatchOp int

const (
	PatchOpInvalid SMPatchOp = 0
	PatchOpAdd     SMPatchOp = 1
	PatchOpRemove  SMPatchOp = 2
	PatchOpReplace SMPatchOp = 3
)

var smPatchOpMap = map[string]SMPatchOp{
	"add":     PatchOpAdd,
	"remove":  PatchOpRemove,
	"replace": PatchOpReplace,
}

type SCNPostSubscription struct {
	Subscriber     string   `json:"Subscriber"`
	Enabled        *bool    `json:"Enabled,omitempty"`
	Roles          []string `json:"Roles,omitempty"`
	SubRoles       []string `json:"SubRoles,omitempty"`
	SoftwareStatus []string `json:"SoftwareStatus,omitempty"`
	States         []string `json:"States,omitempty"`
	Url            string   `json:"Url"`
}

type SCNSubscription struct {
	ID             int64    `json:"ID"`
	Subscriber     string   `json:"Subscriber"`
	Enabled        *bool    `json:"Enabled,omitempty"`
	Roles          []string `json:"Roles,omitempty"`
	SubRoles       []string `json:"SubRoles,omitempty"`
	SoftwareStatus []string `json:"SoftwareStatus,omitempty"`
	States         []string `json:"States,omitempty"`
	Url            string   `json:"Url"`
}

type SCNPatchSubscription struct {
	Op             string   `json:"Op"`
	Enabled        *bool    `json:"Enabled,omitempty"`
	Roles          []string `json:"Roles,omitempty"`
	SubRoles       []string `json:"SubRoles,omitempty"`
	SoftwareStatus []string `json:"SoftwareStatus,omitempty"`
	States         []string `json:"States,omitempty"`
}

type SCNSubscriptionArray struct {
	SubscriptionList []SCNSubscription `json:"SubscriptionList"`
}

type SCNPayload struct {
	Components     []string `json:"Components"`
	Enabled        *bool    `json:"Enabled,omitempty"`
	Flag           string   `json:"Flag,omitempty"`
	Role           string   `json:"Role,omitempty"`
	SubRole        string   `json:"SubRole,omitempty"`
	SoftwareStatus string   `json:"SoftwareStatus,omitempty"`
	State          string   `json:"State,omitempty"`
}

func GetPatchOp(op string) SMPatchOp {
	opInt, ok := smPatchOpMap[strings.ToLower(op)]
	if !ok {
		return PatchOpInvalid
	}
	return opInt
}
