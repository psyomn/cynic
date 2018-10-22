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
	"testing"

	"github.com/psyomn/cynic"
)

func TestServiceIdIncreaseMonotonically(t *testing.T) {
	s1 := cynic.ServiceJSONNew("www.google.com", 1)
	s2 := cynic.ServiceJSONNew("www.hahaha.com", 2)
	s3 := cynic.ServiceJSONNew("www.derp.com", 3)

	assert(t, s1.ID() != s2.ID() &&
		s1.ID() != s3.ID() &&
		s2.ID() != s3.ID())
}

func TestServiceWithQueryAndRepo(t *testing.T) {
	ran := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{}")
		ran = true
	}))
	defer ts.Close()

	ser := cynic.ServiceJSONNew(ts.URL, 1)
	repo := cynic.StatusServerNew("0", "/status/testservicewithqueryandrepo")
	ser.DataRepo(&repo)

	ser.Execute()

	assert(t, ran)
}

func TestServiceWithQueryNoRepo(t *testing.T) {
	ran := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{}")
		ran = true
	}))
	defer ts.Close()

	ser := cynic.ServiceJSONNew(ts.URL, 1)
	repo := cynic.StatusServerNew("0", "/status/testservicewithquerynorepo")
	ser.DataRepo(&repo)

	ser.Execute()

	assert(t, ran)
}

func TestServiceExecution(t *testing.T) {
	ran := false
	ser := cynic.ServiceNew(1)
	ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		ran = true
		return false, 0
	})

	ser.Execute()

	assert(t, ran)
}
