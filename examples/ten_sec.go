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
	"github.com/psyomn/cynic"
	"log"
)

func main() {
	var events []cynic.Event
	event := cynic.EventNew(10)
	event.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		log.Println("tick")
		return false, 0
	})

	events = append(events, event)

	session := cynic.Session{
		Events: events,
	}

	cynic.Start(session)
}

// output
// $ ./examples/ten_sec
// 2019/02/08 15:04:30 tick
