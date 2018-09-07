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

package cynictesting

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/psyomn/cynic"
)

func TestCRUD(t *testing.T) {
	server := cynic.StatusServerNew("0", "TestCRUD")

	server.Update("hello", "kitty")
	server.Update("goodbye", "human")
	server.Update("blarrgh", "arggh")
	Assert(t, server.NumEntries() == 3)

	server.Delete("blarrgh")
	Assert(t, server.NumEntries() == 2)

	server.Delete("blarrgh")
	Assert(t, server.NumEntries() == 2)

	server.Update("potato", "tomato")
	Assert(t, server.NumEntries() == 3)
	server.Update("potato", "AAARGH")
	Assert(t, server.NumEntries() == 3)
}

func TestRestEndpoint(t *testing.T) {
	endpoint := "/testrestendpoint"
	server := cynic.StatusServerNew("0", endpoint)

	server.Update("hello", "kitty")
	server.Update("whosagood", "doggo")
	server.Update("ARGH", "BLARGH")
	Assert(t, server.NumEntries() == 3)

	port := strconv.Itoa(server.GetPort())

	go func() { server.Start() }()

	resp, err := http.Get("http://127.0.0.1:" + port + endpoint)
	if err != nil {
		t.Fatal("could not connect: ", err)
	}
	defer resp.Body.Close()

	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var values map[string]string

	jsonErr := json.Unmarshal(text, &values)

	if jsonErr != nil {
		t.Fatal(err)
	}

	Assert(t, values["hello"] == "kitty")
	Assert(t, values["whosagood"] == "doggo")
	Assert(t, values["ARGH"] == "BLARGH")

	server.Stop()
}
