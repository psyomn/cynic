// +build ignore

/*
Example code on cynic usage.

Copyright 2019 Simon Symeonidis (psyomn)

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
package main

import (
	"log"
	"time"

	"github.com/psyomn/cynic/lib"
)

func main() {
	var events []cynic.Event
	event := cynic.EventNew(10)
	event.AddHook(func(params *cynic.HookParameters) (bool, interface{}) {
		log.Println("tick")
		params.Status.Update("time at tick", time.Now().Unix())
		return false, 0
	})
	event.Repeat(true)

	statusCache := cynic.StatusServerNew("", "9999", cynic.DefaultStatusEndpoint)
	event.SetDataRepo(&statusCache)

	events = append(events, event)

	session := cynic.Session{
		Events:      events,
		StatusCache: &statusCache,
		SnapshotConfig: &cynic.SnapshotConfig{
			Interval:  time.Minute,
			DumpEvery: time.Minute * 2,
			Path:      "./",
		},
	}

	cynic.Start(session)
}

// output
// ./snapshot
// ...
// 2019/02/11 13:51:21 tick
// 2019/02/11 13:51:31 tick
// 2019/02/11 13:51:41 tick
// 2019/02/11 13:51:51 tick
// 2019/02/11 13:52:01 tick
// 2019/02/11 13:52:11 tick
// 2019/02/11 13:52:21 tick
// 2019/02/11 13:52:31 tick
// 2019/02/11 13:52:41 tick
// 2019/02/11 13:52:51 tick
// 2019/02/11 13:53:01 tick
// ...

// After some time, snapshots are dumped on disk
//
// ls *.cynic
// 2019-02-11T13:32:00-05:00.1.cynic
// 2019-02-11T13:38:00-05:00.1.cynic
// 2019-02-11T13:44:00-05:00.1.cynic
// 2019-02-11T13:50:00-05:00.1.cynic
// 2019-02-11T13:34:00-05:00.1.cynic
// 2019-02-11T13:40:00-05:00.1.cynic
// 2019-02-11T13:46:00-05:00.1.cynic
// 2019-02-11T13:52:00-05:00.1.cynic
// 2019-02-11T13:36:00-05:00.1.cynic
// 2019-02-11T13:42:00-05:00.1.cynic
// 2019-02-11T13:48:00-05:00.1.cynic

// cynic comes with a very small tool which can parse most of these
//
// cynic-store
// $ cynic-store -input 2019-02-11T13:32:00-05:00.1.cynic
// version: 1
// 1549909860:{"time at tick":1549909851}
// 1549909920:{"time at tick":1549909911}
