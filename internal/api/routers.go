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

package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Route - struct containing name,method, pattern and handlerFunction to invoke.
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes - a collection of Route
type Routes []Route

// Logger - used for logging what methods were invoked and how long they took to complete
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		if name != "servicestatusAPI" {
			logrus.Printf(
				"%s %s %s %s",
				r.Method,
				r.RequestURI,
				name,
				time.Since(start),
			)
		}
	})
}

// NewRouter - create a new mux Router; and initializes it with the routes
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

var routes = Routes{

	Route{
		"Index",
		"GET",
		"/",
		Index,
	},
	// ACTIONS
	// GET actions
	Route{
		"GetActions",
		strings.ToUpper("get"),
		"/actions",
		GetAction,
	},
	// POST actions
	Route{
		"CreateAction",
		strings.ToUpper("post"),
		"/actions",
		CreateAction,
	},
	// GET actions/{actionID}
	Route{
		"GetActionID",
		strings.ToUpper("get"),
		"/actions/{actionID}",
		GetAction,
	},
	// DELETE actions/{actionID}
	Route{
		"DeleteActionID",
		strings.ToUpper("delete"),
		"/actions/{actionID}",
		DeleteAction,
	},
	// DELETE actions/{actionID}/instance
	Route{
		"AbortActionID",
		strings.ToUpper("delete"),
		"/actions/{actionID}/instance",
		AbortActionID,
	},
	// GET actions/{actionID}/operations/{operationsID}
	Route{
		"GetActionOperationID",
		strings.ToUpper("get"),
		"/actions/{actionID}/operations/{operationID}",
		GetActionOperationID,
	},
	// GET operations/{operationsID}
	Route{
		"GetActionOperationID",
		strings.ToUpper("get"),
		"/operations/{operationID}",
		GetOperationID,
	},
	// GET actions/{actionID}/operations
	Route{
		"GetActionIDOperations",
		strings.ToUpper("get"),
		"/actions/{actionID}/operations",
		GetActionIDOperations,
	},
	// GET actions/{actionID}/status
	Route{
		"GetActionIDStatus",
		strings.ToUpper("get"),
		"/actions/{actionID}/status",
		GetActionIDStatus,
	},

	Route{
		"ServiceStatusVersion",
		strings.ToUpper("get"),
		"/service/version",
		ServiceStatusVersion,
	},
	Route{
		"ServiceStatus",
		strings.ToUpper("get"),
		"/service/status",
		ServiceStatus,
	},
	Route{
		"ServiceStatusDetails",
		strings.ToUpper("get"),
		"/service/status/details",
		ServiceStatusDetails,
	},

	Route{
		"getImages",
		strings.ToUpper("get"),
		"/images",
		GetImages,
	},
	Route{
		"UpdateImage",
		strings.ToUpper("put"),
		"/images/{imageID}",
		UpdateImage,
	},
	Route{
		"CreateImage",
		strings.ToUpper("post"),
		"/images",
		CreateImage,
	},
	Route{
		"GetImage",
		strings.ToUpper("get"),
		"/images/{imageID}",
		GetImage,
	},
	Route{
		"DeleteImage",
		strings.ToUpper("delete"),
		"/images/{imageID}",
		DeleteImage,
	},
	Route{
		"getSnapshots",
		strings.ToUpper("get"),
		"/snapshots",
		GetSnapshots,
	},
	Route{
		"getSnapshot",
		strings.ToUpper("get"),
		"/snapshots/{name}",
		GetSnapshot,
	},
	Route{
		"createSnapshot",
		strings.ToUpper("post"),
		"/snapshots",
		CreateSnapshot,
	},
	Route{
		"startRestoreSnapshot",
		strings.ToUpper("post"),
		"/snapshots/{name}/restore",
		StartRestoreSnapshot,
	},
	Route{
		"deleteSnapshot",
		strings.ToUpper("delete"),
		"/snapshots/{name}",
		DeleteSnapshot,
	},
	Route{
		"loaderStatus",
		strings.ToUpper("get"),
		"/loader",
		LoaderStatus,
	},
	Route{
		"loaderStatusID",
		strings.ToUpper("get"),
		"/loader/{loaderID}",
		LoaderStatusID,
	},
	Route{
		"loaderDelete",
		strings.ToUpper("delete"),
		"/loader/{loaderID}",
		LoaderDeleteID,
	},
	Route{
		"loaderLoad",
		strings.ToUpper("post"),
		"/loader",
		LoaderLoad,
	},
	Route{
		"loaderLoadNexus",
		strings.ToUpper("post"),
		"/loader/nexus",
		LoaderLoadNexus,
	},
}
