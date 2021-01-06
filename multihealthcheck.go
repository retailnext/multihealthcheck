// Copyright 2021 RetailNext, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/retailnext/multihealthcheck/checker"
	"github.com/retailnext/multihealthcheck/handler"
)

type config struct {
	ServeAddr    string `json:"serve_addr"`
	HealthChecks map[string]struct {
		URL string `json:"url"`
	} `json:"health_checks"`
}

func (c config) makeChecker() *checker.MultiChecker {
	var mc checker.MultiChecker
	for name, check := range c.HealthChecks {
		mc.AddGetCheck(name, check.URL)
	}
	return &mc
}

func (c config) validateOrExit() {
	if len(c.HealthChecks) == 0 {
		fmt.Println("ERROR: no health_checks defined")
		os.Exit(1)
	}
	if c.ServeAddr == "" {
		fmt.Println("ERROR: serve_addr not defined")
		os.Exit(1)
	}
}

func loadConfigOrExit() config {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s config.json\n", os.Args[0])
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			panic(closeErr)
		}
	}()
	var cfg config
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	cfg.validateOrExit()
	return cfg
}

func runWithInterrupt(server *http.Server) {
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	defer signal.Stop(signalCh)

	<-signalCh

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	cfg := loadConfigOrExit()
	mc := cfg.makeChecker()
	mux := handler.NewMux(mc)
	server := &http.Server{
		Addr:    cfg.ServeAddr,
		Handler: mux,
	}
	runWithInterrupt(server)
}
