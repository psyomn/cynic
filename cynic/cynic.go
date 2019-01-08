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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

func simpleGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("problem accessing url: ", url)
		return "", err
	}
	defer resp.Body.Close()

	bytesResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	str := string(bytesResp)

	if resp.StatusCode != 200 {
		log.Println(resp.StatusCode, " ", http.StatusText(resp.StatusCode))
		return "", errors.New(str)
	}

	return str, nil
}

// You need to respect this interface so that you can bind hooks to
// your events. You can return a struct with json hints as shown
// bellow, and cynic will add that to the /status endpoint.
func exampleHook(params *cynic.HookParameters) (alert bool, data interface{}) {
	fmt.Println("executing exampleHook!")

	message := ""
	url := "http://localhost:9001/one"
	resp, err := simpleGet(url)

	if err != nil {
		message = err.Error()
		params.Status.Update("exampleHook", message)
		return false, result{
			Alert:   true,
			Message: message,
		}
	}

	message = resp

	var target interface{}
	json.Unmarshal([]byte(message), &target)
	params.Status.Update("exampleHook", target)

	return false, result{
		Alert:   true,
		Message: message,
	}
}

// Another example hook
func anotherExampleHook(params *cynic.HookParameters) (alert bool, data interface{}) {
	fmt.Println("execute anotherExampleHook!")
	url := "http://localhost:9001/two"
	resp, err := simpleGet(url)
	message := ""

	if err != nil {
		message = err.Error()
	} else {
		message = resp
	}

	var target interface{}
	json.Unmarshal([]byte(message), &target)
	params.Status.Update("anotherExampleHook", target)

	return false, result{
		Alert:   true,
		Message: message,
	}
}

func finalHook(params *cynic.HookParameters) (alert bool, data interface{}) {
	fmt.Println("execute finalHook!")
	url := "http://localhost:9001/flappyerror"

	message := ""
	resp, err := simpleGet(url)
	if err != nil {
		message = err.Error()
		params.Status.Update("finalHook", message)

		return (time.Now().Unix()%2 == 0), result{
			Alert:   false,
			Message: message,
		}
	}

	message = resp

	var target interface{}
	json.Unmarshal([]byte(message), &target)
	params.Status.Update("finalHook", target)

	return (time.Now().Unix()%2 == 0), result{
		Alert:   false,
		Message: message,
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

	var events []cynic.Event

	events = append(events, cynic.EventNew(1))
	events = append(events, cynic.EventNew(2))
	events = append(events, cynic.EventNew(3))

	events[0].AddHook(anotherExampleHook)
	events[0].SetOffset(10) // delay 10 seconds before starting
	events[0].Repeat(true)

	events[1].AddHook(exampleHook)
	events[1].Repeat(true)

	events[2].AddHook(finalHook)
	events[2].Repeat(true)

	statusServer := cynic.StatusServerNew(statusPort, cynic.DefaultStatusEndpoint)

	for i := 0; i < len(events); i++ {
		events[i].SetDataRepo(&statusServer)
	}

	alerter := cynic.AlerterNew(20, exampleAlerter)
	session := cynic.Session{
		Events:      events,
		Alerter:     &alerter,
		StatusCache: &statusServer,
		SnapshotConfig: &cynic.SnapshotConfig{
			Enabled:   true,
			Interval:  time.Minute,
			DumpEvery: time.Hour,
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	cynic.Start(session)
	wg.Wait()
}
