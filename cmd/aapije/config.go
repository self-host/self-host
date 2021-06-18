// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

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
