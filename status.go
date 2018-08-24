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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// StatusServer is a server that will serve information about all the
// services cynic will be observing
type StatusServer struct {
	contractResults *sync.Map
	server          *http.Server
	slackHook       *string
	alerter         *time.Ticker
	wg              *sync.WaitGroup
}

const (
	// StatusPort is the default port the status http server will
	// respond on.
	StatusPort = "9999"

	// statusPokeTime is how much time to check the map, and then if the
	// map has entries, poke on the channel. This eventually has to be
	// done a little better
	statusPokeTime = 60
)

// StatusServerNew creates a new status server for cynic
func StatusServerNew(port string, slackHook *string) StatusServer {
	server := &http.Server{
		Addr:           ":" + port,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return StatusServer{&sync.Map{}, server, slackHook, nil, nil}
}

// Start stats a new server. Should be running in the background.
func (s *StatusServer) Start() {
	if s.slackHook != nil {
		log.Print("starting alerter")

		var wg sync.WaitGroup
		s.alerter = time.NewTicker(time.Duration(statusPokeTime) * time.Second)

		s.wg = &wg
		s.wg.Add(1)
		go s.alertLoop()
	}

	http.HandleFunc("/status", s.makeResponse)
	log.Print(s.server.ListenAndServe())
}

// Stop gracefully shuts down the server
func (s *StatusServer) Stop() {
	if s.slackHook != nil {
		s.alerter.Stop()
		s.wg.Wait()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	nilAndOk(s.server.Shutdown(ctx), "could not shutdown server gracefully")
}

// Update updates the information about all the contracts that are
// running on different endpoints
func (s *StatusServer) Update(key string, value interface{}) {
	s.contractResults.Store(key, value)
}

func (s *StatusServer) makeResponse(w http.ResponseWriter, _ *http.Request) {
	var tmp map[string]interface{}
	tmp = make(map[string]interface{})
	s.contractResults.Range(func(k interface{}, v interface{}) bool {
		keyStr, _ := k.(string)
		tmp[keyStr] = v
		return true
	})

	jsonEnc, err := json.Marshal(tmp)

	if err != nil {
		// TODO maybe there's something cleaner I can do here
		nilAndOk(err, "problem generating json for status endpoint")
		fmt.Fprintf(w, "{\"error\":\"could not format status data\"}")
	} else {
		fmt.Fprintf(w, string(jsonEnc))
	}
}

func countMap(m *sync.Map) int {
	count := 0
	m.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func (s *StatusServer) alertLoop() {
	for range s.alerter.C {
		if countMap(s.contractResults) > 0 {
			values := map[string]interface{}{
				"text":       "I LIVE!",
				"link_names": 1}

			jsonValue, _ := json.Marshal(values)
			_, err := http.Post(*s.slackHook,
				"application/json",
				bytes.NewBuffer(jsonValue))

			if err != nil {
				fmt.Println("UH OH: ", err)
			}

		}
	}
}
