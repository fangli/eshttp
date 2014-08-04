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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fangli/eshttp/eshttp"
	"github.com/fangli/eshttp/parsecfg"
)

// Main
func main() {

	var im eshttp.InstanceManager
	var reloadTimes int64 = 0
	startTime := time.Now()

	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGHUP,
	)

	configManager := parsecfg.ConfigManager{}
	configManager.Initial()

	for {
		select {
		case config := <-configManager.ReloadChn:
			if reloadTimes != 0 {
				config.AppLog.Info("Load new configuration from " + configManager.Source)
				config.AppLog.Info("Shutting-down eshttp instance manager...")
				im.Shutdown()
				config.AppLog.Info("Eshttp instance manager stopped...")
				time.Sleep(time.Second)
			}

			config.AppLog.Info("Starting system...")
			config.AppLog.Info("This is the " + strconv.FormatInt(reloadTimes+1, 10) + " time(s) reloading since " + startTime.UTC().String())
			config.AppLog.Info("Initializing eshttp instance manager...")

			im = eshttp.InstanceManager{
				Config: config,
			}
			im.Run()
			reloadTimes++
		case sig := <-signalChannel:
			if sig == syscall.SIGHUP {
				im.Config.AppLog.Info("Signal SIGHUP received, reload configuration...")
				err := configManager.SendReload()
				if err != nil {
					im.Config.AppLog.Error(err.Error())
				}
			} else {
				fmt.Println("Signal stop received, stopping eshttp...")
				fmt.Println("This may takes up to a few minutes, see " + im.Config.Main.LogFile + " for details...")
				im.Config.AppLog.Info("Signal stop received, stopping eshttp...")
				im.Shutdown()
				im.Config.AppLog.Info("Eshttp exited.")
				os.Exit(0)
			}

		}
	}
}
