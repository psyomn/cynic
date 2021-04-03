/*
Package cynic monitors you from the ceiling

Copyright 2018-2021 Simon Symeonidis (psyomn)

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
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/psyomn/cynic"
)

func TestCRUD(t *testing.T) {
	server := cynic.StatusServerNew("", "0", "TestCRUD")

	server.Update("hello", "kitty")
	server.Update("goodbye", "human")
	server.Update("blarrgh", "arggh")

	getHello, _ := server.Get("hello")
	assert(t, getHello.(string) == "kitty")
	getGoodbye, _ := server.Get("goodbye")
	assert(t, getGoodbye.(string) == "human")
	getBlargh, _ := server.Get("blarrgh")
	assert(t, getBlargh.(string) == "arggh")
	assert(t, server.NumEntries() == 3)

	server.Delete("blarrgh")
	assert(t, server.NumEntries() == 2)

	server.Delete("blarrgh")
	assert(t, server.NumEntries() == 2)

	server.Update("potato", "tomato")
	getPotato, _ := server.Get("potato")
	assert(t, getPotato.(string) == "tomato")
	assert(t, server.NumEntries() == 3)

	server.Update("potato", "AAARGH")
	getPotato, _ = server.Get("potato")
	assert(t, server.NumEntries() == 3)
	assert(t, getPotato.(string) == "AAARGH")
}

func TestGetNonExistantKey(t *testing.T) {
	status := cynic.StatusServerNew("", "0", "9999")
	status.Update("somekey", "hassomething")

	_, err := status.Get("somekey")
	assert(t, err == nil)

	_, err = status.Get("doesntexist")
	assert(t, err != nil)
}

func TestConcurrentCRUD(t *testing.T) {
	status := cynic.StatusServerNew("", "0", "9999")
	var wgw, wgr sync.WaitGroup
	n := 100

	status.Update("counter", 1)
	status.Update("timestamp", time.Now().Unix())

	for i := 0; i < n; i++ {
		wgw.Add(1)
		go func(index int) {
			defer wgw.Done()
			key := fmt.Sprintf("blargh-%d", index)
			status.Update(key, index)
		}(i)
	}

	for i := 0; i < n; i++ {
		wgr.Add(1)
		go func(index int) {
			defer wgr.Done()
			key := fmt.Sprintf("blargh-%d", index)

			// this might or might not find some of the required keys
			// due to the async reads/writes, and that's OK. mostly
			// here to evoke errors on the race condition tester
			_, _ = status.Get(key)
		}(i)
	}

	wgr.Wait()
	wgw.Wait()
}

func TestRestEndpoint(t *testing.T) {
	endpoint := "/testrestendpoint"
	server := cynic.StatusServerNew("", "0", endpoint)

	server.Update("hello", "kitty")
	server.Update("whosagood", "doggo")
	server.Update("ARGH", "BLARGH")
	assert(t, server.NumEntries() == 3)

	port := strconv.Itoa(server.GetPort())

	go func() { server.Start() }()

	cli := &http.Client{}
	req, err := makeBackgroundRequest("http://127.0.0.1:" + port + endpoint)
	if err != nil {
		t.Fatal("could not create request:", err)
	}

	resp, err := cli.Do(req)
	if err != nil {
		t.Fatal("could not connect:", err)
	}
	defer resp.Body.Close()

	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("error reading all:", err)
	}

	var values map[string]string

	jsonErr := json.Unmarshal(text, &values)

	if jsonErr != nil {
		t.Fatal(err)
	}

	assert(t, values["hello"] == "kitty")
	assert(t, values["whosagood"] == "doggo")
	assert(t, values["ARGH"] == "BLARGH")

	server.Stop()
}
