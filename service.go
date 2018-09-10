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
	"log"
	"net/url"
	"time"
)

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
	hooks      []HookSignature
	ticker     *time.Ticker
	running    bool
	tickerChan chan int
	immediate  bool
	offset     int
}

// TODO make sure that this is used everywhere.
type serviceError struct {
	Error string `json:"error"`
}

// ServiceNew creates a new service instance
func ServiceNew(rawurl string, secs int) Service {
	u, err := url.Parse(rawurl)
	nilOrDie(err, "invalid http endpoint url")
	ticker := time.NewTicker(time.Duration(secs) * time.Second)
	hooks := make([]HookSignature, 0)
	tchan := make(chan int)

	return Service{
		URL:        *u,
		Secs:       secs,
		hooks:      hooks,
		ticker:     ticker,
		running:    false,
		tickerChan: tchan,
		immediate:  false,
		offset:     0,
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
func (s *Service) AddHook(fn HookSignature) {
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

// Offset sets the time before the service starts ticking
func (s *Service) Offset(offset int) {
	s.offset = offset
}
