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
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/fangli/eshttp/parsecfg"
	elastigo "github.com/mattbaird/elastigo/lib"
)

type EsSender struct {
	Config          *parsecfg.Config
	inputChunkFile  chan string
	doneSenderChan  chan bool
	doneScannerChan chan bool
	esConn          *elastigo.Conn
}

func (e *EsSender) bufferScanner() {
	for {
		select {
		case <-time.After(time.Millisecond):
			tempFiles, _ := filepath.Glob(e.Config.Main.BufferPath + "/es/*.buffer")
			if tempFiles == nil {
				time.Sleep(time.Millisecond * 200)
			} else {
				sort.Strings(tempFiles)
				e.Config.AppLog.Debug("ES buffer scanner found chunk " + tempFiles[0] + ", make it ready for sending.")
				e.inputChunkFile <- MakeSendReady(tempFiles[0])
			}
		case <-e.doneScannerChan:
			return
		}
	}
}

func (e *EsSender) send(chunkName string) error {
	var err error
	b, err := ioutil.ReadFile(chunkName)
	if err != nil {
		panic(err)
	}
	_, err = e.esConn.DoCommand("POST", "/_bulk", nil, b)
	return err
}

func (e *EsSender) sender() {
	for {
		select {
		case chunk := <-e.inputChunkFile:
			err := e.send(chunk)
			if err != nil {
				e.Config.AppLog.Warning("ES Sender err: " + err.Error())
				e.Config.AppLog.Warning("Rollback transaction for chunk cache " + chunk)
				RollbackChunk(chunk)
			} else {
				e.Config.AppLog.Debug("ES chunk file sent successfully: " + chunk)
				FinishChunk(chunk)
			}
		case <-e.doneSenderChan:
			return
		}
	}
}

func (e *EsSender) Shutdown() {
	e.Config.AppLog.Info("Shutting-down ES scanner channel...")
	e.doneScannerChan <- true
	e.Config.AppLog.Info("Shutting-down all ES sender threads...")
	for i := 0; i < e.Config.Elasticsearch.MaxConcurrent; i++ {
		e.doneSenderChan <- true
	}
	e.Config.AppLog.Info("ES Sender stopped")
}

func (e *EsSender) Run() {

	e.Config.AppLog.Info(
		"Starting ES Sender with" +
			" hosts=" + e.Config.Elasticsearch.Raw_Seed_Nodes +
			" username=" + e.Config.Elasticsearch.BasicUser +
			" password=" + e.Config.Elasticsearch.BasicPasswd)
	e.esConn = elastigo.NewConn()
	e.esConn.Hosts = e.Config.Elasticsearch.SeedNodes
	e.esConn.Username = e.Config.Elasticsearch.BasicUser
	e.esConn.Password = e.Config.Elasticsearch.BasicPasswd

	e.doneSenderChan = make(chan bool)
	e.doneScannerChan = make(chan bool)
	e.inputChunkFile = make(chan string)

	e.Config.AppLog.Info("Spawning " + strconv.Itoa(e.Config.Elasticsearch.MaxConcurrent) + " ES sender threads")
	for i := 0; i < e.Config.Elasticsearch.MaxConcurrent; i++ {
		go e.sender()
	}

	// inputChunkFile is a chan that contains the filename of ready-to-send
	// elasticsearch chunk file, and EsBufferScanner() will scan those files and
	// make them ready to post, then push the oldest one to chan.
	go e.bufferScanner()

}
