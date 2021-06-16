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
along with Self-host.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	"os"
	"time"

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

	viper.SetConfigName(os.Getenv("CONFIG_FILENAME"))
	viper.SetConfigType("yaml")

	// How do we handle multiple OS?
	viper.AddConfigPath("/etc/selfhost/")
	viper.AddConfigPath("$HOME/.config/selfhost")
	viper.AddConfigPath(".")

	// Default settings
	viper.SetDefault("rate_control.req_per_hour", 600)
	viper.SetDefault("rate_control.maxburst", 10)
	viper.SetDefault("rate_control.cleanup", 3*time.Minute)

	err = viper.ReadInConfig()
	if err != nil {
		logger.Fatal("Fatal error config file", zap.Error(err))
	}

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
