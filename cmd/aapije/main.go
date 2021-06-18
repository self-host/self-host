// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	initLogger()
	initConfig()
}

func main() {
	errC, err := Server(fmt.Sprintf("%v:%v", viper.GetString("listen.host"), viper.GetInt("listen.port")))
	if err != nil {
		logger.Fatal("Fatal error couldn't run", zap.Error(err))
	}

	if err := <-errC; err != nil {
		logger.Fatal("Fatal error while running", zap.Error(err))
	}
}
