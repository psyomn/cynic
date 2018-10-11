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
// - A service is bound to a data repository/cache
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
}

var lastID uint64

// TODO make sure that this is used everywhere.
type serviceError struct {
	Error string `json:"error"`
}

// ServiceNew creates a new service that is primarily used for pure
// execution
func ServiceNew(secs int) Service {
	atomic.AddUint64(&lastID, 1)
	return Service{
		url:       nil,
		secs:      secs,
		hooks:     nil,
		immediate: false,
		offset:    0,
		repeat:    false,
		id:        lastID,
	}
}

// ServiceJSONNew creates a new service instance, which will query a
// json restful endpoint.
func ServiceJSONNew(rawurl string, secs int) Service {
	u, err := url.Parse(rawurl)
	nilOrDie(err, "invalid http endpoint url")
	hooks := make([]HookSignature, 0)

	atomic.AddUint64(&lastID, 1)

	return Service{
		url:       u,
		secs:      secs,
		hooks:     hooks,
		immediate: false,
		offset:    0,
		repeat:    false,
		id:        lastID,
	}
}

// Stop service will stop the ticker, and gracefully exit it.
func (s *Service) Stop() {
	log.Print("stopping service: ", s.url.String())
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

// UniqStr combines the label and id in order to have a unique, human
// readable label.
func (s *Service) UniqStr() string {
	return fmt.Sprintf("%s-%d", *s.Label, s.id)
}

// DataRepo sets where the data processed should be stored in
func (s *Service) DataRepo(repo *StatusServer) {
	s.repo = repo
}

// Execute the service
func (s *Service) Execute() {
	// TODO this should eventually be split into something else
	// (ie services should have some sort)
	if s.url != nil && s.repo != nil {
		// If there is a url and repo specified, then fetch
		// the data and store it
		workerQuery(s, s.repo)
	}

	for _, hook := range s.hooks {
		hook(s.repo)
	}
}

func workerQuery(s *Service, t *StatusServer) {
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
	log.Println("need to use the body: ", body)

	// TODO need to handle this at some point
	// var json EndpointJSON = parseEndpointJSON(body[:])

	// TODO: need to apply contracts here -- or do i?
	// results := applyContracts(addressBook, s, &json)
	results := 0
	t.Update(address, results)
}

// TODO this could probably be a object method instead...
// func applyContracts(s *Service, json *EndpointJSON) interface{} {
// 	type result struct {
// 		HookResults map[string]interface{} `json:"hooks"`
// 		Timestamp   int64                  `json:"timestamp"`
// 		HumanTime   string                 `json:"human_time"`
// 		Alert       bool                   `json:"alert"`
// 	}
//
// 	var ret result
// 	sumAlerts := false
// 	ret.HookResults = make(map[string]interface{})
//
// 	for i := 0; i < len(s.hooks); i++ {
// 		hookName := getFuncName(s.hooks[i])
// 		retAlert, hookRet := s.hooks[i](*json)
// 		sumAlerts = sumAlerts || retAlert
//
// 		if res, ok := ret.HookResults[hookName]; ok {
// 			tempResult := res.(result)
// 			tempResult.HookResults[hookName] = hookRet
// 			tempResult.Timestamp = time.Now().Unix()
// 			ret.HookResults[hookName] = tempResult
// 			ret.Alert = retAlert
// 		} else {
// 			ret.HookResults[hookName] = hookRet
// 			ret.Timestamp = time.Now().Unix()
// 			ret.HumanTime = time.Now().Format(time.RFC850)
// 			ret.Alert = retAlert
// 		}
// 	}
//
// 	if sumAlerts {
// 		hostname, err := os.Hostname()
// 		if err != nil {
// 			hostname = "nohost"
// 		}
// 		message := AlertMessage{
// 			Endpoint:      s.url.String(),
// 			Response:      ret,
// 			CynicHostname: hostname,
// 			Now:           time.Now().Format(time.RFC850),
// 		}
//
// 		// TODO, need a better alerting mechanism
// 		log.Println("This would be added to the queue alert thingy: ", message)
// 		// addressBook.queueAlert(&message)
// 	}
//
// 	return ret
// }
