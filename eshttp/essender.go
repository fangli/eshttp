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
				RollbackChunk(chunk)
			} else {
				FinishChunk(chunk)
			}
		case <-e.doneSenderChan:
			return
		}
	}
}

func (e *EsSender) Shutdown() {
	e.doneScannerChan <- true
	for i := 0; i < e.Config.Elasticsearch.MaxConcurrent; i++ {
		e.doneSenderChan <- true
	}
}

func (e *EsSender) Run() {

	e.esConn = elastigo.NewConn()
	e.esConn.Hosts = e.Config.Elasticsearch.SeedNodes
	e.esConn.Username = e.Config.Elasticsearch.BasicUser
	e.esConn.Password = e.Config.Elasticsearch.BasicPasswd

	e.doneSenderChan = make(chan bool)
	e.doneScannerChan = make(chan bool)
	e.inputChunkFile = make(chan string)

	for i := 0; i < e.Config.Elasticsearch.MaxConcurrent; i++ {
		go e.sender()
	}

	// inputChunkFile is a chan that contains the filename of ready-to-send
	// elasticsearch chunk file, and EsBufferScanner() will scan those files and
	// make them ready to post, then push the oldest one to chan.
	go e.bufferScanner()

}
