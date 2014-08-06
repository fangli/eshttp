package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"code.google.com/p/gcfg"
)

type Param struct {
	Daemon struct {
		Args string `gcfg:"args"`
	}
	Main struct {
		Cores           int    `gcfg:"cores"`
		BufferPath      string `gcfg:"buffer-path"`
		LogFile         string `gcfg:"log-file"`
		Raw_LogLevel    string `gcfg:"log-level"`
		StatusFile      string `gcfg:"status-file"`
		StatusUploadUrl string `gcfg:"status-upload-url"`
	}
	Http struct {
		Listen           string `gcfg:"listen-address"`
		Port             string `gcfg:"port"`
		Raw_ReadTimeout  string `gcfg:"read-timeout"`
		Raw_WriteTimeout string `gcfg:"write-timeout"`
		Raw_ShutdownWait string `gcfg:"max-shutdown-wait"`
		HttpBuffer       int    `gcfg:"http-buffer-docs"`
	}
	Elasticsearch struct {
		Raw_Seed_Nodes   string `gcfg:"nodes"`
		BasicUser        string `gcfg:"basic-user"`
		BasicPasswd      string `gcfg:"basic-passwd"`
		MaxBulkSize      int    `gcfg:"max-bulk-size"`
		MaxBulkDocs      int    `gcfg:"max-bulk-docs"`
		Raw_MaxBulkDelay string `gcfg:"max-bulk-delay"`
		MaxConcurrent    int    `gcfg:"max-connections"`
	}
	S3 struct {
		Raw_AccessKey string `gcfg:"accesskey"`
		Raw_SecretKey string `gcfg:"secret"`
		Raw_Region    string `gcfg:"region-name"`
		Bucket        string `gcfg:"bucket"`
		Path          string `gcfg:"path"`
		MaxConcurrent int    `gcfg:"concurrent-upload"`
	}
}

func getConfigPath() string {
	configPath := flag.String("c", "/etc/eshttp.conf", "Path of config file or URI")
	flag.Parse()
	return *configPath
}

func main() {
	var err error
	param := new(Param)
	param.Daemon.Args = "-c /etc/eshttp.conf -p /var/run/eshttp.pid"

	err = gcfg.ReadFileInto(param, getConfigPath())

	if err != nil {
		log.Fatalln(err.Error())
	}

	cmd := exec.Command("eshttp", strings.Split(param.Daemon.Args, " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(2 * time.Second):
		os.Exit(0)
	case err := <-done:
		if err != nil {
			log.Fatalf("")
		}
	}

}
