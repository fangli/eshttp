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
	"os"
	"path/filepath"
	"strings"
)

func RecoveryS3File(rootPath string) {
	tempFiles, _ := filepath.Glob(rootPath + "/s3/.*.gz.temp")
	if tempFiles != nil {
		for _, tempFile := range tempFiles {
			base, fname := filepath.Split(tempFile)
			os.Rename(tempFile, base+fname[1:len(fname)-5])
		}
	}

	sendingFiles, _ := filepath.Glob(rootPath + "/s3/.*.gz.sending")
	if sendingFiles != nil {
		for _, sendingFile := range sendingFiles {
			base, fname := filepath.Split(sendingFile)
			os.Rename(sendingFile, base+fname[1:len(fname)-8])
		}
	}

}

func ParseS3FilePath(fname string) string {
	_, s3oriPath := filepath.Split(fname)
	cacheName := strings.Replace(s3oriPath, ":", "/", -1)
	return cacheName[1 : len(cacheName)-8]
}
