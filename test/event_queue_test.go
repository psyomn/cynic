/*
Package cynic_testing tests that it can monitor you from the ceiling.

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
	"container/heap"
	"testing"

	"github.com/psyomn/cynic"
)

func makeEventQueue() cynic.EventQueue {
	events := make(cynic.EventQueue, 0)
	heap.Init(&events)
	return events
}

func TestEventQueueTimestamp(t *testing.T) {
	events := makeEventQueue()

	s1 := cynic.EventNew(10)
	s2 := cynic.EventNew(2)
	s3 := cynic.EventNew(15)

	ss := [...]cynic.Event{s1, s2, s3}

	for i := 0; i < len(ss); i++ {
		heap.Push(&events, &ss[i])
	}

	heap.Init(&events)

	{
		expectedID := s2.ID()
		actualID, ok := events.PeekID()

		assert(t, ok)
		assert(t, expectedID == actualID)
	}

	{
		s4 := cynic.EventNew(1)
		expectedID := s4.ID()
		heap.Push(&events, &s4)

		actualID, ok := events.PeekID()

		assert(t, ok)
		assert(t, expectedID == actualID)
	}

}

func TestPeekEmpty(t *testing.T) {
	events := makeEventQueue()
	_, ok := events.PeekID()
	assert(t, !ok)
}

func BenchmarkAdditionsPerSecond(b *testing.B) {
	events := make([]*cynic.Event, 0)

	numNodes := 5000
	for i := 0; i < numNodes; i++ {
		ser := cynic.EventNew(i + 1)
		events = append(events, &ser)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eventq := makeEventQueue()
		for j := 0; j < numNodes; j++ {
			heap.Push(&eventq, events[j])
		}
	}
}
