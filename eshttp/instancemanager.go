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
	"runtime"
	"strconv"
	"sync"

	"github.com/fangli/eshttp/parsecfg"
)

type InstanceManager struct {
	Config     *parsecfg.Config
	esChn      chan EsMsg
	s3Chn      chan EsMsg
	esIndexer  *EsIndexer
	esSender   *EsSender
	s3Indexer  *S3Indexer
	s3Sender   *S3Sender
	httpServer *HttpServer
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
}

func (i *InstanceManager) Run() {

	i.Config.AppLog.Info("Setting max CPU core: " + strconv.Itoa(i.Config.Main.Cores))
	runtime.GOMAXPROCS(i.Config.Main.Cores)

	i.Config.AppLog.Info("Creating HTTP buffer with " + strconv.Itoa(i.Config.Http.HttpBuffer) + " backlog items")
	i.esChn = make(chan EsMsg, i.Config.Http.HttpBuffer)
	i.s3Chn = make(chan EsMsg, i.Config.Http.HttpBuffer)

	// Roll back broken transactions, move temp file and sending file back
	i.Config.AppLog.Info("Do some cleanning: recoverying broken transaction and buffer files")
	RecoveryEsFile(i.Config.Main.BufferPath)
	RecoveryS3File(i.Config.Main.BufferPath)

	// Initial elasticsearch indexer instance
	i.Config.AppLog.Info("Initializing ES Indexer...")
	i.esIndexer = &EsIndexer{
		Config:  i.Config,
		EsInput: i.esChn,
	}
	i.esIndexer.Run()

	i.Config.AppLog.Info("Initializing ES Sender...")
	i.esSender = &EsSender{
		Config: i.Config,
	}
	i.esSender.Run()

	// Initial S3 indexer instance
	i.Config.AppLog.Info("Initializing S3 Indexer...")
	i.s3Indexer = &S3Indexer{
		Config:  i.Config,
		S3Input: i.s3Chn,
	}
	i.s3Indexer.Run()

	i.Config.AppLog.Info("Initializing S3 Sender...")
	i.s3Sender = &S3Sender{
		Config: i.Config,
	}
	i.s3Sender.Run()

	// Initial HTTP service instance
	i.Config.AppLog.Info("Initializing HTTP server...")
	i.httpServer = &HttpServer{
		Config:   i.Config,
		EsOutput: i.esChn,
		S3Output: i.s3Chn,
	}
	i.httpServer.Run()
}
