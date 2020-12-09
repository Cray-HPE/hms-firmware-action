// Copyright (c) 2018-2019 Cray Inc. All Rights Reserved.
package sm

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	base "stash.us.cray.com/HMS/hms-base"
)

////////////////////////////////////////////////////////////////////////////
//
// Error and debug logging.  For now just send to stdout.
//
////////////////////////////////////////////////////////////////////////////

var rfDebug int = 0
var rfVerbose int = 0
var errlog *log.Logger = log.New(os.Stdout, "", log.LstdFlags)

func SetDebug(level int) {
	rfDebug = level
}

func SetVerbose(level int) {
	rfVerbose = level
}

func SetLogger(l *log.Logger) {
	errlog = l
}

////////////////////////////////////////////////////////////////////////////
//
// Discovery
//
////////////////////////////////////////////////////////////////////////////

// Valid values for the DiscoveryStatus Status field below.
const (
	DiscNotStarted = "NotStarted"
	DiscPending    = "Pending"
	DiscComplete   = "Complete"
	DiscInProgress = "InProgress"
)

// Returns info on the current status of discovery for id (just 0 for now)
type DiscoveryStatus struct {
	ID         uint             `json:"ID"`
	Status     string           `json:"Status"`
	LastUpdate string           `json:"LastUpdateTime"`
	Details    *json.RawMessage `json:"Details,omitempty"`
}

// POST object to kick of discovery
type DiscoverIn struct {
	XNames []string `json:"xnames"`
	Force  bool     `json:"force"`
}

////////////////////////////////////////////////////////////////////////////
//
// Job Sync
//
////////////////////////////////////////////////////////////////////////////

// Valid values for the DiscoveryStatus Status field below.
const (
	JobNotStarted = "NotStarted"
	JobPending    = "Pending"
	JobComplete   = "Complete"
	JobInProgress = "InProgress"
	JobError      = "Error"
)

const (
	JobTypeSRFP = "StateRFPoll"
)

type JobData struct {
	Id         string
	Type       string
	Status     string
	LastUpdate string
	Lifetime   int
	KeepAlive  int
}

type Job struct {
	JobData
	Data interface{}
}

type SrfpJobData struct {
	CompId string
	Delay  int
	Poll   int
}

func NewStateRFPollJob(xname string, delay, poll, lifetime, keepAlive int) (*Job, error) {
	job := new(Job)
	job.Type = JobTypeSRFP
	job.Status = JobNotStarted
	job.Lifetime = lifetime
	job.KeepAlive = keepAlive

	// SRFP Job specific fields
	data := new(SrfpJobData)
	data.CompId = base.VerifyNormalizeCompID(xname)
	data.Delay = delay
	data.Poll = poll
	job.Data = data

	// Set minimum for keepAlive
	if keepAlive < 5 {
		job.KeepAlive = 5
	}

	// Keep lifetime at least 5 seconds longer than keep alive so we don't
	// accidentally expire.
	if lifetime-5 < job.KeepAlive {
		// Keep lifetime at least 5 seconds longer than keep alive so we don't
		// accidentally expire.
		job.Lifetime = job.KeepAlive + 5
	}

	// Validate the xname
	if len(data.CompId) == 0 {
		return nil, fmt.Errorf("xname ID '%s' is invalid", xname)
	}

	// At minimum 1 second delay before starting to poll
	if delay < 1 {
		data.Delay = 1
	}

	// Set minimum for polling so we don't over stress things
	if poll < 5 {
		data.Poll = 5
	}

	return job, nil
}
