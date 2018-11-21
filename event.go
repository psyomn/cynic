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
	"os"
	"sync/atomic"
	"time"
)

const (
	// EventDefault is for events that query in the default
	// way, which is JSON, restful endpoints
	EventDefault = iota

	// EventCustom is for events that query other endpoints
	// instead of JSON. It is up to the user to implement support
	// for these endpoints.
	EventCustom
)

// HookParameters is any state that should be passed to the hook
type HookParameters struct {
	// Planner is access to the planner that the hook executes
	// on. The user for example, should be able to add more events
	// through a hook.
	Planner *Planner

	// Status exposes the status repo. It acts as a repository for
	// hooks to store information after execution.
	Status *StatusServer

	// Extra is meant to be used by the user for any extra state
	// that needs to be passed to the hooks.
	Extra interface{}
}

// HookSignature specifies what the event hooks should look like.
type HookSignature = func(*HookParameters) (bool, interface{})

// Event is some event that should be executed in a specified
// amount of time. There are no real time guarantees.
// - A event is an action
// - A event can have many:
//   - hooks (that can act as contracts)
// - A event may be bound to a data repository/cache
type Event struct {
	id        uint64
	url       *url.URL
	secs      int
	hooks     []HookSignature
	immediate bool
	offset    int
	repeat    bool
	Label     string
	planner   *Planner

	repo    *StatusServer
	alerter *Alerter

	absExpiry int64

	index    int
	priority int
	deleted  bool

	extra interface{}
}

var lastID uint64

// EventNew creates a new event that is primarily used for pure
// execution
func EventNew(secs int) Event {
	if secs <= 0 {
		log.Fatal("NO. GOD. NO. GOD PLEASE NO. NO. NO. NOOOOOOOO.")
	}

	hooks := make([]HookSignature, 0)
	id := atomic.AddUint64(&lastID, 1)

	priority := secs + int(time.Now().Unix())

	return Event{
		url:       nil,
		secs:      secs,
		hooks:     hooks,
		immediate: false,
		offset:    0,
		repeat:    false,
		id:        id,
		alerter:   nil,
		priority:  priority,
		deleted:   false,
	}
}

// EventJSONNew creates a new event instance, which will query a
// json restful endpoint.
func EventJSONNew(rawurl string, secs int) Event {
	if secs <= 0 {
		log.Fatal("NO. GOD. NO. GOD PLEASE NO. NO. NO. NOOOOOOOO.")
	}

	u, err := url.Parse(rawurl)
	nilOrDie(err, "invalid http endpoint url")
	hooks := make([]HookSignature, 0)

	priority := secs + int(time.Now().Unix())
	id := atomic.AddUint64(&lastID, 1)

	return Event{
		url:       u,
		secs:      secs,
		hooks:     hooks,
		immediate: false,
		offset:    0,
		repeat:    false,
		id:        id,
		alerter:   nil,
		priority:  priority,
		deleted:   false,
	}
}

// AddHook appends a hook to the event
func (s *Event) AddHook(fn HookSignature) {
	s.hooks = append(s.hooks, fn)
}

// NumHooks counts the hooks
func (s *Event) NumHooks() int {
	return len(s.hooks)
}

// Immediate will make the event run immediately
func (s *Event) Immediate(val bool) {
	s.immediate = val
}

// IsImmediate returns true if event is immediate
func (s *Event) IsImmediate() bool {
	return s.immediate
}

// Offset sets the time before the event starts ticking
func (s *Event) Offset(offset int) {
	s.offset = offset
}

// Repeat makes the event repeatable
func (s *Event) Repeat(rep bool) {
	s.repeat = rep
}

// IsRepeating says whether a event repeats or not
func (s *Event) IsRepeating() bool {
	return s.repeat
}

// ID returns the unique identifier of the event
func (s *Event) ID() uint64 {
	return s.id
}

// GetSecs returns the number of seconds.
func (s *Event) GetSecs() int {
	return s.secs
}

// SetSecs sets the seconds of the event to fire on.
func (s *Event) SetSecs(secs int) {
	s.secs = secs
}

// UniqStr combines the label and id in order to have a unique, human
// readable label.
func (s *Event) UniqStr() string {
	var ret string

	if s.Label != "" {
		ret = fmt.Sprintf("%s-%d", s.Label, s.id)
	} else {
		ret = fmt.Sprintf("%d", s.id)
	}

	return ret
}

// DataRepo sets where the data processed should be stored in
func (s *Event) DataRepo(repo *StatusServer) {
	s.repo = repo
}

// Execute the event
func (s *Event) Execute() {
	// TODO this should eventually be split into something else
	// (ie events should have some sort of interface, and split
	// the logic of http querying and hook execution)
	if s.url != nil && s.repo != nil {
		// If there is a url and repo specified, then fetch
		// the data and store it
		jsonQuery(s, s.repo)
	}

	if s.url != nil && s.repo == nil {
		// At least warn that somethign is awry
		// TODO eventually this should be removed
		log.Println("event is a json event without repo bound: ", s.String())
	}

	for _, hook := range s.hooks {
		ok, result := hook(&HookParameters{
			s.planner,
			s.repo,
			s.extra,
		})

		s.maybeAlert(ok, result)
	}
}

func (s *Event) maybeAlert(shouldAlert bool, result interface{}) {
	if s.alerter == nil || !shouldAlert {
		return
	}

	hostVal, err := os.Hostname()
	if err != nil {
		hostVal = "badhost"
	}

	// TODO clean this up -- url should no longer be a thing
	endpoint := ""
	if s.url != nil {
		endpoint = s.url.String()
	}

	s.alerter.Ch <- AlertMessage{
		Response:      result,
		Endpoint:      endpoint,
		Now:           time.Now().Format(time.RFC3339),
		CynicHostname: hostVal,
	}
}

// GetOffset returns the offset time of the event
func (s *Event) GetOffset() int {
	return s.offset
}

// SetAbsExpiry sets the timestamp that the event is suposed to
// expire on.
func (s *Event) SetAbsExpiry(ts int64) {
	s.absExpiry = ts
	s.priority = int(ts)
}

// GetAbsExpiry gets the timestamp
func (s *Event) GetAbsExpiry() int64 {
	return s.absExpiry
}

func (s *Event) String() string {
	return fmt.Sprintf(
		"Event<url:%v secs:%d hooks:%v immediate:%t offset:%d repeat:%t label:%v id:%d repo:%v>",
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

// Delete marks event for deletion
func (s *Event) Delete() {
	s.deleted = true
}

// IsDeleted returns if event is marked for deletion
func (s *Event) IsDeleted() bool {
	return s.deleted
}

func (s *Event) setPlanner(planner *Planner) {
	s.planner = planner
}

// SetExtra state you may want passed to hooks
func (s *Event) SetExtra(extra interface{}) {
	s.extra = extra
}

// SetAlerter sets the alerter for an event.
// TODO: this should be moved to planner
func (s *Event) SetAlerter(alerter *Alerter) {
	s.alerter = alerter
}

func jsonQuery(s *Event, t *StatusServer) {
	type eventError struct {
		Error string `json:"error"`
	}

	address := s.url.String()

	resp, err := http.Get(address)
	if err != nil {
		message := "problem getting response"
		nilAndOk(err, message)
		t.Update(address, eventError{Error: message})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buff := fmt.Sprintf("got non 200 code: %d", resp.StatusCode)
		t.Update(address, eventError{Error: buff})
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		message := "problem reading data from endpoint"
		nilAndOk(err, message)
		t.Update(address, eventError{Error: message})
		return
	}

	var json EndpointJSON = parseEndpointJSON(body[:])

	// The applications of contracts/results should only be done
	// for know json event endpoints. If we have a custom hook,
	// the hook must be the one that decides what goes in the
	// status cache.
	t.Update(s.UniqStr(), json)
}
