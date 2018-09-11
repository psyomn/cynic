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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	// StopService is the signal to stop the running querying service
	StopService = iota

	// AddService adds a service to a running cynic instance
	AddService

	// DeleteService removes a service from a running cynic instance
	DeleteService
)

// AlertFunc defines the hook signature for alert messages
type AlertFunc = func([]AlertMessage)

// AlertMessage defines a simple alert structure that can be used by
// users of the library, and decide how to show information about the
// alerts.
type AlertMessage struct {
	Response      interface{} `json:"response_text"`
	Endpoint      string      `json:"endpoint"`
	Now           string      `json:"now"`
	CynicHostname string      `json:"cynic_hostname"`
}

// Session is the configuration a cynic instance requires to start
// running and working
type Session struct {
	StatusPort     string
	StatusEndpoint string
	Services       []Service
	Alerter        AlertFunc
	AlertTime      int
}

// HookSignature specifies what the service hooks should look like.
type HookSignature = func(*AddressBook, interface{}) (bool, interface{})

// Start starts a cynic instance, with any provided hooks.
func Start(session Session) {
	addressBook := AddressBookNew(session)
	signal := make(chan int)
	addressBook.Run(signal)
}

// TODO this could probably be a object method instead...
func workerQuery(addressBook *AddressBook, s *Service, t *StatusServer) {
	address := s.URL.String()

	resp, err := http.Get(address)
	if err != nil {
		message := "problem getting response"
		nilAndOk(err, message)
		t.Update(address, serviceError{Error: message})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buff := fmt.Sprintf("got non 200 code: %d", resp.StatusCode)
		t.Update(address, serviceError{Error: buff})
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		message := "problem reading data from endpoint"
		nilAndOk(err, message)
		t.Update(address, serviceError{Error: message})
		return
	}

	var json EndpointJSON = parseEndpointJSON(body[:])

	results := applyContracts(addressBook, s, &json)
	t.Update(address, results)
}

func applyContracts(addressBook *AddressBook, s *Service, json *EndpointJSON) interface{} {
	type result struct {
		HookResults map[string]interface{} `json:"hooks"`
		Timestamp   int64                  `json:"timestamp"`
		HumanTime   string                 `json:"human_time"`
		Alert       bool                   `json:"alert"`
	}

	var ret result
	sumAlerts := false
	ret.HookResults = make(map[string]interface{})

	for i := 0; i < len(s.hooks); i++ {
		hookName := getFuncName(s.hooks[i])
		retAlert, hookRet := s.hooks[i](addressBook, *json)
		sumAlerts = sumAlerts || retAlert

		if res, ok := ret.HookResults[hookName]; ok {
			tempResult := res.(result)
			tempResult.HookResults[hookName] = hookRet
			tempResult.Timestamp = time.Now().Unix()
			ret.HookResults[hookName] = tempResult
			ret.Alert = retAlert
		} else {
			ret.HookResults[hookName] = hookRet
			ret.Timestamp = time.Now().Unix()
			ret.HumanTime = time.Now().Format(time.RFC850)
			ret.Alert = retAlert
		}
	}

	if sumAlerts {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "nohost"
		}
		message := AlertMessage{
			Endpoint:      s.URL.String(),
			Response:      ret,
			CynicHostname: hostname,
			Now:           time.Now().Format(time.RFC850),
		}
		addressBook.queueAlert(&message)
	}

	return ret
}
