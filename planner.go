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
	"sync"
	"time"
)

type eventMap map[uint64]*Event

// Planner is a structure that manages events inserted with expiration
// timestamps. The underlying data structures are magic, and you
// shouldn't care about them, unless you're opening up the hatch and
// stuff.
type Planner struct {
	events       EventQueue
	ticks        int
	uniqueEvents eventMap
	mux          sync.Mutex
}

// PlannerNew creates a new, empty, timing wheel.
func PlannerNew() *Planner {
	var tw Planner
	tw.events = make(EventQueue, 0)
	tw.uniqueEvents = make(eventMap)
	return &tw
}

func (s *Planner) Len() int {
	return len(s.events)
}

func (s *Planner) String() string {
	mkline := func(s string) string {
		return s + "\n"
	}
	var str string
	str += mkline("=======================")
	str += mkline("Planner")
	str += mkline("=======================")
	str += mkline("Events: \n")

	for _, el := range s.events {
		str += mkline("  - " + el.String())
	}
	str += mkline("=======================")
	return str
}

// Tick moves the cursor of the timing wheel, by one second.
func (s *Planner) Tick() {
	for {
		if s.events.Len() == 0 {
			break
		}

		rootTimestamp, _ := s.events.PeekTimestamp()

		if s.ticks >= int(rootTimestamp) {
			event := heap.Pop(&s.events).(*Event)

			if event.IsDeleted() {
				continue
			}

			event.Execute()

			if event.IsRepeating() {
				s.Add(event)
			}

		} else {
			break
		}
	}

	s.ticks++
}

// Add adds an event to the planner
func (s *Planner) Add(event *Event) {
	s.mux.Lock()
	defer s.mux.Unlock()

	var expiry int64

	if event.IsImmediate() {
		expiry = 1
		event.Immediate(false)
	} else {
		expiry = int64(event.GetSecs() + s.ticks)
	}

	s.uniqueEvents[event.ID()] = event
	event.SetAbsExpiry(expiry)
	event.setPlanner(s)
	heap.Push(&s.events, event)
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

// Delete marks a Event to be deleted. Returns true if event
// found and marked for deletion, false if not.
func (s *Planner) Delete(event *Event) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	id := event.ID()

	if value, ok := s.uniqueEvents[id]; ok {
		value.Delete()
		delete(s.uniqueEvents, id)
		return true
	}

	return false
}
