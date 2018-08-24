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
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/oliveagle/jsonpath"
)

const (
	// StopService is the signal to stop the running querying service
	StopService = iota
)

// Session is the configuration a cynic instance requires to start
// running and working
type Session struct {
	StatusPort string
	SlackHook  *string
}

// Service is some http service location that should be queried in a
// specified amount of time. Result is then tested against the given
// jsonpath contracts.
type Service struct {
	URL       url.URL
	Secs      int
	Contracts []JSONPathSpec
}

// AddressBook contains all the required services inside a map.
type AddressBook struct {
	entries      map[string]Service
	statusServer StatusServer
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
	return AddressBook{entries, statusServer}
}

// FromPath adds entries to an AddressBook, given a path that contains
// json contracts.
func (s *AddressBook) FromPath(path string) {
	configs := parseConfig(path)

	for _, entry := range configs {
		contracts := make([]string, 0)
		log.Printf("loaded service query %s, with %d contract(s)", entry.URL, len(entry.Contracts))

		for _, contract := range entry.Contracts {
			contracts = append(contracts, contract)
		}

		s.Add(entry.URL, entry.Secs, contracts[:])
	}
}

// Add adds a service by a configuration triad
func (s *AddressBook) Add(rawurl string, secs int, contracts []string) {
	ser := makeService(rawurl, secs)
	ser.Contracts = contracts
	s.entries[rawurl] = ser
}

// NumEntries returns the number of entries in the AddressBook
func (s *AddressBook) NumEntries() int {
	return len(s.entries)
}

// RunQueryService will run the address book against given services
func (s *AddressBook) RunQueryService(signal chan int) {
	log.Println("starting the query service")

	tickers := make([]*time.Ticker, 0)

	for _, service := range s.entries {
		// TODO this needs improvements
		ticker := time.NewTicker(time.Duration(service.Secs) * time.Second)
		tickers = append(tickers, ticker)

		go func(service Service, status *StatusServer) {
			for range ticker.C {
				workerQuery(service, status)
			}
		}(service, &s.statusServer)
	}

	go func() { s.statusServer.Start() }()

	for {
		code := <-signal
		log.Print("received stop signal")
		if code != StopService {
			log.Print("StopService signal only currently supported")
		} else {
			break
		}
	}

	for _, ticker := range tickers {
		ticker.Stop()
	}

	s.statusServer.Stop()
}

func workerQuery(s Service, t *StatusServer) {
	address := s.URL.String()

	resp, err := http.Get(address)
	if err != nil {
		message := "problem getting response"
		nilAndOk(err, message)
		t.Update(address, message)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Update(address, "got non 200 code: "+string(resp.StatusCode))
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		message := "problem reading data from endpoint"
		nilAndOk(err, message)
		t.Update(address, message)
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
	return Service{*u, secs, make([]JSONPathSpec, 0)}
}

func applyContracts(s *Service, json *EndpointJSON) map[string]interface{} {
	results := make(map[string]interface{})

	for _, contract := range s.Contracts {
		res, err := jsonpath.JsonPathLookup(*json, contract)

		if err != nil {
			log.Println("problem with jsonpath: ", contract)
		} else {
			results[contract] = res
			log.Println("res: ", res)
		}
	}

	return results
}
