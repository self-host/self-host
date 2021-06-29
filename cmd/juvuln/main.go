// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/self-host/self-host/pkg/configdir"
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
	for _, p := range configdir.SystemConfig("selfhost") {
		viper.AddConfigPath(p)
	}
	for _, p := range configdir.LocalConfig("selfhost") {
		viper.AddConfigPath(p)
	}
	viper.AddConfigPath(".")

	viper.SetDefault("worker.timeout", 30*time.Second)

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
	sErrC, err := Server(quit, fmt.Sprintf("%v:%v", viper.GetString("listen.host"), viper.GetInt("listen.port")))
	if err != nil {
		logger.Fatal("Fatal error couldn't run", zap.Error(err))
	}

	// Program Manager
	pmErrC, err := ProgramManager(quit)
	if err != nil {
		logger.Fatal("Fatal error couldn't create", zap.Error(err))
	}

	waitfor := 2

	// Wait for API server to terminate
	for waitfor > 0 {
		select {
		case err := <-sErrC:
			waitfor -= 1
			if err != nil {
				logger.Fatal("Fatal error while running", zap.Error(err))
				goto giveup
			}

			// Wait for Program Manager to terminate
		case err := <-pmErrC:
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
