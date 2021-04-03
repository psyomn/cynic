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
package cynic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

// StatusCache stores any sort of information that is possibly
// retrieved or calculated by events. A server can be started to
// retrieve information in the map in json format.
type StatusCache struct {
	server          *http.Server
	contractResults *sync.Map
	listener        net.Listener
	alerter         *time.Ticker
	root            string

	snapshot       *SnapshotStore
	snapshotConfig *SnapshotConfig
}

const (
	// StatusPort is the default port the status http server will
	// respond on.
	StatusPort = "9999"

	// DefaultStatusEndpoint is where the default status json can
	// be retrieved from
	DefaultStatusEndpoint = "/status/"

	defaultLinksEndpoint = "/links"
)

// StatusServerNew creates a new status server for cynic
func StatusServerNew(host, port, root string) StatusCache {
	server := &http.Server{
		Addr:           host + ":" + port,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		panic(err)
	}

	return StatusCache{
		contractResults: &sync.Map{},
		listener:        listener,
		server:          server,
		alerter:         nil,
		root:            root,
		snapshot:        nil,
		snapshotConfig:  nil,
	}
}

// WithSnapshots will make the cache dump snapshots of the data with
// given intervals when the service starts
func (s *StatusCache) WithSnapshots(config *SnapshotConfig) {
	store := snapshotStoreNew()
	s.snapshotConfig = config
	s.snapshot = &store
}

// Start starts all services associated with status caches. This
// includes the web interface if enabled, and the dumping of statuses
// in files.
func (s *StatusCache) Start() {
	if s.snapshotConfig != nil {
		tickerSnap := time.NewTicker(s.snapshotConfig.Interval)
		go func() {
			for range tickerSnap.C {
				s.snap()
			}
		}()

		tickerDump := time.NewTicker(s.snapshotConfig.DumpEvery)
		go func() {
			for range tickerDump.C {
				s.dump()
			}
		}()
	}

	http.HandleFunc(s.root, s.makeResponse)
	http.HandleFunc(defaultLinksEndpoint, s.makeLinks)
	err := s.server.Serve(s.listener)

	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("problem shutting down status http server: ", err)
	}
}

// Stop gracefully shuts down the server
func (s *StatusCache) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		log.Println("could not shutdown status server gracefully: ", err)
	}
}

// Update updates the information about all the contracts that are
// running on different endpoints
func (s *StatusCache) Update(key string, value interface{}) {
	s.contractResults.Store(key, value)
}

// Delete removes an entry from the sync map
func (s *StatusCache) Delete(key string) {
	s.contractResults.Delete(key)
}

// Get gets the value inside the contract results
func (s *StatusCache) Get(key string) (interface{}, error) {
	value, ok := s.contractResults.Load(key)

	if !ok {
		return nil, errors.New("could not find required value")
	}

	return value, nil
}

// NumEntries returns the number of entries in the map
func (s *StatusCache) NumEntries() (count int) {
	s.contractResults.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return
}

// GetPort this will return the port where the server was
// started. This is useful if you assign port 0 when initializing.
func (s *StatusCache) GetPort() int {
	port := s.listener.Addr().(*net.TCPAddr).Port
	return port
}

// Dump will dump the contents of the map into a snapshot file.
func (s *StatusCache) makeResponse(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Path[len(s.root):]

	jsonBuff, err := s.statusCacheToJSON(query)

	w.Header().Set("Content-Type", "application/json")

	var ret string
	if err != nil {
		log.Println("problem generating json for status endpoint: ", err)
		ret = "{\"error\":\"could not format status data\"}"
	} else {
		ret = string(jsonBuff[:])
	}

	fmt.Fprintf(w, "%s", ret)
}

func (s *StatusCache) makeLinks(w http.ResponseWriter, req *http.Request) {
	var builder strings.Builder
	builder.WriteString("<html><head></head><body><ul>")

	if s.NumEntries() == 0 {
		builder.WriteString("<h1>No links here yet.</h1>")
		goto end
	}

	builder.WriteString("<h1>Links to services</h1>")
	s.contractResults.Range(func(k interface{}, v interface{}) bool {
		keyStr, _ := k.(string)

		link := fmt.Sprintf("%v%v", s.root, keyStr)
		atag := fmt.Sprintf(`<a href="%v" target="_blank">%v</a>`, link, keyStr)

		builder.WriteString("<li>")
		builder.WriteString(atag)
		builder.WriteString("</li>")
		return true
	})

end:
	// TODO this needs cleanup
	builder.WriteString("</body></html>")
	if _, err := w.Write([]byte(builder.String())); err != nil {
		log.Println(err)
	}
}

func (s *StatusCache) statusCacheToJSON(query string) ([]byte, error) {
	tmp := make(map[string]interface{})
	s.contractResults.Range(func(k interface{}, v interface{}) bool {
		keyStr, _ := k.(string)
		tmp[keyStr] = v
		return true
	})

	var toEncode interface{}
	if len(query) > 0 {
		toEncode = tmp[query]
	} else {
		toEncode = tmp
	}

	jsonEnc, err := json.Marshal(toEncode)
	return jsonEnc, err
}

func (s *StatusCache) snap() {
	data, err := s.statusCacheToJSON("")
	if err != nil {
		log.Println("problem snapping map data")
		return
	}

	snp := snapshot{
		Timestamp: time.Now().Unix(),
		Data:      string(data),
	}
	s.snapshot.add(&snp)
}

func (s *StatusCache) dump() {
	strDate := time.Now().Format(time.RFC3339)
	filename := fmt.Sprintf("%s.%v.cynic", strDate, s.snapshot.Version)

	dumpPath := path.Join(s.snapshotConfig.Path, filename)
	error := s.snapshot.encodeToFile(dumpPath)
	if error != nil {
		log.Println("problem encoding and dumping to file: ", error)
	}

	s.snapshot.clear()
}
