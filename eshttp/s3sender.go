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
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/crowdmob/goamz/s3"
	"github.com/fangli/eshttp/parsecfg"
)

type S3Sender struct {
	Config          *parsecfg.Config
	bucket          *s3.Bucket
	inputChunkFile  chan string
	doneSenderChan  chan bool
	doneScannerChan chan bool
}

func (s *S3Sender) bufferScanner() {
	for {
		select {
		case <-time.After(time.Millisecond):
			tempFiles, _ := filepath.Glob(s.Config.Main.BufferPath + "/s3/*.gz")
			if tempFiles == nil {
				time.Sleep(time.Millisecond * 200)
			} else {
				sort.Strings(tempFiles)
				s.inputChunkFile <- MakeSendReady(tempFiles[0])
			}
		case <-s.doneScannerChan:
			return
		}
	}
}

func (s *S3Sender) send(chunkName string) error {
	var err error
	f, err := os.Open(chunkName)
	if err != nil {
		panic(err)
	}
	b := bufio.NewReader(f)
	stat, err := f.Stat()
	if err != nil {
		panic(err)
	}
	err = s.bucket.PutReader(ParseS3FilePath(chunkName), b, stat.Size(), "binary/octet-stream", s3.Private, s3.Options{})
	return err
}

func (s *S3Sender) sender() {
	for {
		select {
		case chunk := <-s.inputChunkFile:
			err := s.send(chunk)
			if err != nil {
				s.Config.AppLog.Error(err.Error())
				RollbackChunk(chunk)
			} else {
				FinishChunk(chunk)
			}
		case <-s.doneSenderChan:
			return
		}
	}
}

func (s *S3Sender) Shutdown() {
	s.doneScannerChan <- true
	for i := 0; i < s.Config.S3.MaxConcurrent; i++ {
		s.doneSenderChan <- true
	}
}

func (s *S3Sender) initialBucket(bucket *s3.Bucket) {
	bucket.PutBucket(s3.Private)
}

func (s *S3Sender) Run() {

	s3Conn := s3.New(*s.Config.S3.Auth, *s.Config.S3.Region)
	s.bucket = s3Conn.Bucket(s.Config.S3.Bucket)
	s.initialBucket(s.bucket)

	s.doneSenderChan = make(chan bool)
	s.doneScannerChan = make(chan bool)
	s.inputChunkFile = make(chan string)

	for i := 0; i < s.Config.S3.MaxConcurrent; i++ {
		go s.sender()
	}

	// inputChunkFile is a chan that contains the filename of ready-to-send
	// S3 chunk file, and EsBufferScanner() will scan those files and
	// make them ready to post, then push the oldest one to chan.
	go s.bufferScanner()

}
