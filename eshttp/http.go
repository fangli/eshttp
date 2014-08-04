/*************************************************************************
* This file is a part of eshttp, A decentralized and distributed HTTP
* Service for bulked and buffered Elasticseatch index

* Copyright (C) 2014  Fang Li <surivlee@gmail.com> and Funplus, Inc.
*
* This program is free software; you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation; either version 2 of the License, or
* (at your option) any later version.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
* GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License along
* with this program; if not, see http://www.gnu.org/licenses/gpl-2.0.html
*************************************************************************/

package eshttp

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fangli/eshttp/parsecfg"
)

// The replay when request finished
var OkReply = []byte("OK")

// The replay when flash crossdomain request
var CrossdomainReplay = []byte(`<?xml version="1.0"?>
<cross-domain-policy>
  <allow-access-from domain="*"/>
</cross-domain-policy>
`)

type HttpServer struct {
	Config         *parsecfg.Config
	EsOutput       chan EsMsg
	S3Output       chan EsMsg
	isRunning      bool
	isRunningMutex sync.Mutex
	requestWg      sync.WaitGroup
	ln             *net.Listener
	serve          *http.Server
	denyRequestChn chan bool
}

func (h *HttpServer) SetStopStatus() {
	h.isRunningMutex.Lock()
	h.isRunning = false
	h.isRunningMutex.Unlock()
	h.serve.SetKeepAlivesEnabled(false)
	h.Config.AppLog.Info("HTTP server comes into shutting-down status, wait for up to " + h.Config.Http.ShutdownWait.String())
	h.Config.AppLog.Info("Waiting for Load balancer removing self from cluster...")
	time.Sleep(h.Config.Http.ShutdownWait)
}

func (h *HttpServer) CloseSocket() {
	ln := *h.ln
	ln.Close()
	h.Config.AppLog.Info("TCP socket closed, all further requests will be rejected")
	h.denyRequestChn <- true
	h.Config.AppLog.Info("Wait until all in flight request finished")
	h.requestWg.Wait()
}

func (h *HttpServer) Shutdown() {
	h.Config.AppLog.Info("Set /status to 503(shutting-down) and keep alives to disabled")
	h.SetStopStatus()
	h.CloseSocket()
	h.Config.AppLog.Info("HTTP server stopped")
}

func (h *HttpServer) parseJson(rawJson []byte) ([]byte, error) {
	rawLen := len(rawJson)
	// Check the length and first "{" and last "}"
	if rawLen < 7 || rawJson[0] != 123 || rawJson[rawLen-1] != 125 {
		return []byte(""), errors.New("400 Invalid JSON: " + string(rawJson))
	}
	payload := []byte("{\"@timestamp\":" + strconv.FormatInt(time.Now().Unix(), 10) + "000,")
	return append(payload, rawJson[1:rawLen]...), nil
}

func (h *HttpServer) logHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	var params url.Values
	var scanner *bufio.Scanner
	var project, group, isEs, isArchive = "", "", "", ""

	h.requestWg.Add(1)
	defer h.requestWg.Done()

	params, err = url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		h.Config.AppLog.Warning("400 Invalid parameters received: " + r.URL.RawQuery)
		http.Error(w, "Invalid param", 400)
		return
	}

	isArchive = params.Get("archive")
	isEs = params.Get("es")

	project = params.Get("project")
	group = params.Get("group")
	if project == "" || group == "" {
		h.Config.AppLog.Warning("400 Bad parameters received: " + r.URL.RawQuery)
		http.Error(w, "Invalid param", 400)
		return
	}

	// For a better performance let's make the loop inside
	if isEs == "1" {
		scanner = bufio.NewScanner(r.Body)
		for scanner.Scan() {
			msg := scanner.Bytes()
			payload, err := h.parseJson(msg)
			if err == nil {
				h.EsOutput <- EsMsg{
					Index: project,
					Type:  group,
					Doc:   payload,
				}
				if isArchive == "1" {
					h.S3Output <- EsMsg{
						Index: project,
						Type:  group,
						Doc:   msg,
					}
				}
			} else {
				h.Config.AppLog.Warning(err.Error())
			}
		}
	} else if isArchive == "1" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			h.Config.AppLog.Warning(err.Error())
		} else {
			h.S3Output <- EsMsg{
				Index: project,
				Type:  group,
				Doc:   bytes.TrimSpace(body),
			}
		}
	}
	w.Write(OkReply)
}

func (h *HttpServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	h.isRunningMutex.Lock()
	status := h.isRunning
	h.isRunningMutex.Unlock()
	if status {
		w.Write(OkReply)
	} else {
		http.Error(w, "Server is shutting-down!", 503)
	}

}

func (h *HttpServer) corsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=604800")
	w.Write(CrossdomainReplay)
}

func (h *HttpServer) createServer(mux *http.ServeMux, port string) {
	var err error
	ln, err := net.Listen("tcp", h.Config.Http.Listen+":"+port)
	if err != nil {
		h.Config.AppLog.Fatal(err.Error())
	}
	h.ln = &ln

	h.serve = &http.Server{
		Handler:      mux,
		ReadTimeout:  h.Config.Http.ReadTimeout,
		WriteTimeout: h.Config.Http.WriteTimeout,
		ConnState: func(conn net.Conn, state http.ConnState) {
			if len(h.denyRequestChn) == 1 {
				conn.Close()
			}
		},
	}
}

func (h *HttpServer) runServer() {
	h.Config.AppLog.Info("HTTP server started")
	err := h.serve.Serve(*h.ln)
	if !strings.Contains(err.Error(), "accept") {
		h.Config.AppLog.Fatal(err.Error())
	}

}

func (h *HttpServer) Run() {

	h.isRunning = true
	h.denyRequestChn = make(chan bool, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/log", h.logHandler)
	mux.HandleFunc("/crossdomain.xml", h.corsHandler)
	mux.HandleFunc("/status", h.statusHandler)

	h.Config.AppLog.Info("Binding TCP socket " + h.Config.Http.Listen + ":" + h.Config.Http.Port + " for HTTP service")
	h.createServer(mux, h.Config.Http.Port)
	go h.runServer()
}
