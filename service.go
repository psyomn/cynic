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
	"sync/atomic"
	"time"
)

const (
	// ServiceDefault is for services that query in the default
	// way, which is JSON, restful endpoints
	ServiceDefault = iota

	// ServiceCustom is for services that query other endpoints
	// instead of JSON. It is up to the user to implement support
	// for these endpoints.
	ServiceCustom
)

// HookSignature specifies what the service hooks should look like.
type HookSignature = func(*StatusServer) (bool, interface{})

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
// - A service is an action
// - A service can have many:
//   - hooks (that can act as contracts)
// - A service may be bound to a data repository/cache
type Service struct {
	url       *url.URL
	secs      int
	hooks     []HookSignature
	immediate bool
	offset    int
	repeat    bool
	Label     *string
	id        uint64

	repo *StatusServer

	absSecs int // TODO: eventually remove

	alerter *Alerter

	absExpiry int64

	index    int
	priority int
}

var lastID uint64

// ServiceNew creates a new service that is primarily used for pure
// execution
func ServiceNew(secs int) Service {
	if secs <= 0 {
		log.Fatal("NO. GOD. NO. GOD PLEASE NO. NO. NO. NOOOOOOOO.")
	}

	id := atomic.AddUint64(&lastID, 1)

	priority := secs + int(time.Now().Unix())

	return Service{
		url:       nil,
		secs:      secs,
		hooks:     nil,
		immediate: false,
		offset:    0,
		repeat:    false,
		id:        id,
		absSecs:   0,
		alerter:   nil,

		priority: priority,
	}
}

// ServiceJSONNew creates a new service instance, which will query a
// json restful endpoint.
func ServiceJSONNew(rawurl string, secs int) Service {
	if secs <= 0 {
		log.Fatal("NO. GOD. NO. GOD PLEASE NO. NO. NO. NOOOOOOOO.")
	}

	u, err := url.Parse(rawurl)
	nilOrDie(err, "invalid http endpoint url")
	hooks := make([]HookSignature, 0)

	priority := secs + int(time.Now().Unix())

	atomic.AddUint64(&lastID, 1)

	return Service{
		url:       u,
		secs:      secs,
		hooks:     hooks,
		immediate: false,
		offset:    0,
		repeat:    false,
		id:        lastID,
		absSecs:   0,
		alerter:   nil,
		priority:  priority,
	}
}

// Stop service will stop the ticker, and gracefully exit it.
// TODO DEPRACATED
func (s *Service) Stop() {
	log.Print("stopping service: ", s.url.String())
	log.Fatal("do not run me no more")
}

// AbsSecs sets the absolute seconds of last timer addition
func (s *Service) AbsSecs(secs int) {
	s.absSecs = secs
}

// GetAbsSecs returns the absolute seconds of last timer addition
func (s *Service) GetAbsSecs() int {
	return s.absSecs
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
func (s *Service) Immediate(val bool) {
	s.immediate = val
}

// IsImmediate returns true if service is immediate
func (s *Service) IsImmediate() bool {
	return s.immediate
}

// Offset sets the time before the service starts ticking
func (s *Service) Offset(offset int) {
	s.offset = offset
}

// Repeat makes the service repeatable
func (s *Service) Repeat(rep bool) {
	s.repeat = rep
}

// IsRepeating says whether a service repeats or not
func (s *Service) IsRepeating() bool {
	return s.repeat
}

// ID returns the unique identifier of the service
func (s *Service) ID() uint64 {
	return s.id
}

// GetSecs returns the number of seconds
func (s *Service) GetSecs() int {
	return s.secs
}

// UniqStr combines the label and id in order to have a unique, human
// readable label.
func (s *Service) UniqStr() string {
	var ret string

	if s.Label != nil {
		ret = fmt.Sprintf("%s-%d", *s.Label, s.id)
	} else {
		ret = fmt.Sprintf("%d", s.id)
	}

	return ret
}

// DataRepo sets where the data processed should be stored in
func (s *Service) DataRepo(repo *StatusServer) {
	s.repo = repo
}

// Execute the service
func (s *Service) Execute() {
	// TODO this should eventually be split into something else
	// (ie services should have some sort of interface, and split
	// the logic of http querying and hook execution)
	if s.url != nil && s.repo != nil {
		// If there is a url and repo specified, then fetch
		// the data and store it
		jsonQuery(s, s.repo)
	}

	for _, hook := range s.hooks {
		ok, result := hook(s.repo)
		s.maybeAlert(ok, result)
	}
}

func (s *Service) maybeAlert(shouldAlert bool, result interface{}) {
	if s.alerter == nil || !shouldAlert {
		return
	}

	s.alerter.Ch <- AlertMessage{
		Response:      result,
		Endpoint:      "TODO",
		Now:           "TODO",
		CynicHostname: "TODO",
	}
}

// SetSecs sets the seconds of the service to fire on. This will not
// take effect on the wheel, unless it's a repeatable service, and was
// re-added on the next tick.
func (s *Service) SetSecs(secs int) {
	s.secs = secs
}

// GetOffset returns the offset time of the service
func (s *Service) GetOffset() int {
	return s.offset
}

// SetAbsExpiry sets the timestamp that the service is suposed to
// expire on.
func (s *Service) SetAbsExpiry(ts int64) {
	s.absExpiry = ts + int64(s.GetSecs())
}

// GetAbsExpiry gets the timestamp
func (s *Service) GetAbsExpiry() int64 {
	return s.absExpiry
}

func (s *Service) String() string {
	return fmt.Sprintf(
		"Service<url:%v secs:%d hooks:%v immediate:%t offset:%d repeat:%t label:%v id:%d repo:%v>",
		s.url,
		s.secs,
		s.hooks,
		s.immediate,
		s.offset,
		s.repeat,
		s.Label,
		s.id,
		s.repo)
}

func jsonQuery(s *Service, t *StatusServer) {
	type serviceError struct {
		Error string `json:"error"`
	}

	address := s.url.String()

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

	// The applications of contracts/results should only be done
	// for know json service endpoints. If we have a custom hook,
	// the hook must be the one that decides what goes in the
	// status cache.
	t.Update(address, json) // TODO: better use service.UniqStr() here
}
