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
	"crypto/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	_datetimeLayout = []string{
		"{YYYY}", "2006",
		"{YY}", "06",
		"{MM}", "01",
		"{DD}", "02",
		"{hh}", "15",
		"{mm}", "04",
		"{ss}", "05",
	}
	_datetimeReplacer = strings.NewReplacer(_datetimeLayout...)
)

func fileSize(file string) (int64, error) {
	f, e := os.Stat(file)
	if e != nil {
		return 0, e
	}
	return f.Size(), nil
}

func GlobSize(files []string) int64 {
	var tmpSize, ret int64
	var err error

	for _, f := range files {
		tmpSize, err = fileSize(f)
		if err != nil {
			return -1
		}
		ret += tmpSize
	}
	return ret
}

func randString(n int) string {
	const rainbow = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = rainbow[b%byte(len(rainbow))]
	}
	return string(bytes)
}

func SendStatus(statusChn chan StatusInfo, moduleName string, statusName string, value interface{}) {
	statusChn <- StatusInfo{
		ModuleName: moduleName,
		StatusName: statusName,
		Value:      value,
	}
}

func MakeDirs(path string) {
	_, err := os.Stat(path)
	if err != nil {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

func dateFormatter(pattern string) string {
	return time.Now().UTC().Format(_datetimeReplacer.Replace(pattern))
}

func indexParser(prefix string) string {
	return prefix + "-" + time.Now().UTC().Format("2006.01.02")
}

func GetChunkExpires(pattern string, now time.Time) time.Time {

	if strings.Contains(pattern, "{ss}") {
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1, 0, time.UTC)
	}
	if strings.Contains(pattern, "{mm}") {
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, time.UTC)
	}
	if strings.Contains(pattern, "{hh}") {
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, time.UTC)
	}
	if strings.Contains(pattern, "{DD}") {
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	}
	if strings.Contains(pattern, "{MM}") {
		return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	}
	if strings.Contains(pattern, "{YYYY}") {
		return time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if strings.Contains(pattern, "{YY}") {
		return time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(now.Year()+999, 1, 1, 0, 0, 0, 0, time.UTC)
}

func NewS3CacheName(pattern string, project string, group string) (string, time.Time) {
	now := time.Now().UTC()
	expires := GetChunkExpires(pattern, now)
	name := now.Format(_datetimeReplacer.Replace(pattern))
	name = strings.Replace(name, "{project}", project, -1)
	name = strings.Replace(name, "{group}", group, -1)
	name = strings.Replace(name, "/", ":", -1)
	return name + "." + strconv.FormatInt(time.Now().UTC().UnixNano(), 10) + "." + randString(8), expires
}
