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
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fangli/eshttp/parsecfg"
	elastigo "github.com/mattbaird/elastigo/lib"
)

type EsIndexer struct {
	Config         *parsecfg.Config
	EsInput        chan EsMsg
	StatusOutput   chan StatusInfo
	chunkCurrent   int64
	chunkStatsChn  chan int64
	shutdownStatus chan bool
	isRunning      bool
	isRunningMutex sync.Mutex
	indexer        *elastigo.BulkIndexer
}

func (e *EsIndexer) Shutdown() {
	e.Config.AppLog.Info("Waiting for ES input buffer empty...")
	for {
		if len(e.EsInput) == 0 {
			close(e.EsInput)
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	e.Config.AppLog.Info("ES input buffer closed...")
	e.Config.AppLog.Info("Stopping and flush ES buffer to local FS...")
	e.indexer.Stop()
	e.shutdownStatus <- true
	e.Config.AppLog.Info("ES buffer flushed and closed")

}

func (e *EsIndexer) runIndexer() {
	e.indexer.Start()
	e.Config.AppLog.Info("ES indexer started")
	for esMsg := range e.EsInput {
		e.chunkStatsChn <- int64(len(esMsg.Doc))
		e.indexer.Index(indexParser(esMsg.Index), esMsg.Type, "", "", nil, esMsg.Doc, false)
	}
}

func (e *EsIndexer) rotateStatus() {
	for sz := range e.chunkStatsChn {
		atomic.AddInt64(&e.chunkCurrent, sz)
	}
	e.Config.AppLog.Info("Status reporter of ES Indexer stopped")
}

func (e *EsIndexer) SendRatios() {
	var lastByte int64 = 0
	var total int64 = 0
	for {
		select {
		case <-time.After(time.Second * 10):
			total = atomic.LoadInt64(&e.chunkCurrent)
			SendStatus(e.StatusOutput, "es_indexer", "bytes_per_second", (total-lastByte)/10)
			lastByte = total
		case <-e.shutdownStatus:
			close(e.chunkStatsChn)
			return
		}
	}
}

func (e *EsIndexer) initial() {
	e.chunkStatsChn = make(chan int64, 100000)
	e.shutdownStatus = make(chan bool)
}

func (e *EsIndexer) Run() {

	e.initial()

	go e.rotateStatus()
	go e.SendRatios()

	e.Config.AppLog.Info(
		"Starting ES local indexer with" +
			" max-bulk-size=" + strconv.Itoa(e.Config.Elasticsearch.MaxBulkSize) +
			" max-bulk-docs=" + strconv.Itoa(e.Config.Elasticsearch.MaxBulkDocs) +
			" max-bulk-delay=" + e.Config.Elasticsearch.MaxBulkDelay.String())
	elastConn := elastigo.NewConn()
	e.indexer = elastConn.NewBulkIndexer(10)
	e.indexer.BulkMaxBuffer = e.Config.Elasticsearch.MaxBulkSize
	e.indexer.BulkMaxDocs = e.Config.Elasticsearch.MaxBulkDocs
	e.indexer.BufferDelayMax = e.Config.Elasticsearch.MaxBulkDelay
	e.indexer.Sender = func(buf *bytes.Buffer) error {
		WriteEsCacheFile(e.Config.Main.BufferPath, buf)
		return nil
	}

	go e.runIndexer()
}
