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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/oliveagle/jsonpath"
)

const (
	// StopService is the signal to stop the running querying service
	StopService = iota

	// AddService adds a service to a running cynic instance
	AddService = iota

	// DeleteService removes a service from a running cynic instance
	DeleteService = iota
)

// Session is the configuration a cynic instance requires to start
// running and working
type Session struct {
	StatusPort string
	SlackHook  *string
}

// Service is some http service location that should be queried in a
// specified amount of time. Result is then tested against the given
// jsonpath, or hook contracts.
//
// Hooks: it is possible to assign a hook per service. When the
// service successfully fires and finishes, the retrieved information
// will be passed to the hook. The hook can do whatever, and should
// return a encodable JSON structure.
//
// The hook can do whatever with said information. User defined things
// happen in the hook, and the hook returns a structure ready to be
// encoded into a JSON object, inserted in the sync map, and then
// served back to the client whenever queried.
//
// - A service is an HTTP endpoint
// - A service can have many:
//   - jsonpath contracts
//   - hooks (that can act as contracts)
type Service struct {
	URL       url.URL
	Secs      int
	Contracts []JSONPathSpec
	Hooks     []interface{}
	ticker    *time.Ticker
	running   bool
}

type serviceError struct {
	Error string `json:"error"`
}

// AddressBook contains all the required services inside a map.
type AddressBook struct {
	entries      map[string]Service
	statusServer StatusServer
	Mutex        sync.Mutex
}

// Config is configuration that can boostrap a cynic instance, in
// json format.
type Config struct {
	URL       string   `json:"url"`
	Secs      int      `json:"secs"`
	Contracts []string `json:"contracts"`
}

// AddressBookNew creates a new address book
func AddressBookNew(session Session) AddressBook {
	entries := make(map[string]Service)
	statusServer := StatusServerNew(session.StatusPort, session.SlackHook)
	return AddressBook{entries, statusServer, sync.Mutex{}}
}

// FromPath adds entries to an AddressBook, given a path that contains
// json contracts.
func (s *AddressBook) FromPath(maybePath *string) {
	if maybePath == nil {
		log.Print("no config file loaded")
		return
	}

	path := *maybePath
	configs := parseConfig(path)

	for _, entry := range configs {
		contracts := make([]string, 0)
		log.Printf("loaded service query %s, with %d contract(s)", entry.URL, len(entry.Contracts))

		for _, contract := range entry.Contracts {
			contracts = append(contracts, contract)
		}

		s.AddService(entry.URL, entry.Secs, contracts[:])
	}
}

// AddService adds a service by a configuration triad
func (s *AddressBook) AddService(rawurl string, secs int, contracts []string) {
	s.Mutex.Lock()
	ser := makeService(rawurl, secs)
	ser.Contracts = contracts
	s.entries[rawurl] = ser
	s.Mutex.Unlock()
}

// Get gets a reference to the service with the given rawurl.
func (s *AddressBook) Get(rawurl string) (*Service, bool) {
	val, found := s.entries[rawurl]
	return &val, found
}

// NumEntries returns the number of entries in the AddressBook
func (s *AddressBook) NumEntries() int {
	return len(s.entries)
}

// Run will run the address book against given services
func (s *AddressBook) Run(signal chan int) {
	log.Println("starting the query service")
	s.startTickers()

	go func() { s.statusServer.Start() }()

commands:
	for {
		code := <-signal

		// TODO going to leave some of the signals here for now, but will
		// probably remove them in the future
		switch code {
		case StopService:
			log.Print("received stop signal")
			break commands
		case AddService:
			log.Printf("adding service")
			// Go over anything that has not been started already
			s.startTickers()
		case DeleteService:
			log.Printf("removing service")
			// Remove from synced map since we only insert things
			// in the sync map (and deletes would not be updated)
		default:
			log.Printf("signal not supported: %d", code)
		}
	}

	s.stopTickers()
	s.statusServer.Stop()
}

// AddHook attaches a function hook to the service. If a service
// name is provided, it will attach the hook to that service, or
// create a new service with just that hook.
func (s *AddressBook) AddHook(fn interface{}, rawurl string) {
	s.Mutex.Lock()
	if service, ok := s.entries[rawurl]; ok {
		service.Hooks = append(service.Hooks, fn)
		s.entries[rawurl] = service
	} else {
		url, err := url.Parse(rawurl)
		nilOrDie(err, "provided url for hook could not be parsed: ")

		service.URL = *url
		service.Hooks = append(service.Hooks, fn)
		service.Secs = 1 // TODO argh

		s.entries[rawurl] = service
	}
	s.Mutex.Unlock()
}

// DeleteService removes a service completely from an address book
func (s *AddressBook) DeleteService(rawurl string) {
	s.Mutex.Lock()
	if service, ok := s.entries[rawurl]; ok {
		service.ticker.Stop()
		delete(s.entries, rawurl)
	} else {
		log.Print("no such entry to delete", rawurl)
	}
	s.Mutex.Unlock()

	s.statusServer.Delete(rawurl)
}

/* Private */

func (s *AddressBook) startTickers() {
	for _, service := range s.entries {
		if !service.running {
			go func(service Service, status *StatusServer) {
				for range service.ticker.C {
					workerQuery(service, status)
				}
			}(service, &s.statusServer)

			service.running = true
		}
	}
}

func (s *AddressBook) stopTickers() {
	for _, service := range s.entries {
		service.ticker.Stop()
	}
}

// TODO this could probably be a object method instead...
func workerQuery(s Service, t *StatusServer) {
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

	var json EndpointJSON = ParseEndpointJSON(body[:])
	results := applyContracts(&s, &json)
	t.Update(address, results)
}

func parseConfig(path string) []Config {
	contents, err := ioutil.ReadFile(path)
	nilOrDie(err, "problem opening config file")

	var configs []Config
	err2 := json.
		NewDecoder(strings.NewReader(string(contents[:]))).
		Decode(&configs)
	nilOrDie(err2, "problem decoding configuration file")

	return configs
}

func makeService(rawurl string, secs int) Service {
	u, err := url.Parse(rawurl)
	nilOrDie(err, "invalid http endpoint url")
	ticker := time.NewTicker(time.Duration(secs) * time.Second)
	jsonPathContracts := make([]JSONPathSpec, 0)
	hooks := make([]interface{}, 0)
	return Service{*u, secs, jsonPathContracts, hooks, ticker, false}
}

// TODO need refactoring here
func applyContracts(s *Service, json *EndpointJSON) map[string]interface{} {
	results := make(map[string]interface{})

	type result struct {
		ContractResults interface{}            `json:"contracts"`
		HookResults     map[string]interface{} `json:"hooks"`
		Timestamp       int64                  `json:"timestamp"`
		HumanTime       string                 `json:"human_time"`
	}

	// apply jsonpath contracts
	for _, contract := range s.Contracts {
		if res, err := jsonpath.JsonPathLookup(*json, contract); err != nil {
			log.Println("problem with jsonpath: ", contract, ": ", err)
		} else {
			results[contract] = result{
				ContractResults: res,
				HookResults:     nil,
				Timestamp:       time.Now().Unix(),
				HumanTime:       time.Now().Format(time.RFC850),
			}
		}
	}

	// apply hook contracts
	for i := 0; i < len(s.Hooks); i++ {
		hookName := runtime.FuncForPC(reflect.ValueOf(s.Hooks[i]).Pointer()).Name()
		hookRet := s.Hooks[i].(func(interface{}) interface{})(json) // poetry

		if res, ok := results[hookName]; ok {
			tempResult := res.(result)
			tempResult.HookResults[hookName] = hookRet
			tempResult.Timestamp = time.Now().Unix()

			results[hookName] = tempResult
		} else {
			m := make(map[string]interface{})
			m[hookName] = hookRet

			results[hookName] = result{
				ContractResults: nil,
				HookResults:     m,
				Timestamp:       time.Now().Unix(),
				HumanTime:       time.Now().Format(time.RFC850),
			}
		}
	}

	return results
}
