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
	"container/heap"
	"time"
)

// Planner is a structure that manages events inserted with expiration
// timestamps. The underlying data structures are magic, and you
// shouldn't care about them, unless you're opening up the hatch and
// stuff.
type Planner struct {
	services ServiceQueue
	ticks    int
}

// PlannerNew creates a new, empty, timing wheel.
func PlannerNew() *Planner {
	var tw Planner
	tw.services = make(ServiceQueue, 0)
	return &tw
}

// Tick moves the cursor of the timing wheel, by one second.
func (s *Planner) Tick() {
	for {
		if s.services.Len() == 0 {
			break
		}

		rootTimestamp, _ := s.services.PeekTimestamp()

		if s.ticks >= int(rootTimestamp) {
			service := heap.Pop(&s.services).(*Service)
			service.Execute()

			if service.IsRepeating() {
				s.Add(service)
			}

		} else {
			break
		}
	}

	s.ticks++
}

// Add adds an event to the planner
func (s *Planner) Add(service *Service) {
	var expiry int64

	if service.IsImmediate() {
		expiry = 1
		service.Immediate(false)
	} else {
		expiry = int64(service.GetSecs() + s.ticks)
	}

	service.SetAbsExpiry(expiry)
	s.services.Push(service)
}

// Run runs the wheel, with a 1s tick
func (s *Planner) Run() {
	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			s.Tick()
		}
	}()
	defer ticker.Stop()
}
