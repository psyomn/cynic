/*
Package cynic monitors you from the ceiling.

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
package cynic

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	configFile  string
	statusPort  = StatusPort
	slackHook   string
	emailAlerts = false
	version     = false
	help        = false
	logPath     string
)

// ServiceHooks are the hooks you may want to provide
type ServiceHooks struct {
	RawURL string
	Hooks  []interface{}
}

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
	fmt.Fprintf(os.Stderr, "cynic %s\n", VERSION)
}

func usage() {
	flag.Usage()
}

func handleLog(logPath string) {
	if logPath != "" {
		file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)

		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(file)
	}
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

// StartWithHooks starts a cynic instance, with any provided hooks.
func StartWithHooks(givenHooks []ServiceHooks) {
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

	config := handleConfig(configFile)
	sh := handleSlackHook(slackHook)

	// Execute
	log.Printf("status-port: %s\n", statusPort)

	session := Session{StatusPort: statusPort, SlackHook: sh}

	addressBook := AddressBookNew(session)
	addressBook.FromPath(config)

	if len(givenHooks) != 0 {
		log.Print("adding custom hooks to services")
	}

	for _, entry := range givenHooks {
		for _, hook := range entry.Hooks {
			addressBook.AddHook(hook, entry.RawURL)
		}
	}

	signal := make(chan int)
	go func() { addressBook.Run(signal) }()

	for {
		// TODO might trash this in the future.
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")

		text, err := reader.ReadString('\n')

		if err != nil {
			log.Println("error: ", err)
			continue
		} else {
			cmd := strings.TrimRight(text, "\n")

			switch cmd {
			case "stop":
				log.Println("sending exit signal...")
				signal <- StopService
				return
			case "add service":
				handleAddService(&addressBook, reader)
				signal <- AddService
			case "count":
				handleCount(&addressBook)
			case "delete service":
				handleDeleteService(&addressBook, reader)
				signal <- DeleteService
			case "help":
				fmt.Println("current commands: ")
				fmt.Println("stop - stop cynic instance")
				fmt.Println("add service - add a service to cynic")
				fmt.Println("delete service - delete a service")
			}
		}
	}
}

func handleAddService(book *AddressBook, reader *bufio.Reader) {
	log.Println("adding service...")

getURL:
	fmt.Print("url of service: ")
	_url, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
		goto getURL
	}
	url := strings.TrimRight(_url, "\n")

getSecs:
	fmt.Print("secs of service: ")
	_secs, err := reader.ReadString('\n')
	if err != nil {
		fmt.Print(err)
		goto getSecs
	}

	secs, err := strconv.Atoi(strings.TrimRight(_secs, "\n"))
	if err != nil {
		fmt.Print(err)
		goto getSecs
	}

getContract:
	// Only care for one contract for now
	fmt.Print("jsonpath contract: ")
	_contract, err := reader.ReadString('\n')
	if err != nil {
		fmt.Print(err)
		goto getContract
	}

	contract := strings.TrimRight(_contract, "\n")
	contracts := make([]string, 1)
	contracts[0] = contract
	book.AddService(url, secs, contracts)
}

func handleCount(book *AddressBook) {
	log.Println("num of entries: ", book.NumEntries())
}

func handleDeleteService(book *AddressBook, reader *bufio.Reader) {
	log.Println("deleting service...")
read:
	_text, err := reader.ReadString('\n')
	if err != nil {
		fmt.Print(err)
		goto read
	}

	text := strings.TrimRight(_text, "\n")
	book.DeleteService(text)
}
