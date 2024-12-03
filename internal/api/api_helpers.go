/*
 * MIT License
 *
 * (C) Copyright [2020-2021,2024] Hewlett Packard Enterprise Development LP
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
	"encoding/json"
	"io"
	"net/http"

	"github.com/Cray-HPE/hms-firmware-action/internal/model"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	logrus "github.com/sirupsen/logrus"
)

// WriteJSON - writes JSON to the open http connection
func WriteJSON(w http.ResponseWriter, i interface{}) {
	obj, err := json.Marshal(i)
	if err != nil {
		logrus.Error(err)
	}
	_, err = w.Write(obj)
	if err != nil {
		logrus.Error(err)
	}
}

// WriteHeaders - writes JSON to the open http connection along with headers
func WriteHeaders(w http.ResponseWriter, pb model.Passback) {
	if pb.IsError{
		w.Header().Add("Content-Type", "application/problem+json")
		w.WriteHeader(pb.StatusCode)
		WriteJSON(w, pb.Error)
	} else if pb.Obj != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(pb.StatusCode)
		switch val := pb.Obj.(type) {
		case []uuid.UUID:
			WriteJSON(w, model.IDList{val})
		case uuid.UUID:
			WriteJSON(w, model.IDResp{val})
		default:
			WriteJSON(w, pb.Obj)
		}
	} else {
		w.WriteHeader(pb.StatusCode)
	}
}

func WriteHeadersWithLocation(w http.ResponseWriter, pb model.Passback, location string) {
	w.Header().Add("Location", location)
	WriteHeaders(w, pb)
}

// GetUUIDFromVars - attempts to retrieve a UUID from an http.Request URL
// returns a passback
func GetUUIDFromVars(key string, r *http.Request) (passback model.Passback) {
	vars := mux.Vars(r)
	value := vars[key]
	logrus.WithFields(logrus.Fields{"key": value}).Debug("Attempting to parse UUID")

	UUID, err := uuid.Parse(value)

	if err != nil {
		passback = model.BuildErrorPassback(http.StatusBadRequest, err)
		logrus.WithFields(logrus.Fields{"ERROR": err,}).Error("Could not parse UUID: " + key)
		return passback
	}
	passback = model.BuildSuccessPassback(http.StatusOK, UUID)
	return passback
}

// While it is generally not a requirement to close request bodies in server
// handlers, it is good practice.  If a body is only partially read, there can
// be a resource leak.  Additionally, if the body is not read at all, the
// server may not be able to reuse the connection.
func DrainAndCloseRequestBody(req *http.Request) {
	if req != nil && req.Body != nil {
		_, _ = io.Copy(io.Discard, req.Body) // ok even if already drained
		req.Body.Close()
	}
}

// Response bodies on the other hand, should always be drained and closed,
// else we leak resources
func DrainAndCloseResponseBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body) // ok even if already drained
		resp.Body.Close()
	}
}