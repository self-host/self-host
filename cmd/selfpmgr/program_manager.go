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
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/self-host/self-host/api/selfpmgr"
	"github.com/self-host/self-host/pkg/util"
	pg "github.com/self-host/self-host/postgres"
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
				err := pg.AddDB(domain, pguri)
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
			domains := pg.GetDomains()
			for domain := range v.GetStringMapString("domains") {
				index := util.StringSliceIndex(domains, domain)
				if index == -1 || len(domains) == 0 {
					continue
				} else if len(domains) == 1 {
					// Absolue last element in the slice
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
				pg.RemoveDB(domain)
			}

			// Add new/existing domain DBs
			for domain, pguri := range v.GetStringMapString("domains") {
				err := pg.AddDB(domain, pguri)
				if err != nil {
					logger.Error("Error while adding domain", zap.Error(err))
				}
			}
		})
	}

	go func() {
		<-quit

		/*
		   ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		   defer func() {
		           stop()
		           cancel()
		           close(errC)
		   }()
		*/
	}()

	go func() {
		selfpmgr.UpdateProgramCache()
		for {
			select {
			case <-util.AtInterval(1 * time.Minute):
				selfpmgr.UpdateProgramCache()
			}
		}
	}()

	return errC, nil
}
