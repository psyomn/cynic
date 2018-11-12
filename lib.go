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
	"time"
)

const (
	// StopService is the signal to stop the running querying service
	StopService = iota

	// AddService adds a service to a running cynic instance
	AddService

	// DeleteService removes a service from a running cynic instance
	DeleteService
)

// Session is the configuration a cynic instance requires to start
// running and working
type Session struct {
	Services      []Service
	StatusServers []StatusServer
	Alerter       *Alerter
}

// Start starts a cynic instance, with any provided hooks.
func Start(session Session) {
	session.Alerter.Start()
	defer session.Alerter.Stop()

	planner := PlannerNew()

	for i := 0; i < len(session.Services); i++ {
		planner.Add(&session.Services[i])
		session.Services[i].alerter = session.Alerter
	}

	ticker := time.NewTicker(time.Second)
	// TODO: maybe use planner.run
	go func() {
		for range ticker.C {
			planner.Tick()
		}
	}()
	defer ticker.Stop()

	for _, statusSer := range session.StatusServers {
		statusSer.Start()
		defer statusSer.Stop()
	}
}
