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
)

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

func (pq *ServiceQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Service)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *ServiceQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *ServiceQueue) Timestamp() int64 {
	old := *pq
	n := len(old)
	item := old[n-1]
	return item.absExpiry
}

// update modifies the priority and value of an Item in the queue.
func (pq *ServiceQueue) update(item *Service, priority int) {
	item.priority = priority
	heap.Fix(pq, item.index)
}

// PeekTimestamp gives the timestamp at the root of the heap
func (pq *ServiceQueue) PeekTimestamp() int64 {
	old := *pq
	n := len(old)
	item := old[n-1]
	return item.absExpiry
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
