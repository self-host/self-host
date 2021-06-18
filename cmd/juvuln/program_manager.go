// Copyright 2021 The Self-host Authors. All rights reserved.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package main

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/self-host/self-host/api/juvuln"
	"github.com/self-host/self-host/pkg/util"
	"github.com/self-host/self-host/pkg/workforce"
	"github.com/self-host/self-host/postgres"
)

func ProgramManager(quit <-chan struct{}) (<-chan error, error) {
	errC := make(chan error, 1)

	domainfile := viper.GetString("domainfile")
	if domainfile != "" {
		v := viper.New()
		v.SetConfigName(domainfile)
		v.SetConfigType("yaml")
		v.AddConfigPath("/etc/selfhost/")
		v.AddConfigPath("$HOME/.config/selfhost")
		v.AddConfigPath(".")

		err := v.ReadInConfig()
		if err != nil {
			errC <- err
		}

		if v.IsSet("domains") {
			for domain, pguri := range v.GetStringMapString("domains") {
				err := postgres.AddDB(domain, pguri)
				if err != nil {
					errC <- err
				}
			}
		}

		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			err := v.ReadInConfig()
			if err != nil {
				errC <- err
			}

			// Find inactive databases
			domains := postgres.GetDomains()
			for domain := range v.GetStringMapString("domains") {
				index := util.StringSliceIndex(domains, domain)
				if index == -1 || len(domains) == 0 {
					continue
				} else if len(domains) == 1 {
					// Absolute last element in the slice
					domains = make([]string, 0)
				} else {
					// Place last element at position
					domains[index] = domains[len(domains)-1]
					// "delete" last element
					domains[len(domains)-1] = ""
					// Truncate slice
					domains = domains[:len(domains)-1]
				}
			}

			// What remains in "domains" is all domains no longer active in config file
			for _, domain := range domains {
				postgres.RemoveDB(domain)
			}

			// Add new/existing domain DBs
			for domain, pguri := range v.GetStringMapString("domains") {
				err := postgres.AddDB(domain, pguri)
				if err != nil {
					logger.Error("Error while adding domain", zap.Error(err))
				}
			}
		})
	}

	go func() {
		juvuln.UpdateProgramCache()

		for {
			every5s := util.AtInterval(5 * time.Second)
			every1m := util.AtInterval(1 * time.Minute)

			select {
			case <-every1m:
				juvuln.UpdateProgramCache()
			case <-every5s:
				rejected := workforce.ClearInactive()
				for _, obj := range rejected {
					w, ok := obj.(*juvuln.Worker)
					if ok {
						logger.Info("fired worker", zap.String("id", w.Id))
					}
				}
			case <-quit:
				return
			}
		}
	}()

	return errC, nil
}
