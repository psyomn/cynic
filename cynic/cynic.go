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
	"sync"
	"time"

	"github.com/psyomn/cynic"
)

var (
	statusPort = cynic.StatusPort
	slackHook  string
	version    = false
	help       = false
	logPath    string
)

func initFlag() {
	// General
	flag.StringVar(&statusPort, "status-port", statusPort, "http status server port")
	flag.StringVar(&logPath, "log", logPath, "path to log file")

	// Alerts
	flag.StringVar(&slackHook, "slack-hook", slackHook, "set slack hook url")

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
func exampleAlerter(messages []cynic.AlertMessage) {
	fmt.Println("############################################")
	fmt.Println("# Hey you! Better pay attention!            ")
	fmt.Println("############################################")
	fmt.Println("# messages: ")

	for ix, el := range messages {
		fmt.Println("# ", ix)
		fmt.Println("#  response: ", el.Response)
		fmt.Println("#  now     : ", el.Now)
		fmt.Println("#  cynichos: ", el.CynicHostname)
		fmt.Println("#        ##########################")
	}

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

type result struct {
	Alert   bool   `json:"alert"`
	Message string `json:"message"`
}

// You need to respect this interface so that you can bind hooks to
// your events. You can return a struct with json hints as shown
// bellow, and cynic will add that to the /status endpoint.
func exampleHook(_ *cynic.HookParameters) (alert bool, data interface{}) {
	fmt.Println("Firing the example hook yay!")
	return false, result{
		Alert:   true,
		Message: "AARRRGGHHHHHH",
	}
}

var exHook2Cnt int

// Another example hook
func anotherExampleHook(_ *cynic.HookParameters) (alert bool, data interface{}) {
	fmt.Println("Firing example hook 2 yay!")
	fmt.Println("exhook2Cnt: ", exHook2Cnt)
	exHook2Cnt++

	return false, result{
		Alert:   true,
		Message: "I feel calm and collected inside.",
	}
}

func finalHook(_ *cynic.HookParameters) (alert bool, data interface{}) {
	fmt.Println("IT'S THE FINAL HOOKDOWN")
	return (time.Now().Unix()%2 == 0), result{
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

	log.Printf("status-port: %s\n", statusPort)

	var events []cynic.Event

	events = append(events, cynic.EventNew(1)) // "http://localhost:9001/one",
	events = append(events, cynic.EventNew(2)) // "http://localhost:9001/two",
	events = append(events, cynic.EventNew(3)) // "http://localhost:9001/flappyerror",

	events[0].AddHook(exampleHook)
	events[0].AddHook(anotherExampleHook)
	events[0].AddHook(finalHook)
	events[0].Offset(10) // delay 10 seconds before starting
	events[0].Repeat(true)

	events[1].AddHook(exampleHook)
	events[1].Repeat(true)

	events[2].AddHook(exampleHook)
	events[2].AddHook(anotherExampleHook)
	events[2].AddHook(finalHook)
	events[2].Repeat(true)

	for i := 0; i < len(events); i++ {
		fmt.Println("entry: ", events[i])
		fmt.Printf("address: %p\n", &events[i])
	}

	var statusServers []cynic.StatusServer
	statusServer := cynic.StatusServerNew(statusPort, cynic.DefaultStatusEndpoint)
	statusServers = append(statusServers, statusServer)

	for i := 0; i < len(events); i++ {
		events[i].DataRepo(&statusServer)
	}

	alerter := cynic.AlerterNew(20, exampleAlerter)
	session := cynic.Session{
		Events:        events,
		Alerter:       &alerter,
		StatusServers: statusServers,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	cynic.Start(session)
	wg.Wait()
}
