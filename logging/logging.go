/*************************************************************************
* This file is a part of msgfiber, A decentralized and distributed message
* synchronization system

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

package logging

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const (
	STDOUT       = 0
	FILE         = 1
	STD_AND_FILE = 2
)

const (
	DEBUG int = iota
	INFO
	WARNING
	ERROR
	FATAL
)

var LevelStr = []string{"DEBUG", "INFO", "WARNING", "ERROR", "FATAL"}

var LevelInt = map[string]int{
	"DEBUG":   0,
	"INFO":    1,
	"WARNING": 2,
	"ERROR":   3,
	"FATAL":   4,
}

type Log struct {
	Dest     int
	Level    int
	FileName string
	logChn   chan string
	runOnce  sync.Once
}

func (l *Log) Writer() {
	f, err := os.OpenFile(l.FileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer f.Close()

	for msg := range l.logChn {

		if (l.Dest == 0) || (l.Dest == 2) {
			fmt.Print(msg)
		}
		if (l.Dest == 1) || (l.Dest == 2) {
			f.WriteString(msg)
		}
	}
	log.Println("Close log file")
}

func (l *Log) initial() {
	l.logChn = make(chan string, 1000)
	go l.Writer()
}

func (l *Log) Shutdown() {
	close(l.logChn)
}

func (l *Log) write(level int, msg string) {

	l.runOnce.Do(l.initial)

	if l.Level > level {
		return
	}
	output := fmt.Sprintf(
		"%s [%s] %s\n",
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		LevelStr[level],
		msg,
	)

	l.logChn <- output

}

func (l *Log) Debug(msg string) {
	l.write(DEBUG, msg)
}

func (l *Log) Info(msg string) {
	l.write(INFO, msg)
}

func (l *Log) Warning(msg string) {
	l.write(WARNING, msg)
}

func (l *Log) Error(msg string) {
	l.write(ERROR, msg)
}

func (l *Log) Fatal(msg string) {
	l.write(FATAL, msg)
	l.write(FATAL, "Exited with return code 1.")
	os.Exit(1)
}
