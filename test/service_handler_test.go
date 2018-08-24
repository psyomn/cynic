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
	"sync/atomic"
	"testing"
	"time"

	"github.com/psyomn/cynic"
)

func makeSession() cynic.Session {
	return cynic.Session{StatusPort: cynic.StatusPort, SlackHook: nil}
}

func TestMakeService(t *testing.T) {
	services := cynic.AddressBookNew(makeSession())

	services.Add("www.google.com", 60, []string{})
	services.Add("www.example.com", 12, []string{})
}

func TestNumEntries(t *testing.T) {
	services := cynic.AddressBookNew(makeSession())
	Assert(t, services.NumEntries() == 0)
	Assert(t, services.NumEntries() == 0)

	services.Add("www.google.com", 60, []string{})
	Assert(t, services.NumEntries() == 1)

	services.Add("www.example.com", 32, []string{})
	Assert(t, services.NumEntries() == 2)

	services.Add("www.google.com", 60, []string{})
	Assert(t, services.NumEntries() == 2)
}

func TestIntegration(t *testing.T) {
	var count1, count2, count3 int32

	services := cynic.AddressBookNew(makeSession())
	fixtureSimple := ReadFixture(FixturePathSimple)
	fixtureStatus := ReadFixture(FixturePathStatus)

	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, fixtureSimple)
		atomic.AddInt32(&count1, 1)
	}))

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, fixtureStatus)
		atomic.AddInt32(&count2, 1)
	}))

	ts3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - ARRRRRRRRGHHHHH!!!!!!"))
		atomic.AddInt32(&count3, 1)
	}))

	defer ts1.Close()
	defer ts2.Close()
	defer ts3.Close()

	{ /* get val and extra */
		jpathContracts := []string{
			/* get simple key value */
			"$.random.stuff",
			/* "$..val", "$..extra", */
			"$.entries[?(@.val>10)]",
		}

		services.Add(ts1.URL, 1, jpathContracts)
	}

	{ /* get simple key/values */
		jpathContracts := []string{
			"$.status.state",
			"$.status.date",
			"$.status.build",
		}

		services.Add(ts2.URL, 1, jpathContracts)
	}

	{ /* test for bad requests/bad server */
		services.Add(ts3.URL, 1, nil)
	}

	signal := make(chan int)

	go func() { services.RunQueryService(signal) }()

	for atomic.LoadInt32(&count1) == 0 ||
		atomic.LoadInt32(&count2) == 0 ||
		atomic.LoadInt32(&count3) == 0 {
		time.Sleep(500 * time.Millisecond)
	}

	signal <- cynic.StopService
}
