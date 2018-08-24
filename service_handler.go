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
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
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
	Services   []Service
}

// Service is some http service location that should be queried in a
// specified amount of time.
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
//   - hooks (that can act as contracts)
type Service struct {
	URL        url.URL
	Secs       int
	hooks      []interface{}
	ticker     *time.Ticker
	running    bool
	tickerChan chan int
	immediate  bool
}

// TODO make sure that this is used everywhere.
type serviceError struct {
	Error string `json:"error"`
}

// AddressBook contains all the required services inside a map.
type AddressBook struct {
	entries      map[string]*Service
	statusServer StatusServer
	Mutex        *sync.Mutex
}

// Start starts a cynic instance, with any provided hooks.
func Start(session Session) {
	addressBook := AddressBookNew(session)
	signal := make(chan int)
	addressBook.Run(signal)
}

// AddressBookNew creates a new address book
func AddressBookNew(session Session) *AddressBook {
	entries := make(map[string]*Service)
	statusServer := StatusServerNew(session.StatusPort, session.SlackHook)
	addressBook := AddressBook{entries, statusServer, &sync.Mutex{}}

	for i := 0; i < len(session.Services); i++ {
		addressBook.AddService(&session.Services[i])
	}

	return &addressBook
}

// AddService adds a service
func (s *AddressBook) AddService(service *Service) {
	s.Mutex.Lock()
	rawurl := service.URL.String()
	if entry, ok := s.entries[rawurl]; ok {
		if entry.running {
			entry.Stop()
		}
	}
	s.entries[rawurl] = &*service
	s.Mutex.Unlock()
}

// Get gets a reference to the service with the given rawurl.
func (s *AddressBook) Get(rawurl string) (*Service, bool) {
	s.Mutex.Lock()
	val, found := s.entries[rawurl]
	s.Mutex.Unlock()
	return val, found
}

// NumEntries returns the number of entries in the AddressBook
func (s *AddressBook) NumEntries() int {
	return len(s.entries)
}

// Contains checks to see if a url is contained in the service list
func (s *AddressBook) Contains(rawurl string) bool {
	_, ok := s.entries[rawurl]
	return ok
}

// Run will run the address book against given services
func (s *AddressBook) Run(signal chan int) {
	log.Println("starting the query service")
	s.StartTickers()
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
			s.StartTickers()
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

// DeleteService removes a service completely from an address book. It
// is OK to pass non-existant rawurls to delete.
func (s *AddressBook) DeleteService(rawurl string) {
	s.Mutex.Lock()
	if service, ok := s.entries[rawurl]; ok {
		service.Stop()
		delete(s.entries, rawurl)
	} else {
		log.Print("no such entry to delete", rawurl)
	}
	s.Mutex.Unlock()
	s.statusServer.Delete(rawurl)
}

// StartTickers starts the tickers on the associated services. This
// might go away in the future
func (s *AddressBook) StartTickers() {
	s.Mutex.Lock()
	for _, service := range s.entries {
		if service.running {
			continue
		}

		log.Print(service.URL, " is not started, starting.")

		go func(service *Service, status *StatusServer) {
			if !service.running && service.immediate {
				// Force first tick if service is immediate
				workerQuery(s, service, status)
			}
			service.running = true

			for {
				select {
				case <-service.ticker.C:
					workerQuery(s, service, status)
				case <-service.tickerChan:
					return
				}
			}
		}(service, &s.statusServer)
	}
	s.Mutex.Unlock()
}

func (s *AddressBook) stopTickers() {
	s.Mutex.Lock()
	for _, service := range s.entries {
		service.ticker.Stop()
	}
	s.Mutex.Unlock()
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

	var json EndpointJSON = ParseEndpointJSON(body[:])

	results := applyContracts(addressBook, s, &json)
	t.Update(address, results)
}

// ServiceNew creates a new service instance
func ServiceNew(rawurl string, secs int) Service {
	u, err := url.Parse(rawurl)
	nilOrDie(err, "invalid http endpoint url")
	ticker := time.NewTicker(time.Duration(secs) * time.Second)
	hooks := make([]interface{}, 0)
	tchan := make(chan int)

	return Service{
		*u,
		secs,
		hooks,
		ticker,
		false,
		tchan,
		false,
	}
}

// Stop service will stop the ticker, and gracefully exit it.
func (s *Service) Stop() {
	log.Print("stopping service: ", s.URL.String())
	s.ticker.Stop()
	s.tickerChan <- 0
	close(s.tickerChan)
}

// AddHook appends a hook to the service
func (s *Service) AddHook(fn interface{}) {
	s.hooks = append(s.hooks, fn)
}

// NumHooks counts the hooks
func (s *Service) NumHooks() int {
	return len(s.hooks)
}

// Immediate will make the service run immediately
func (s *Service) Immediate() {
	s.immediate = true
}

func applyContracts(addressBook *AddressBook, s *Service, json *EndpointJSON) interface{} {
	type result struct {
		HookResults map[string]interface{} `json:"hooks"`
		Timestamp   int64                  `json:"timestamp"`
		HumanTime   string                 `json:"human_time"`
		Alert       bool                   `json:"alert"`
	}

	var ret result
	ret.HookResults = make(map[string]interface{})

	for i := 0; i < len(s.hooks); i++ {
		hookName := getFuncName(s.hooks[i])
		hookRet := s.hooks[i].(func(*AddressBook, interface{}) interface{})(addressBook, *json) // poetry

		if res, ok := ret.HookResults[hookName]; ok {
			tempResult := res.(result)
			tempResult.HookResults[hookName] = hookRet
			tempResult.Timestamp = time.Now().Unix()
			ret.HookResults[hookName] = tempResult
		} else {
			ret.HookResults[hookName] = hookRet
			ret.Timestamp = time.Now().Unix()
			ret.HumanTime = time.Now().Format(time.RFC850)
			ret.Alert = false
		}
	}

	return ret
}
