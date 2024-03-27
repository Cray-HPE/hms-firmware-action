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

package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
	"github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-firmware-action/internal/api"
	"github.com/Cray-HPE/hms-firmware-action/internal/domain"
	"github.com/Cray-HPE/hms-firmware-action/internal/hsm"
	"github.com/Cray-HPE/hms-firmware-action/internal/logger"
	"github.com/Cray-HPE/hms-firmware-action/internal/storage"
	trsapi "github.com/Cray-HPE/hms-trs-app-api/pkg/trs_http_api"
	"github.com/Cray-HPE/hms-certs/pkg/hms_certs"
)

// Default Port to use
const defaultPORT = ":28800"

// Redfish path to use
const redfishPath = "/redfish/v1/UpdateService"

const redfishSimpleUpdate = redfishPath + "/Actions/SimpleUpdate"
const redfishFirmwareInventory = redfishPath + "/FirmwareInventory"
const redfishSoftwareInventory = redfishPath + "/SoftwareInventory"

const defaultSMSServer = "https://api-gw-service-nmn.local/apis/smd"
const defaultNodeBlacklist = "ignore_ignore_ignore"

const manufacturerCray = "cray"
const manufacturerGigabyte = "gigabyte"
const manufacturerIntel = "intel"
const manufacturerHPE = "hpe"
const manufacturerFoxconn = "foxconn"

const (
	dfltMaxHTTPRetries = 5
	dfltMaxHTTPTimeout = 40
	dfltMaxHTTPBackoff = 8
)
const defaultS3Endpoint = "s3"
const defaultTFTPEndpoint = "TFTP"

var S3_ENDPOINT string
var TFTP_ENDPOINT string

var nodeBlacklistSt string
var nodeBlacklist []string

var FileCheckClient *retryablehttp.Client

var Running = true
var DSP storage.StorageProvider
var HSM hsm.HSMProvider

var restSrv *http.Server = nil
var waitGroup sync.WaitGroup
var mainLogger *logrus.Logger

var rfClient, svcClient *hms_certs.HTTPClientPair
var TLOC_rf, TLOC_svc trsapi.TrsAPI
var caURI string
var rfClientLock *sync.RWMutex = &sync.RWMutex{}
var serviceName string

func main() {

	//Setup logging
	mainLogger = logger.Init()

	var VaultEnabled bool
	var VaultKeypath string
	var StateManagerServer string
	var hsmlockEnabled bool = true
	var runControl bool = true
	var err error
	var DaysToKeepActions int
	srv := &http.Server{Addr: defaultPORT}

	///////////////////////////////
	//ENVIRONMENT PARSING
	//////////////////////////////

	flag.StringVar(&StateManagerServer, "sms_server", defaultSMSServer, "SMS Server")
	flag.StringVar(&nodeBlacklistSt, "node_blacklist", defaultNodeBlacklist, "Node Black List")
	flag.StringVar(&S3_ENDPOINT, "s3_endpoint", defaultS3Endpoint, "S3 Endpoint")
	flag.StringVar(&TFTP_ENDPOINT, "tftp_endpoint", defaultTFTPEndpoint, "TFTP Endpoint")
	flag.BoolVar(&runControl, "run_control", runControl, "run control loop; false runs API only")
	flag.BoolVar(&hsmlockEnabled, "hsmlock_enabled", true, "Use HSM Locking")
	flag.BoolVar(&VaultEnabled, "vault_enabled", true, "Should vault be used for credentials?")
	flag.StringVar(&VaultKeypath, "vault_keypath", "secret/hms-creds",
		"Keypath for Vault credentials.")
	flag.IntVar(&DaysToKeepActions, "days_to_keep_actions", 0, "Days to Keep Actions before deleting")

	flag.Parse()

	serviceName, err := base.GetServiceInstanceName()
	if err != nil {
		serviceName = "FAS"
		mainLogger.Info("WARNING: could not get service/instance name, using: " + serviceName)
	}
	mainLogger.Info("Service/Instance name: " + serviceName)

	nodeBlacklist = append(nodeBlacklist, strings.Split(nodeBlacklistSt, ",")...)

	mainLogger.Info("SMS Server: " + StateManagerServer)
	mainLogger.Info("Node Black List: ", nodeBlacklist)
	mainLogger.Info("HSM Lock Enabled: ", hsmlockEnabled)
	mainLogger.Info("Vault Enabled: ", VaultEnabled)
	mainLogger.Info("Days To Keep Actions: ", DaysToKeepActions)
	mainLogger.SetReportCaller(true)

	///////////////////////////////
	//CONFIGURATION
	//////////////////////////////

	//Create a new Http Client (for non trs stuff)
	// For performance reasons we'll keep the client that was created for this base request and reuse it later.
	FileCheckClient = retryablehttp.NewClient()
	FileCheckClient.Logger = nil
	fc_transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	FileCheckClient.HTTPClient.Transport = fc_transport
	FileCheckClient.HTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	var BaseTRSTask trsapi.HttpTask
	BaseTRSTask.ServiceName = serviceName
	BaseTRSTask.Timeout = 40 * time.Second
	BaseTRSTask.Request, _ = http.NewRequest("GET", "", nil)
	BaseTRSTask.Request.Header.Set("Content-Type", "application/json")
	BaseTRSTask.Request.Header.Add("HMS-Service", BaseTRSTask.ServiceName)

	//INITIALIZE TRS

	logy := logrus.New()
	logy.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logy.SetLevel(logrus.InfoLevel)
	logy.SetReportCaller(true)

	var envstr string
	envstr = os.Getenv("TRS_IMPLEMENTATION")

	if envstr == "REMOTE" {
		workerSec := &trsapi.TRSHTTPRemote{}
		workerSec.Logger = logy
		workerInsec := &trsapi.TRSHTTPRemote{}
		workerInsec.Logger = logy
		TLOC_rf = workerSec
		TLOC_svc = workerInsec
	} else {
		workerSec := &trsapi.TRSHTTPLocal{}
		workerSec.Logger = logy
		workerInsec := &trsapi.TRSHTTPLocal{}
		workerInsec.Logger = logy
		TLOC_rf = workerSec
		TLOC_svc = workerInsec
	}

	//Set the kafka level to the same level.
	logy.SetLevel(mainLogger.GetLevel())

	//Set up TRS TLOCs and HTTP clients, all insecure to start with

	envstr = os.Getenv("FAS_CA_URI")
	if envstr != "" {
		caURI = envstr
	}
	//These are for debugging/testing
	envstr = os.Getenv("FAS_CA_PKI_URL")
	if envstr != "" {
		logrus.Printf("INFO: Using CA PKI URL: '%s'", envstr)
		hms_certs.ConfigParams.VaultCAUrl = envstr
	}
	envstr = os.Getenv("FAS_VAULT_PKI_URL")
	if envstr != "" {
		logrus.Printf("INFO: Using VAULT PKI URL: '%s'", envstr)
		hms_certs.ConfigParams.VaultPKIUrl = envstr
	}
	envstr = os.Getenv("FAS_VAULT_JWT_FILE")
	if envstr != "" {
		logrus.Printf("INFO: Using Vault JWT file: '%s'", envstr)
		hms_certs.ConfigParams.VaultJWTFile = envstr
	}
	envstr = os.Getenv("FAS_LOG_INSECURE_FAILOVER")
	if envstr != "" {
		yn, _ := strconv.ParseBool(envstr)
		if yn == false {
			logrus.Printf("INFO: Not logging Redfish insecure failovers.")
			hms_certs.ConfigParams.LogInsecureFailover = false
		}
	}

	TLOC_rf.Init(serviceName, logy)
	TLOC_svc.Init(serviceName, logy)
	rfClient, _ = hms_certs.CreateRetryableHTTPClientPair("", dfltMaxHTTPTimeout, dfltMaxHTTPRetries, dfltMaxHTTPBackoff)
	svcClient, _ = hms_certs.CreateRetryableHTTPClientPair("", dfltMaxHTTPTimeout, dfltMaxHTTPRetries, dfltMaxHTTPBackoff)

	////STORAGE CONFIGURATION
	envstr = os.Getenv("STORAGE")
	if envstr == "" || envstr == "MEMORY" {
		tmpStorageImplementation := &storage.MemStorage{
			Logger: logy,
		}
		DSP = tmpStorageImplementation
		mainLogger.Info("Storage Provider: In Memory")
	} else if envstr == "ETCD" {
		tmpStorageImplementation := &storage.ETCDStorage{
			Logger: logy,
		}
		DSP = tmpStorageImplementation
		mainLogger.Info("Storage Provider: ETCD")
	}
	DSP.Init(logy)

	//Hardware State Manager CONFIGURATION
	tmpHSM := &hsm.HSMv0{} //TODO this can more to config section

	HSM = tmpHSM

	var hsmGlob hsm.HSM_GLOBALS
	hsmGlob.NewGlobals(logy, &BaseTRSTask, &TLOC_rf, &TLOC_svc, rfClient,
		svcClient, rfClientLock, StateManagerServer, VaultEnabled,
		VaultKeypath, &Running, hsmlockEnabled)
	HSM.Init(&hsmGlob)

	//DOMAIN CONFIGURATION
	var domainGlobals domain.DOMAIN_GLOBALS
	domainGlobals.NewGlobals(&BaseTRSTask, &TLOC_rf, &TLOC_svc, rfClient,
		svcClient, rfClientLock, &Running, &DSP, &HSM, DaysToKeepActions)

	//Wait for vault PKI to respond for CA bundle.  Once this happens, re-do
	//the globals.  This goroutine will run forever checking if the CA trust
	//bundle has changed -- if it has, it will reload it and re-do the globals.

	//Set a flag "CA not ready" that the /liveness and /readiness APIs will
	//use to signify that FAS is not ready based on the transport readiness.

	go func() {
		if caURI != "" {
			var err error
			var caChain string
			var prevCaChain string
			RFTransportReady := false

			tdelay := time.Duration(0)
			for {
				time.Sleep(tdelay)
				tdelay = 3 * time.Second

				caChain, err = hms_certs.FetchCAChain(caURI)
				if err != nil {
					logrus.Errorf("Error fetching CA chain from Vault PKI: %v, retrying...",
						err)
					continue
				} else {
					logrus.Printf("CA trust chain loaded.")
				}

				//If chain hasn't changed, do nothing, expand retry time.

				if caChain == prevCaChain {
					tdelay = 10 * time.Second
					continue
				}

				//CA chain accessible.  Re-do the verified transports

				logrus.Infof("CA trust chain has changed, re-doing Redfish HTTP transports.")
				rfClient, err = hms_certs.CreateRetryableHTTPClientPair(caURI, dfltMaxHTTPTimeout, dfltMaxHTTPRetries, dfltMaxHTTPBackoff)
				if err != nil {
					logrus.Errorf("Error creating TLS-verified transport: %v, retrying...",
						err)
					continue
				}
				logrus.Infof("Locking RF operations...")
				rfClientLock.Lock() //waits for all RW locks to release
				tchain := hms_certs.NewlineToTuple(caChain)
				secInfo := trsapi.TRSHTTPLocalSecurity{CACertBundleData: tchain}
				err = TLOC_rf.SetSecurity(secInfo)
				if err != nil {
					logrus.Errorf("Error setting TLOC security info: %v, retrying...",
						err)
					rfClientLock.Unlock()
					continue
				} else {
					logrus.Printf("TRS CA security updated.")
				}
				prevCaChain = caChain

				//update RF tloc and rfclient to the global areas!
				domainGlobals.RFTloc = &TLOC_rf
				domainGlobals.RFHttpClient = rfClient
				hsmGlob.RFTloc = &TLOC_rf
				hsmGlob.RFHttpClient = rfClient
				HSM.Init(&hsmGlob)
				rfClientLock.Unlock()
				RFTransportReady = true
				domainGlobals.RFTransportReady = &RFTransportReady
			}
		}
	}()

	///////////////////////////////
	//INITIALIZATION
	//////////////////////////////
	domain.Init(&domainGlobals)

	///////////////////////////////
	//SIGNAL HANDLING -- //TODO does this need to move up ^ so it happens sooner?
	//////////////////////////////

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	idleConnsClosed := make(chan struct{})
	go func() {
		<-c
		Running = false

		//TODO; cannot Cancel the context on retryablehttp; because I havent set them up!
		//cancel()

		// Gracefully shutdown the HTTP server.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			mainLogger.Infof("HTTP server Shutdown: %v", err)
		}

		ctx := context.Background()
		if restSrv != nil {
			if err := restSrv.Shutdown(ctx); err != nil {
				logrus.Panic("ERROR: Unable to stop REST collection server!")
			}
		}

		close(idleConnsClosed)
	}()

	///////////////////////
	// START
	///////////////////////
	// TODO: Have a way to load the database before starting
	//Master Control
	if runControl {
		mainLogger.Info("Starting control loop")
		go controlLoop(&domainGlobals)
		envstr = os.Getenv("LOAD_NEXUS_WAIT_MIN")
		if envstr != "" {
			waitTime, err := strconv.Atoi(envstr)
			if err == nil {
				mainLogger.Info("Starting Do Load From Nexus, wait time: ", waitTime)
				go domain.DoLoadFromNexus(waitTime)
			} else {
				mainLogger.Error("Could not convert Nexus Wait Time: ", envstr)
			}
		} else {
			mainLogger.Info("Not running Do Load From Nexus")
		}
	} else {
		mainLogger.Info("NOT starting control loop")
	}
	//Rest Server
	waitGroup.Add(1)
	doRest("28800")

	//////////////////////
	// WAIT FOR GOD
	/////////////////////

	waitGroup.Wait()
	mainLogger.Info("HTTP server shutdown, waiting for idle connection to close...")
	<-idleConnsClosed
	mainLogger.Info("Done. Exiting.")
}

func doRest(serverPort string) {

	mainLogger.Info("**RUNNING -- Listening on " + defaultPORT)

	srv := &http.Server{Addr: ":" + serverPort}
	router := api.NewRouter()

	http.Handle("/", router)

	go func() {
		defer waitGroup.Done()
		if err := srv.ListenAndServe(); err != nil {
			// Cannot panic because this is probably just a graceful shutdown.
			mainLogger.Error(err)
			mainLogger.Info("REST collection server shutdown.")
		}
	}()

	mainLogger.Info("REST collection server started on port " + serverPort)
	restSrv = srv
}
