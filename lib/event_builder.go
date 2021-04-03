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

type distributionParams struct {
	maxTime int
}

// EventBuilder is a helper to set properties to a lot of
// events. For example, if you have 10 events you want to run
// within 100 seconds, you can use this builder in oder to disperse
// everything over 10 seconds.
type EventBuilder struct {
	events []Event

	evenDistribute bool
	allRepeatable  bool

	distribution *distributionParams
}

// EventBuilderNew creates a new events builder. Simple
// configurations. If you want something more complex, you should do
// it on your own.
func EventBuilderNew(events []Event) EventBuilder {
	return EventBuilder{
		events:         events,
		evenDistribute: false,
		allRepeatable:  false,
		distribution:   nil,
	}
}

// Build takes all the things you gave the builder, puts them
// together, and gives you a session object to do whatever you
// will with it.
func (s *EventBuilder) Build() (Session, bool) {
	ret := s.makeRepeatable() && s.makeDistributeEvents()

	sess := Session{
		Events:  s.events,
		Alerter: nil,
	}

	return sess, ret
}

// DistributeEvents over a max time interval.
func (s *EventBuilder) DistributeEvents(maxTime int) {
	s.distribution = &distributionParams{
		maxTime: maxTime,
	}
}

func (s *EventBuilder) makeDistributeEvents() bool {
	if s.distribution == nil ||
		s.distribution.maxTime <= 0 ||

		// min granularity is a sec, so 11 events in 10 secs
		// do not guarantee some sort of distribution
		len(s.events) > s.distribution.maxTime {
		return false
	}

	eventCount := len(s.events)
	interval := s.distribution.maxTime / eventCount

	for i := 0; i < eventCount; i++ {
		s.events[i].SetSecs(interval)
		s.events[i].SetOffset(interval * i)
	}

	return true
}

// Repeatable will mark all events as repeatable.
func (s *EventBuilder) Repeatable() {
	s.allRepeatable = true
}

func (s *EventBuilder) makeRepeatable() bool {
	for _, el := range s.events {
		el.Repeat(true)
	}
	return true
}
