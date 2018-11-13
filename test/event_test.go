/*
Package cynic_testing tests that it can monitor you from the ceiling.

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

package cynictesting

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/psyomn/cynic"
)

func TestEventIdIncreaseMonotonically(t *testing.T) {
	s1 := cynic.EventJSONNew("www.google.com", 1)
	s2 := cynic.EventJSONNew("www.hahaha.com", 2)
	s3 := cynic.EventJSONNew("www.derp.com", 3)

	assert(t, s1.ID() != s2.ID() &&
		s1.ID() != s3.ID() &&
		s2.ID() != s3.ID())
}

func TestAtomicEventIds(t *testing.T) {
	var wg sync.WaitGroup
	routines := 30
	eventCount := 20

	var ids sync.Map

	for j := 0; j < routines; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < eventCount; i++ {
				event := cynic.EventNew(1)
				eventID := event.ID()

				if actual, ok := ids.Load(eventID); ok {
					ids.Store(eventID, actual.(int)+1)
				} else {
					ids.Store(eventID, 1)
				}
			}
		}()
	}
	wg.Wait()

	ids.Range(func(_, v interface{}) bool {
		assert(t, v.(int) == 1)
		return true
	})
}

func TestEventWithQueryAndRepo(t *testing.T) {
	ran := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{}")
		ran = true
	}))
	defer ts.Close()

	ser := cynic.EventJSONNew(ts.URL, 1)
	repo := cynic.StatusServerNew("0", "/status/testeventwithqueryandrepo")
	ser.DataRepo(&repo)

	ser.Execute()

	assert(t, ran)
}

func TestEventWithQueryNoRepo(t *testing.T) {
	ran := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{}")
		ran = true
	}))
	defer ts.Close()

	ser := cynic.EventJSONNew(ts.URL, 1)
	repo := cynic.StatusServerNew("0", "/status/testeventwithquerynorepo")
	ser.DataRepo(&repo)

	ser.Execute()

	assert(t, ran)
}

func TestEventExecution(t *testing.T) {
	ran := false
	ser := cynic.EventNew(1)
	ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		ran = true
		return false, 0
	})

	ser.Execute()

	assert(t, ran)
}
