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

package parsecfg

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fangli/eshttp/logging"

	"code.google.com/p/gcfg"
	"github.com/crowdmob/goamz/aws"
)

var SYS_VER string
var SYS_BUILD_VER string
var SYS_BUILD_DATE string

type Config struct {
	Main struct {
		Cores        int    `gcfg:"cores"`
		BufferPath   string `gcfg:"buffer-path"`
		LogFile      string `gcfg:"log-file"`
		LogLevel     int
		Raw_LogLevel string `gcfg:"log-level"`
	}
	Http struct {
		Listen           string `gcfg:"listen-address"`
		Port             string `gcfg:"port"`
		Raw_ReadTimeout  string `gcfg:"read-timeout"`
		ReadTimeout      time.Duration
		Raw_WriteTimeout string `gcfg:"write-timeout"`
		WriteTimeout     time.Duration
		Raw_ShutdownWait string `gcfg:"max-shutdown-wait"`
		ShutdownWait     time.Duration
		HttpBuffer       int `gcfg:"http-buffer-docs"`
	}
	Elasticsearch struct {
		Raw_Seed_Nodes   string `gcfg:"nodes"`
		SeedNodes        []string
		BasicUser        string `gcfg:"basic-user"`
		BasicPasswd      string `gcfg:"basic-passwd"`
		MaxBulkSize      int    `gcfg:"max-bulk-size"`
		MaxBulkDocs      int    `gcfg:"max-bulk-docs"`
		Raw_MaxBulkDelay string `gcfg:"max-bulk-delay"`
		MaxBulkDelay     time.Duration
		MaxConcurrent    int `gcfg:"max-connections"`
	}
	S3 struct {
		Raw_AccessKey string `gcfg:"accesskey"`
		Raw_SecretKey string `gcfg:"secret"`
		Auth          *aws.Auth
		Raw_Region    string `gcfg:"region-name"`
		Region        *aws.Region
		Bucket        string `gcfg:"bucket"`
		Path          string `gcfg:"path"`
		MaxConcurrent int    `gcfg:"concurrent-upload"`
	}
	AppLog *logging.Log
}

func mkdirs(path string) error {
	f, err := os.Stat(path)
	if err != nil {
		if os.MkdirAll(path, os.ModePerm) != nil {
			return errors.New("Unable to initialize dir: " + path)
		}
	}

	f, err = os.Stat(path)
	if !f.IsDir() {
		return errors.New(path + " must be a directory")
	}
	return nil
}

func initialDataDir(datadir string) error {

	// Initial all folders
	var err error
	if err = mkdirs(datadir); err != nil {
		return err
	}
	if err = mkdirs(datadir + "/s3"); err != nil {
		return err
	}
	if err = mkdirs(datadir + "/es"); err != nil {
		return err
	}
	return nil
}

func showVersion() {
	fmt.Println("ESHttp: A distributed HTTP service for Elasticsearch indexing")
	fmt.Println("Version", SYS_VER)
	fmt.Println("Build", SYS_BUILD_VER)
	fmt.Println("Compile at", SYS_BUILD_DATE)
	os.Exit(0)
}

func getConfigPath() string {
	configPath := flag.String("c", "/etc/eshttp.conf", "Path of config file")
	version := flag.Bool("version", false, "Show version information")
	v := flag.Bool("v", false, "Show version information")
	flag.Parse()

	if *version || *v {
		showVersion()
	}
	return *configPath
}

func initialDefault() *Config {
	config := new(Config)

	config.Main.Cores = 1
	config.Main.BufferPath = "/mnt/eshttp_buffer/"
	config.Main.LogFile = "/var/log/eshttp.log"
	config.Main.Raw_LogLevel = "INFO"

	config.Http.Listen = "0.0.0.0"
	config.Http.Port = "80"
	config.Http.Raw_ReadTimeout = "30s"
	config.Http.Raw_WriteTimeout = "30s"
	config.Http.Raw_ShutdownWait = "20s"
	config.Http.HttpBuffer = 100000

	config.Elasticsearch.Raw_Seed_Nodes = "localhost"
	config.Elasticsearch.BasicUser = ""
	config.Elasticsearch.BasicPasswd = ""
	config.Elasticsearch.MaxBulkSize = 10240000
	config.Elasticsearch.MaxBulkDocs = 50000
	config.Elasticsearch.Raw_MaxBulkDelay = "30s"
	config.Elasticsearch.MaxConcurrent = 10

	config.S3.Raw_AccessKey = ""
	config.S3.Raw_SecretKey = ""
	config.S3.Raw_Region = "us-east-1"
	config.S3.Bucket = "eshttpdefault"
	config.S3.Path = "/eshttp/{project}/{YYYY}/{MM}/{DD}/{hh}/archive-{group}"
	config.S3.MaxConcurrent = 5

	return config
}

func Parse() *Config {

	configPath := getConfigPath()

	config := initialDefault()
	err := gcfg.ReadFileInto(config, configPath)
	if err != nil {
		log.Fatalf("Failed to read config from %s, Reason: %s", configPath, err)
	}

	// Read timeout parse
	_readTimeout, err := time.ParseDuration(config.Http.Raw_ReadTimeout)
	if err != nil {
		log.Fatalf("Config read-timout is not acceptable")
	}
	config.Http.ReadTimeout = _readTimeout

	// Write timeout parse
	_writeTimeout, err := time.ParseDuration(config.Http.Raw_WriteTimeout)
	if err != nil {
		log.Fatalf("Config write-timout is not acceptable")
	}
	config.Http.WriteTimeout = _writeTimeout

	// Http shutdown wait
	_shutdownWait, err := time.ParseDuration(config.Http.Raw_ShutdownWait)
	if err != nil {
		log.Fatalf("Config max-shutdown-wait is not acceptable")
	}
	config.Http.ShutdownWait = _shutdownWait

	// Max bulk delay parse
	_maxbulkdelay, err := time.ParseDuration(config.Elasticsearch.Raw_MaxBulkDelay)
	if err != nil {
		log.Fatalf("Config max-bulk-delay is not acceptable")
	}
	config.Elasticsearch.MaxBulkDelay = _maxbulkdelay

	// Seed Nodes parse
	config.Elasticsearch.Raw_Seed_Nodes = strings.Replace(
		config.Elasticsearch.Raw_Seed_Nodes, " ", "", -1)
	config.Elasticsearch.SeedNodes = strings.Split(
		config.Elasticsearch.Raw_Seed_Nodes,
		",")

	// Log level upper case
	config.Main.LogLevel = logging.LevelInt[strings.ToUpper(config.Main.Raw_LogLevel)]

	config.Main.BufferPath = strings.TrimRight(config.Main.BufferPath, "/")
	config.S3.Path = strings.TrimRight(config.S3.Path, "/")

	if val, ok := aws.Regions[config.S3.Raw_Region]; ok {
		config.S3.Region = &val
	} else {
		log.Fatalf("S3 Region is invalid")
	}

	if config.S3.Raw_AccessKey == "" || config.S3.Raw_SecretKey == "" {
		log.Fatalf("Unable to initial S3 credential from config")
	}

	config.S3.Auth = &aws.Auth{
		AccessKey: config.S3.Raw_AccessKey,
		SecretKey: config.S3.Raw_SecretKey,
	}

	if config.S3.Bucket == "" {
		log.Fatalf("S3 bucket is invalid")
	}

	if config.S3.Path == "" {
		log.Fatalf("S3 path is invalid")
	}

	err = initialDataDir(config.Main.BufferPath)
	if err != nil {
		log.Fatalf(
			"Unable initializing buffer dir: " +
				config.Main.BufferPath +
				" (" + err.Error() + ")")
	}

	err = mkdirs(filepath.Dir(config.Main.LogFile))
	if err != nil {
		log.Fatalf("Unable initializing log dir: " +
			filepath.Dir(config.Main.LogFile) +
			" (" + err.Error() + ")")
	}

	config.AppLog = &logging.Log{
		Dest:     logging.FILE,
		Level:    config.Main.LogLevel,
		FileName: config.Main.LogFile,
	}

	return config
}
