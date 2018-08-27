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
	return cynic.Session{
		StatusPort: cynic.StatusPort,
		SlackHook:  nil,
		Alerter:    nil,
		AlertTime:  0,
	}
}

func TestMakeService(t *testing.T) {
	services := cynic.AddressBookNew(makeSession())
	ser1 := cynic.ServiceNew("www.google.com", 60)
	ser2 := cynic.ServiceNew("www.example.com", 12)

	services.AddService(&ser1)
	services.AddService(&ser2)
}

func TestNumEntries(t *testing.T) {
	services := cynic.AddressBookNew(makeSession())
	Assert(t, services.NumEntries() == 0)
	Assert(t, services.NumEntries() == 0)

	{
		service := cynic.ServiceNew("www.google.com", 60)
		services.AddService(&service)
		Assert(t, services.NumEntries() == 1)
	}

	{
		service := cynic.ServiceNew("www.example.com", 60)
		services.AddService(&service)
		Assert(t, services.NumEntries() == 2)
	}

	{
		service := cynic.ServiceNew("www.google.com", 60)
		services.AddService(&service)
		Assert(t, services.NumEntries() == 2)
	}
}

func TestIntegration(t *testing.T) {
	var hcnt1, hcnt2, hcnt3 int32
	var count1, count2, count3 int32

	services := cynic.AddressBookNew(makeSession())

	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{}")
		atomic.AddInt32(&count1, 1)
	}))

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{}")
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

	// FIRE
	{ // get val and extra
		service := cynic.ServiceNew(ts1.URL, 1)
		service.AddHook(func(_ *cynic.AddressBook, _ interface{}) (bool, interface{}) {
			atomic.AddInt32(&hcnt1, 1)
			return false, 42
		})

		service.AddHook(func(_ *cynic.AddressBook, _ interface{}) (bool, interface{}) {
			atomic.AddInt32(&hcnt2, 1)
			return false, 42
		})

		service.AddHook(func(_ *cynic.AddressBook, _ interface{}) (bool, interface{}) {
			fmt.Print("BY THE POWER OF GREYSKULL")
			atomic.AddInt32(&hcnt3, 1)
			return true, 42
		})

		services.AddService(&service)
	}

	{ // check service exists
		service, ok := services.Get(ts1.URL)
		if !ok {
			t.Fatal("location should be in map")
		}
		Assert(t, service.NumHooks() == 3)
	}

	{ // get simple key/values
		service := cynic.ServiceNew(ts2.URL, 1)
		services.AddService(&service)
	}

	{ // test for bad requests/bad server
		service := cynic.ServiceNew(ts3.URL, 1)
		services.AddService(&service)
	}

	Assert(t, services.NumEntries() == 3)

	signal := make(chan int)
	go func() { services.Run(signal) }()

	// wait until things have been seen at least once
	for atomic.LoadInt32(&count1) == 0 ||
		atomic.LoadInt32(&count2) == 0 ||
		atomic.LoadInt32(&count3) == 0 ||
		atomic.LoadInt32(&hcnt1) == 0 ||
		atomic.LoadInt32(&hcnt2) == 0 ||
		atomic.LoadInt32(&hcnt3) == 0 {
		time.Sleep(500 * time.Millisecond)
	}

	signal <- cynic.StopService
}

func TestAddHook(t *testing.T) {
	service := cynic.ServiceNew("www.google.com", 60)
	service.AddHook(func(entry interface{}) interface{} {
		return 42
	})
	Assert(t, service.NumHooks() == 1)
}

func TestAddServiceWithHook(t *testing.T) {
	location := "www.google.com"
	service := cynic.ServiceNew(location, 1)
	service.AddHook(func(_ *cynic.AddressBook, _ interface{}) (bool, interface{}) {
		return false, 1
	})

	book := cynic.AddressBookNew(makeSession())
	book.AddService(&service)
	Assert(t, book.NumEntries() == 1)

	getService, ok := book.Get(service.URL.String())
	if !ok {
		t.Fail()
	}

	Assert(t, getService.URL.String() == service.URL.String())
	Assert(t, getService.Secs == service.Secs)
	Assert(t, getService.NumHooks() == service.NumHooks())
}

func TestSwapLocationsDynamically(t *testing.T) {
	fmt.Println("TODO")
}
