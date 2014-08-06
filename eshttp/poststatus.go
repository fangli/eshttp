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
	"container/list"
	"sync"
	"time"
)

type PostStatusMsg struct {
	Ts     time.Time
	Lasts  time.Duration
	Size   int64
	Status bool
}

type PostStatus struct {
	ModuleName     string
	StatusName     string
	OutputChn      chan StatusInfo
	msgChn         chan PostStatusMsg
	postStatus     *list.List
	postStatusLock sync.Mutex
	shutdownChn    chan bool
}

func (p *PostStatus) send() {
	var i int = 0
	p.postStatusLock.Lock()
	var lst [50]map[string]interface{}
	for e := p.postStatus.Front(); e != nil; e = e.Next() {
		x := e.Value.(PostStatusMsg)
		lst[i] = make(map[string]interface{})
		lst[i]["result"] = x.Status
		lst[i]["size"] = x.Size
		lst[i]["took"] = x.Lasts.Nanoseconds() / 1000000000
		lst[i]["started"] = x.Ts.Unix()
		i++
	}
	p.postStatusLock.Unlock()
	SendStatus(p.OutputChn, p.ModuleName, p.StatusName, lst)
}

func (p *PostStatus) periodicSend() {
	for {
		select {
		case <-time.After(time.Second):
			p.send()
		case <-p.shutdownChn:
			return
		}
	}
}

func (p *PostStatus) Shutdown() {
	p.shutdownChn <- true
	close(p.msgChn)
}

func (p *PostStatus) Initial() {
	p.msgChn = make(chan PostStatusMsg, 1000)
	p.shutdownChn = make(chan bool)
	p.postStatus = list.New()
	go p.rotatePostStatus()
	go p.periodicSend()
}

func (p *PostStatus) rotatePostStatus() {
	for status := range p.msgChn {
		p.postStatusLock.Lock()
		if p.postStatus.Len() == 50 {
			p.postStatus.Remove(p.postStatus.Front())
		}
		p.postStatus.PushBack(status)
		p.postStatusLock.Unlock()
	}
}

func (p *PostStatus) Update(msg *PostStatusMsg) {
	p.msgChn <- *msg
}
