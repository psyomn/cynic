/*
Package cynic monitors you from the ceiling. Library interface goes
here.

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
	"sync"
	"time"
)

const (
	// StopEvent is the signal to stop the running querying event
	StopEvent = iota

	// AddEvent adds a event to a running cynic instance
	AddEvent

	// DeleteEvent removes a event from a running cynic instance
	DeleteEvent
)

// Session is the configuration a cynic instance requires to start
// running and working
type Session struct {
	Events        []Event
	StatusServers []StatusServer
	Alerter       *Alerter
}

// Start starts a cynic instance, with any provided hooks.
func Start(session Session) {
	session.Alerter.Start()
	defer session.Alerter.Stop()

	planner := PlannerNew()

	for i := 0; i < len(session.Events); i++ {
		planner.Add(&session.Events[i])
		session.Events[i].alerter = session.Alerter
	}

	ticker := time.NewTicker(time.Second)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for range ticker.C {
			planner.Tick()
		}
		wg.Done()
	}()
	defer ticker.Stop()

	for _, statusSer := range session.StatusServers {
		statusSer.Start()
		defer statusSer.Stop()
	}

	wg.Wait()
}
