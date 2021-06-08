/*
Copyright (C) 2021 The Self-host Authors.
This file is part of Self-host <https://github.com/self-host/self-host>.

Self-host is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Self-host is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with dogtag.  If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic("zap.NewProduction " + err.Error())
	}
}

func main() {
	viper.SetConfigName(os.Getenv("CONFIG_FILENAME"))
	viper.SetConfigType("yaml")

	// How do we handle multiple OS?
	viper.AddConfigPath("/etc/selfhost/")
	viper.AddConfigPath("$HOME/.config/selfhost")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatal("Fatal error config file", zap.Error(err))
	}

	quit := make(chan struct{})
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-ctx.Done()
		stop()
		logger.Info("Shutdown signal received")
		close(quit)
	}()
	defer logger.Sync()

	// API server
	s_errC, err := Server(quit, fmt.Sprintf("%v:%v", viper.GetString("listen.host"), viper.GetInt("listen.port")))
	if err != nil {
		logger.Fatal("Fatal error couldn't run", zap.Error(err))
	}

	// Program Manager
	pm_errC, err := ProgramManager(quit)
	if err != nil {
		logger.Fatal("Fatal error couldn't create", zap.Error(err))
	}

	waitfor := 2

	// Wait for API server to terminate
	for waitfor > 0 {
		select {
		case err := <-s_errC:
			waitfor -= 1
			if err != nil {
				logger.Fatal("Fatal error while running", zap.Error(err))
				goto giveup
			}

			// Wait for Program Manager to terminate
		case err := <-pm_errC:
			waitfor -= 1
			if err != nil {
				logger.Fatal("Fatal error while running", zap.Error(err))
				goto giveup
			}
		}
	}

giveup:

	logger.Info("Shutdown complete")
}
