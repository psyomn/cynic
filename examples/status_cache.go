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

	"github.com/psyomn/cynic"
)

func main() {
	// you can open this in: http://localhost:9999/status
	// also useful link: http://localhost:9999/links
	status := cynic.StatusServerNew("", "9999", cynic.DefaultStatusEndpoint)

	var events []cynic.Event
	event := cynic.EventNew(1)
	event.AddHook(func(params *cynic.HookParameters) (bool, interface{}) {
		log.Println("tick")

		params.Status.Update("hello", "there")
		params.Status.Update("stuff", map[string]interface{}{
			"timestamp": time.Now().Unix(),
		})

		return false, 0
	})
	event.Repeat(true)
	event.SetDataRepo(&status)

	events = append(events, event)

	session := cynic.Session{
		Events:      events,
		StatusCache: &status,
	}

	cynic.Start(session)
}

// output
// ./examples/every_ten_sec
// 2019/02/08 15:05:45 tick
// 2019/02/08 15:05:55 tick
// 2019/02/08 15:06:05 tick
// 2019/02/08 15:06:15 tick
// 2019/02/08 15:06:25 tick
