/*
This is an example, on how you could deploy a cynic instance.

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

// This is to show that you can have a simple alerter, if something is
// detected to be awry in the monitoring.
func exampleAlerter(_ []cynic.AlertMessage) {
	fmt.Println("##################################")
	fmt.Println("# Hey you! Better pay attention! #")
	fmt.Println("##################################")
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

type result struct {
	Alert   bool   `json:"alert"`
	Message string `json:"message"`
}

// You need to respect this interface so that you can bind hooks to
// your services. You can return a struct with json hints as shown
// bellow, and cynic will add that to the /status endpoint.
func exampleHook(s *cynic.AddressBook, resp interface{}) (alert bool, data interface{}) {
	fmt.Println("Firing the example hook yay!")
	return false, result{
		Alert:   true,
		Message: "AARRRGGHHHHHH",
	}
}

// Another example hook
func anotherExampleHook(c *cynic.AddressBook, resp interface{}) (alert bool, data interface{}) {
	fmt.Println("Firing example hook 2 yay!")
	return false, result{
		Alert:   true,
		Message: "I feel calm and collected inside.",
	}
}

func finalHook(c *cynic.AddressBook, resp interface{}) (alert bool, data interface{}) {
	fmt.Println("IT'S THE FINAL HOOKDOWN")
	return false, result{
		Alert:   false,
		Message: "I feel calm and collected inside.",
	}
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

	handleLog(logPath)

	sh = handleSlackHook(slackHook)

	log.Printf("status-port: %s\n", statusPort)

	var services []cynic.Service
	services = append(services, cynic.ServiceNew("http://localhost:9001/one", 1))
	services = append(services, cynic.ServiceNew("http://localhost:9001/two", 1))
	services = append(services, cynic.ServiceNew("http://localhost:9001/flappyerror", 1))

	services[0].AddHook(exampleHook)
	services[0].AddHook(anotherExampleHook)
	services[0].AddHook(finalHook)
	services[0].Offset(10) // delay 10 seconds before starting

	services[1].AddHook(exampleHook)

	services[2].AddHook(exampleHook)
	services[2].AddHook(anotherExampleHook)
	services[2].AddHook(finalHook)

	for i := 0; i < len(services); i++ {
		fmt.Println("entry: ", services[i])
		fmt.Printf("address: %p\n", &services[i])
	}

	fmt.Println("passing to session: ", services)

	session := cynic.Session{
		StatusPort: statusPort,
		SlackHook:  sh,
		Services:   services,
		Alerter:    exampleAlerter,
		AlertTime:  20, // check status every 20 seconds
	}

	cynic.Start(session)
}
