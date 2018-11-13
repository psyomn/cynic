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

// ServiceQueue is a priority queue that sorts events that are to
// happen via their absolute expiry
type ServiceQueue []*Service

func (pq ServiceQueue) Len() int { return len(pq) }

func (pq ServiceQueue) Less(i, j int) bool {
	// Want lowest value here (smaller timestamp = sooner)
	return pq[i].priority < pq[j].priority
}

func (pq ServiceQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push inserts an service/event into the priority queue
func (pq *ServiceQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Service)
	item.index = n
	*pq = append(*pq, item)
}

// Pop retrieves the soonest event
func (pq *ServiceQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// PeekTimestamp gives the timestamp at the root of the heap
func (pq *ServiceQueue) PeekTimestamp() (int64, bool) {
	if len(*pq) == 0 {
		return 0, false
	}

	old := *pq
	item := old[0]
	return item.absExpiry, true
}

// PeekID returns the id of the service at root
func (pq *ServiceQueue) PeekID() (uint64, bool) {
	if len(*pq) == 0 {
		return 0, false
	}

	old := *pq
	item := old[0]
	return item.ID(), true
}
