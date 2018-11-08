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
	"log"
	"testing"

	"github.com/psyomn/cynic"
)

func makeServiceQueue() cynic.ServiceQueue {
	services := make(cynic.ServiceQueue, 0)
	heap.Init(&services)
	return services
}

func TestServiceQueueTimestamp(t *testing.T) {
	services := makeServiceQueue()

	s1 := cynic.ServiceNew(10)
	s2 := cynic.ServiceNew(2)
	s3 := cynic.ServiceNew(15)

	ss := [...]cynic.Service{s1, s2, s3}

	for i := 0; i < len(ss); i++ {
		heap.Push(&services, &ss[i])
	}

	heap.Init(&services)

	{
		expectedID := s2.ID()
		actualID, ok := services.PeekID()

		assert(t, ok)
		assert(t, expectedID == actualID)
	}

	{
		s4 := cynic.ServiceNew(1)
		expectedID := s4.ID()
		heap.Push(&services, &s4)

		actualID, ok := services.PeekID()

		assert(t, ok)
		assert(t, expectedID == actualID)
	}

}

func BenchmarkAdditionsPerSecond(b *testing.B) {
	serviceq := makeServiceQueue()
	services := make([]*Service, 0)

	numNodes := 5000
	for i := 0; i < numNodes; i++ {
		services = append(services, &cynic.ServiceNew(i))
	}

	b.Resettimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < numNodes; j++ {
			heap.Push(&serviceq, services[j])
		}
	}
}
