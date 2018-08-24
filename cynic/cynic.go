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
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/psyomn/cynic"
)

var (
	configFile  = "/etc/cynic/cynic.conf"
	statusPort  = cynic.StatusPort
	slackHook   string
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

func main() {
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

	var sh *string
	if slackHook == "" {
		sh = nil
	} else {
		sh = &slackHook
	}

	if logPath != "" {
		file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}

		defer file.Close()
		log.SetOutput(file)
	}

	log.Printf("config: %s\n", configFile)
	log.Printf("status-port: %s\n", statusPort)
	log.Printf("slack hook %s\n", slackHook)

	session := cynic.Session{StatusPort: statusPort, SlackHook: sh}

	addressBook := cynic.AddressBookNew(session)
	addressBook.FromPath(configFile)

	signal := make(chan int)
	go func() { addressBook.RunQueryService(signal) }()

	for {
		// TODO might trash this in the future.
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		text, err := reader.ReadString('\n')

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			continue
		} else if strings.TrimRight(text, "\n") == "stop" {
			fmt.Println("sending exit signal...")
			signal <- cynic.StopService
			return
		}
	}
}
