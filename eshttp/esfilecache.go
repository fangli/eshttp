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
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func GenerateNewFile(rootPath string) (string, string) {
	fn := strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + randString(8)
	return rootPath + "/es/." + fn + ".buffer.temp", rootPath + "/es/" + fn + ".buffer"
}

func RecoveryEsFile(rootPath string) {
	tempFiles, _ := filepath.Glob(rootPath + "/es/.*.buffer.temp")
	if tempFiles != nil {
		for _, tempFile := range tempFiles {
			base, fname := filepath.Split(tempFile)
			os.Rename(tempFile, base+fname[1:len(fname)-5])
		}
	}

	sendingFiles, _ := filepath.Glob(rootPath + "/es/.*.buffer.sending")
	if sendingFiles != nil {
		for _, sendingFile := range sendingFiles {
			base, fname := filepath.Split(sendingFile)
			os.Rename(sendingFile, base+fname[1:len(fname)-8])
		}
	}

}

func MakeSendReady(fn string) string {
	base, fname := filepath.Split(fn)
	err := os.Rename(fn, base+"."+fname+".sending")
	if err != nil {
		panic(err)
	}
	return base + "." + fname + ".sending"
}

func RollbackChunk(fn string) {
	base, fname := filepath.Split(fn)
	err := os.Rename(fn, base+fname[1:len(fname)-8])
	if err != nil {
		panic(err)
	}
}

func FinishChunk(fn string) {
	err := os.Remove(fn)
	if err != nil {
		panic(err)
	}
}

func WriteEsCacheFile(rootPath string, buf *bytes.Buffer) {
	var err error
	fnOld, fnNew := GenerateNewFile(rootPath)
	f, err := os.Create(fnOld)
	if err != nil {
		panic(err)
	}
	defer os.Rename(fnOld, fnNew)
	defer f.Close()
	_, err = buf.WriteTo(f)
	if err != nil {
		panic(err)
	}
}
