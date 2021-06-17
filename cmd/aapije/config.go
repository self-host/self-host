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
	"os"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func initConfig() {
	var err error

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

	// CORS default settings
	viper.SetDefault("cors.allowed_origins", []string{"https://*", "http://*"})
	viper.SetDefault("cors.allowed_methods", []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Accept", "Authorization", "Content-Type", "If-None-Match"})
	viper.SetDefault("cors.exposed_headers", []string{"Link"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 300) // Maximum value not ignored by any of major browsers

	err = viper.ReadInConfig()
	if err != nil {
		logger.Fatal("Fatal error unable to load config file", zap.Error(err))
	}
}
