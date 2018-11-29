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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// StatusServer is a server that will serve information about all the
// events cynic will be observing
type StatusServer struct {
	contractResults *sync.Map
	listener        net.Listener
	server          *http.Server
	alerter         *time.Ticker
	root            string
}

const (
	// StatusPort is the default port the status http server will
	// respond on.
	StatusPort = "9999"

	// DefaultStatusEndpoint is where the default status json can
	// be retrieved from
	DefaultStatusEndpoint = "/status"
)

// StatusServerNew creates a new status server for cynic
func StatusServerNew(port, root string) StatusServer {
	server := &http.Server{
		Addr:           ":" + port,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		panic(err)
	}

	return StatusServer{
		contractResults: &sync.Map{},
		listener:        listener,
		server:          server,
		alerter:         nil,
		root:            root,
	}
}

// Start stats a new server. Should be running in the background.
func (s *StatusServer) Start() {
	http.HandleFunc(s.root, s.makeResponse)
	err := s.server.Serve(s.listener)

	if err != http.ErrServerClosed {
		log.Fatal("problem shutting down status http server: ", err)
	}
}

// Stop gracefully shuts down the server
func (s *StatusServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		log.Println("could not shutdown status server gracefully: ", err)
	}
}

// Update updates the information about all the contracts that are
// running on different endpoints
func (s *StatusServer) Update(key string, value interface{}) {
	s.contractResults.Store(key, value)
}

// Delete removes an entry from the sync map
func (s *StatusServer) Delete(key string) {
	s.contractResults.Delete(key)
}

// NumEntries returns the number of entries in the map
func (s *StatusServer) NumEntries() (count int) {
	s.contractResults.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return
}

// GetPort this will return the port where the server was
// started. This is useful if you assign port 0 when initializing.
func (s *StatusServer) GetPort() int {
	port := s.listener.Addr().(*net.TCPAddr).Port
	return port
}

func (s *StatusServer) makeResponse(w http.ResponseWriter, _ *http.Request) {
	tmp := make(map[string]interface{})
	s.contractResults.Range(func(k interface{}, v interface{}) bool {
		keyStr, _ := k.(string)
		tmp[keyStr] = v
		return true
	})

	jsonEnc, err := json.Marshal(tmp)

	if err != nil {
		// TODO maybe there's something cleaner I can do here
		log.Println("problem generating json for status endpoint: ", err)
		fmt.Fprintf(w, "{\"error\":\"could not format status data\"}")
	} else {
		fmt.Fprintf(w, string(jsonEnc))
	}
}
