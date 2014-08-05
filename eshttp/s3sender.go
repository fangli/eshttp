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
	"strconv"
	"time"

	"github.com/crowdmob/goamz/s3"
	"github.com/fangli/eshttp/parsecfg"
)

type S3Sender struct {
	Config          *parsecfg.Config
	StatusOutput    chan StatusInfo
	bucket          *s3.Bucket
	inputChunkFile  chan string
	doneSenderChan  chan bool
	doneScannerChan chan bool
	postStatus      *PostStatus
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
				s.Config.AppLog.Debug("Found S3 buffered chunk for sending: " + tempFiles[0])
				s.inputChunkFile <- MakeSendReady(tempFiles[0])
				SendStatus(s.StatusOutput, "s3_uploader", "file_buffer_size", GlobSize(tempFiles[1:]))
			}
		case <-s.doneScannerChan:
			return
		}
	}
}

func (s *S3Sender) send(chunkName string) (int64, error) {
	var err error
	f, err := os.Open(chunkName)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	b := bufio.NewReader(f)
	stat, err := f.Stat()
	if err != nil {
		panic(err)
	}
	err = s.bucket.PutReader(ParseS3FilePath(chunkName), b, stat.Size(), "binary/octet-stream", s3.Private, s3.Options{})
	return stat.Size(), err
}

func (s *S3Sender) sender() {
	var t0 time.Time
	var delta time.Duration
	var buf_size int64
	var err error
	for {
		select {
		case chunk := <-s.inputChunkFile:
			t0 = time.Now()
			buf_size, err = s.send(chunk)
			delta = time.Since(t0)

			s.postStatus.Update(&PostStatusMsg{
				Ts:     t0,
				Lasts:  delta,
				Size:   buf_size,
				Status: err == nil,
			})

			if err != nil {
				s.Config.AppLog.Warning("Error uploading S3 files: " + err.Error() + ", rollback transaction.")
				RollbackChunk(chunk)
			} else {
				FinishChunk(chunk)
				s.Config.AppLog.Debug("S3 chunk uploaded successfully: " + chunk)
			}
		case <-s.doneSenderChan:
			return
		}
	}
}

func (s *S3Sender) Shutdown() {
	s.Config.AppLog.Info("Shutting-down S3 chunk scanner...")
	s.doneScannerChan <- true
	s.Config.AppLog.Info("Shutting-down S3 chunk sender...")
	for i := 0; i < s.Config.S3.MaxConcurrent; i++ {
		s.doneSenderChan <- true
	}
	s.postStatus.Shutdown()
	s.Config.AppLog.Info("S3 sender stopped")
}

func (s *S3Sender) initialBucket(bucket *s3.Bucket) {
	bucket.PutBucket(s3.Private)
}

func (s *S3Sender) Run() {

	s.Config.AppLog.Info(
		"Starting S3 Sender with" +
			" accesskey=" + s.Config.S3.Raw_AccessKey +
			" region=" + s.Config.S3.Raw_Region +
			" bucket=" + s.Config.S3.Bucket)
	s3Conn := s3.New(*s.Config.S3.Auth, *s.Config.S3.Region)
	s.bucket = s3Conn.Bucket(s.Config.S3.Bucket)
	s.initialBucket(s.bucket)

	s.doneSenderChan = make(chan bool)
	s.doneScannerChan = make(chan bool)
	s.inputChunkFile = make(chan string)

	s.postStatus = &PostStatus{
		ModuleName: "s3_uploader",
		StatusName: "upload_status",
		OutputChn:  s.StatusOutput,
	}
	s.postStatus.Initial()

	s.Config.AppLog.Info("Spawning " + strconv.Itoa(s.Config.S3.MaxConcurrent) + " threads for S3 uploading")
	for i := 0; i < s.Config.S3.MaxConcurrent; i++ {
		go s.sender()
	}

	s.Config.AppLog.Info("Starting S3 chunk scanner...")
	go s.bufferScanner()

}
