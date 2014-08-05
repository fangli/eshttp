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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/fangli/eshttp/parsecfg"
)

type StatusInfo struct {
	ModuleName string
	StatusName string
	Value      interface{}
}

type InstanceManager struct {
	Config         *parsecfg.Config
	StartUnixTime  int64
	esChn          chan EsMsg
	s3Chn          chan EsMsg
	esIndexer      *EsIndexer
	esSender       *EsSender
	s3Indexer      *S3Indexer
	s3Sender       *S3Sender
	httpServer     *HttpServer
	statusChn      chan StatusInfo
	status         map[string]map[string]interface{}
	statusLock     sync.Mutex
	shutdownUpdate chan bool
}

func (i *InstanceManager) Shutdown() {
	wg1 := sync.WaitGroup{}
	wg1.Add(3)
	go func() {
		i.Config.AppLog.Info("Shutting-down HTTP server...")
		i.httpServer.Shutdown()
		wg1.Done()
	}()
	go func() {
		i.Config.AppLog.Info("Shutting-down S3 sender...")
		i.s3Sender.Shutdown()
		wg1.Done()
	}()
	go func() {
		i.Config.AppLog.Info("Shutting-down ES sender...")
		i.esSender.Shutdown()
		wg1.Done()
	}()
	wg1.Wait()

	wg2 := sync.WaitGroup{}
	wg2.Add(2)
	go func() {
		i.Config.AppLog.Info("Shutting-down S3 indexer...")
		i.s3Indexer.Shutdown()
		wg2.Done()
	}()
	go func() {
		i.Config.AppLog.Info("Shutting-down ES indexer...")
		i.esIndexer.Shutdown()
		wg2.Done()
	}()
	wg2.Wait()

	i.ShutdownStatusLogger()
}

func (i *InstanceManager) postUrl(url string, b []byte) {
	r := bytes.NewReader(b)
	res, err := http.Post(url, "application/json", r)
	if err != nil {
		return
	}
	_, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()
}

func (i *InstanceManager) postFile(b []byte) {
	err := ioutil.WriteFile(i.Config.Main.StatusFile, b, 0644)
	if err != nil {
		i.Config.AppLog.Error("Err writing status file: " + err.Error())
	}
}

func (i *InstanceManager) UpdateStatus() {
	reloadTime := time.Now().Unix()
	for {
		select {
		case <-time.After(time.Second):
			i.statusLock.Lock()
			i.status["system"] = make(map[string]interface{})
			i.status["system"]["uptime"] = time.Now().Unix() - i.StartUnixTime
			i.status["system"]["last_reload_at"] = reloadTime
			i.status["system"]["seconds_since_reload"] = time.Now().Unix() - reloadTime
			b, _ := json.Marshal(i.status)
			i.statusLock.Unlock()

			if i.Config.Main.StatusFile != "" {
				i.postFile(b)
			}
			if i.Config.Main.StatusUploadUrl != "" {
				i.postUrl(i.Config.Main.StatusUploadUrl, b)
			}

		case <-i.shutdownUpdate:
			return
		}
	}
}

func (i *InstanceManager) StatusLogger() {
	for stat := range i.statusChn {
		i.statusLock.Lock()
		if _, ok := i.status[stat.ModuleName]; !ok {
			i.status[stat.ModuleName] = make(map[string]interface{})
		}
		i.status[stat.ModuleName][stat.StatusName] = stat.Value
		i.statusLock.Unlock()
	}
	i.shutdownUpdate <- true
}

func (i *InstanceManager) ShutdownStatusLogger() {
	close(i.statusChn)
}

func (i *InstanceManager) Run() {

	i.Config.AppLog.Info("Setting max CPU core: " + strconv.Itoa(i.Config.Main.Cores))
	runtime.GOMAXPROCS(i.Config.Main.Cores)

	i.Config.AppLog.Info("Creating HTTP buffer with " + strconv.Itoa(i.Config.Http.HttpBuffer) + " backlog items")
	i.esChn = make(chan EsMsg, i.Config.Http.HttpBuffer)
	i.s3Chn = make(chan EsMsg, i.Config.Http.HttpBuffer)
	i.statusChn = make(chan StatusInfo, 1000)
	i.shutdownUpdate = make(chan bool)
	i.status = make(map[string]map[string]interface{})

	// Start Status logger
	go i.StatusLogger()
	go i.UpdateStatus()

	// Roll back broken transactions, move temp file and sending file back
	i.Config.AppLog.Info("Do some cleanning: recoverying broken transaction and buffer files")
	RecoveryEsFile(i.Config.Main.BufferPath)
	RecoveryS3File(i.Config.Main.BufferPath)

	// Initial elasticsearch indexer instance
	i.Config.AppLog.Info("Initializing ES Indexer...")
	i.esIndexer = &EsIndexer{
		Config:       i.Config,
		EsInput:      i.esChn,
		StatusOutput: i.statusChn,
	}
	i.esIndexer.Run()

	i.Config.AppLog.Info("Initializing ES Sender...")
	i.esSender = &EsSender{
		Config:       i.Config,
		StatusOutput: i.statusChn,
	}
	i.esSender.Run()

	// Initial S3 indexer instance
	i.Config.AppLog.Info("Initializing S3 Indexer...")
	i.s3Indexer = &S3Indexer{
		Config:       i.Config,
		S3Input:      i.s3Chn,
		StatusOutput: i.statusChn,
	}
	i.s3Indexer.Run()

	i.Config.AppLog.Info("Initializing S3 Sender...")
	i.s3Sender = &S3Sender{
		Config:       i.Config,
		StatusOutput: i.statusChn,
	}
	i.s3Sender.Run()

	// Initial HTTP service instance
	i.Config.AppLog.Info("Initializing HTTP server...")
	i.httpServer = &HttpServer{
		Config:       i.Config,
		EsOutput:     i.esChn,
		S3Output:     i.s3Chn,
		StatusOutput: i.statusChn,
	}
	i.httpServer.Run()
}
