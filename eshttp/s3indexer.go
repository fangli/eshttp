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
	"compress/gzip"
	"os"
	"sync"
	"time"

	"github.com/fangli/eshttp/parsecfg"
)

type S3Chunk struct {
	gzipWriter *gzip.Writer
	fd         *os.File
	Project    string
	Group      string
	cacheName  string
	expires    time.Time
	storeName  string
}

func (s *S3Chunk) Close() {
	s.gzipWriter.Close()
	s.fd.Close()
	err := os.Rename(s.cacheName, s.storeName)
	if err != nil {
		panic(err)
	}
}

type S3ChunkManager struct {
	Config     *parsecfg.Config
	chunks     map[string]*S3Chunk
	shutdownWg sync.WaitGroup
}

func (s *S3ChunkManager) NewS3Chunk(project string, group string) *S3Chunk {
	var err error
	filename, expires := NewS3CacheName(s.Config.S3.Path, project, group)
	cacheName := s.Config.Main.BufferPath + "/s3/" + "." + filename + ".gz.temp"
	storeName := s.Config.Main.BufferPath + "/s3/" + filename + ".gz"
	f, err := os.Create(cacheName)
	if err != nil {
		panic(err)
	}
	return &S3Chunk{
		gzipWriter: gzip.NewWriter(f),
		fd:         f,
		Project:    project,
		Group:      group,
		expires:    expires,
		cacheName:  cacheName,
		storeName:  storeName,
	}
}

func (s *S3ChunkManager) WriteChunk(chunk EsMsg) {
	var idx = chunk.Index + ":" + chunk.Type
	if s.chunks[idx] == nil {
		s.chunks[idx] = s.NewS3Chunk(chunk.Index, chunk.Type)
	} else {
		if time.Now().UTC().After(s.chunks[idx].expires) {
			s.chunks[idx].Close()
			delete(s.chunks, idx)
			s.chunks[idx] = s.NewS3Chunk(chunk.Index, chunk.Type)
		}
	}
	s.chunks[idx].gzipWriter.Write(append(chunk.Doc, '\n'))
}

func (s *S3ChunkManager) Shutdown() {
	for idx, s3chunk := range s.chunks {
		s3chunk.Close()
		delete(s.chunks, idx)
	}
}

func (s *S3ChunkManager) Initial() {
	s.chunks = make(map[string]*S3Chunk)
}

type S3Indexer struct {
	Config      *parsecfg.Config
	S3Input     chan EsMsg
	shutdownChn chan bool
}

func (s *S3Indexer) WriteS3Cache() {

	chunkManager := &S3ChunkManager{
		Config: s.Config,
	}
	chunkManager.Initial()

	for chunk := range s.S3Input {
		chunkManager.WriteChunk(chunk)
	}
	chunkManager.Shutdown()
	s.shutdownChn <- true
}

func (s *S3Indexer) Shutdown() {
	for {
		if len(s.S3Input) == 0 {
			close(s.S3Input)
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	<-s.shutdownChn
}

func (s *S3Indexer) Run() {
	s.shutdownChn = make(chan bool)
	go s.WriteS3Cache()
}
