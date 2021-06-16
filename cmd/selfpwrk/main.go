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
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/self-host/self-host/api/selfpwrk"
	"github.com/self-host/self-host/api/selfpwrk/library"
)

var logger *zap.Logger

var subscriberUuid uuid.UUID

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

	viper.SetDefault("module_library.scheme", "http")
	viper.SetDefault("module_library.authority", "127.0.0.1:8097")
	viper.SetDefault("program_manager.scheme", "http")
	viper.SetDefault("program_manager.authority", "127.0.0.1:8097")

	viper.SetDefault("cache.library_timeout", 0) // No cache
	viper.SetDefault("cache.program_timeout", 0) // No cache

	err = viper.ReadInConfig()
	if err != nil {
		logger.Fatal("Fatal error config file", zap.Error(err))
	}

	subscriberUuid, err = uuid.NewRandom()
	if err != nil {
		logger.Fatal("Fatal error", zap.Error(err))
	}
}

type SubscribeRequest struct {
	Uuid      uuid.UUID `json:"uuid"`
	Scheme    string    `json:"scheme"`
	Authority string    `json:"authority"`
	Languages []string  `json:"languages"`
}

type UpdateLoadRequest struct {
	Load int64 `json:"load"`
}

// Get preferred outbound ip of this machine
func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("tcp", viper.GetString("program_manager.authority"))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.TCPAddr)

	return localAddr.IP, nil
}

func CheckSubscribed() (bool, error) {
	uri := fmt.Sprintf("%v://%v/v1/subscribers/%v",
		viper.GetString("program_manager.scheme"),
		viper.GetString("program_manager.authority"),
		subscriberUuid.String())

	resp, err := http.Get(uri)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return false, nil
	}

	return true, nil
}

func Subscribe(ip string) error {
	requestBody, err := json.Marshal(SubscribeRequest{
		Uuid:   subscriberUuid,
		Scheme: "http",
		Authority: fmt.Sprintf(
			"%v:%v@%v:%v",
			randomUser,
			randomPass,
			ip,
			viper.GetInt("listen.port"),
		),
		Languages: []string{"tengo"},
	})
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("%v://%v/v1/subscribers",
		viper.GetString("program_manager.scheme"),
		viper.GetString("program_manager.authority"))

	resp, err := http.Post(uri, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("program manager responded with code: %d", resp.StatusCode)
	}

	return nil
}

func ReportLoad(load int64) error {
	requestBody, err := json.Marshal(UpdateLoadRequest{
		Load: load,
	})
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("%v://%v/v1/subscribers/%v/load",
		viper.GetString("program_manager.scheme"),
		viper.GetString("program_manager.authority"),
		subscriberUuid.String())

	req, err := http.NewRequest("PUT", uri, bytes.NewReader(requestBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return fmt.Errorf("unexpected response from program manager: %v", resp.StatusCode)
	}

	return nil
}

func Unsubscribe() (bool, error) {
	uri := fmt.Sprintf("%v://%v/v1/subscribers/%v",
		viper.GetString("program_manager.scheme"),
		viper.GetString("program_manager.authority"),
		subscriberUuid.String())

	reqURL, err := url.Parse(uri)
	if err != nil {
		return false, err
	}

	req := &http.Request{
		Method: "DELETE",
		URL:    reqURL,
		Header: map[string][]string{},
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return false, nil
	}

	return true, nil
}

func main() {
	selfpwrk.SetCacheTimeout(viper.GetInt("cache.program_timeout"))
	library.SetCacheTimeout(viper.GetInt("cache.library_timeout"))

	uri := fmt.Sprintf("%v://%v",
		viper.GetString("module_library.scheme"),
		viper.GetString("module_library.authority"))
	library.SetIndexServer(uri)

	// Run a background task to ensure we are subscribed
	go func() {
		var outboundip string

		for {
			select {
			// Every 10s check if we are subscribed, and if not subscribe

			case <-time.After(10 * time.Second):
				if outboundip == "" {
					ip, err := GetOutboundIP()
					if err != nil {
						logger.Error("Couldn't get outbound IP", zap.Error(err))
					} else {
						outboundip = ip.String()
					}
				}

				// Unless
				if outboundip != "" {
					ok, err := CheckSubscribed()
					if err != nil {
						logger.Error("Couldn't check subscription", zap.Error(err))
					} else if ok == false {
						Subscribe(outboundip)
					} else {
						err = ReportLoad(selfpwrk.ProgramCacheGetLoad())
						if err != nil {
							logger.Error("unable to report load", zap.Error(err))
						}
					}
				}
			}
		}
	}()

	errC, err := Server(fmt.Sprintf("%v:%v", viper.GetString("listen.host"), viper.GetInt("listen.port")))
	if err != nil {
		logger.Fatal("Fatal error couldn't run", zap.Error(err))
	}

	if err := <-errC; err != nil {
		logger.Fatal("Fatal error while running", zap.Error(err))
	}

	ok, err := Unsubscribe()
	if err != nil {
		logger.Error("Error while unsubscribing", zap.Error(err))
	} else if ok == false {
		logger.Error("Failed to unsubscribe", zap.Error(err))
	}
}
