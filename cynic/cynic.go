/*
Copyright 2018 Simon Symeonidis (psyomn)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/psyomn/cynic"
)

var (
	configFile  string
	config      *string
	statusPort  = cynic.StatusPort
	slackHook   string
	sh          *string
	emailAlerts = false
	version     = false
	help        = false
	logPath     string
)

func initFlag() {
	// General
	flag.StringVar(&configFile, "config", configFile, "cynic config location")
	flag.StringVar(&statusPort, "status-port", statusPort, "http status server port")
	flag.StringVar(&logPath, "log", logPath, "path to log file")

	// Alerts
	flag.StringVar(&slackHook, "slack-hook", slackHook, "set slack hook url")
	flag.BoolVar(&emailAlerts, "email-alerts", emailAlerts, "enable email alerts")

	// Misc
	flag.BoolVar(&version, "v", version, "print the version")
	flag.BoolVar(&help, "h", help, "print this menu")
}

func printVersion() {
	fmt.Fprintf(os.Stderr, "cynic %s\n", cynic.VERSION)
}

func usage() {
	flag.Usage()
}

func handleLog(logPath string) {
	if logPath == "" {
		return
	}

	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
}

func handleConfig(configFile string) *string {
	var conf *string
	if configFile == "" {
		conf = nil
		log.Print("no config loaded")
	} else {
		conf = &configFile
		log.Print("config: ", configFile)
	}
	return conf
}

func handleSlackHook(slackHook string) *string {
	var sh *string
	if slackHook == "" {
		sh = nil
	} else {
		sh = &slackHook
	}
	log.Print("slack hook: ", slackHook)
	return sh
}

func main() {
	// config
	initFlag()
	flag.Parse()

	if version {
		printVersion()
		os.Exit(1)
	}

	if help {
		usage()
		os.Exit(1)
	}

	handleLog(logPath)

	config = handleConfig(configFile)
	sh = handleSlackHook(slackHook)

	log.Printf("status-port: %s\n", statusPort)

	session := cynic.Session{
		Config:     config,
		StatusPort: statusPort,
		SlackHook:  sh,
		Hooks:      nil,
	}

	cynic.Start(session)
}
