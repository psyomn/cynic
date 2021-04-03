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

type alertInfo struct {
	Name      string
	Desc      string
	Timestamp int64
}

func alerter(alerts []cynic.AlertMessage) {
	for _, alert := range alerts {
		info := alert.Response.(alertInfo)
		log.Println("ALERT: ", info.Name, ": ", info.Desc)
		log.Println("  problematic timestamp: ", info.Timestamp)
	}
}

func main() {
	var events []cynic.Event
	event := cynic.EventNew(1)
	event.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		log.Println("tick")
		timestamp := time.Now().Unix()
		shouldAlert := timestamp&1 == 1
		return shouldAlert, alertInfo{
			Name:      "VERY BAD ERROR",
			Desc:      "Welp, that was unfortunate",
			Timestamp: timestamp,
		}
	})
	event.Repeat(true)

	events = append(events, event)

	alertConfig := cynic.AlerterNew(60, alerter)

	session := cynic.Session{
		Events:  events,
		Alerter: &alertConfig,
	}

	cynic.Start(session)
}

// output
// ...
// 2019/02/08 16:33:18 tick
// 2019/02/08 16:33:19 tick
// 2019/02/08 16:33:20 tick
// 2019/02/08 16:33:21 tick
// 2019/02/08 16:33:21 ALERT:  VERY BAD ERROR :  Welp, that was unfortunate
// 2019/02/08 16:33:21   problematic timestamp:  1549661543
// 2019/02/08 16:33:21 ALERT:  VERY BAD ERROR :  Welp, that was unfortunate
// 2019/02/08 16:33:21   problematic timestamp:  1549661545
// 2019/02/08 16:33:21 ALERT:  VERY BAD ERROR :  Welp, that was unfortunate
// 2019/02/08 16:33:21   problematic timestamp:  1549661547
// 2019/02/08 16:33:21 ALERT:  VERY BAD ERROR :  Welp, that was unfortunate
// 2019/02/08 16:33:21   problematic timestamp:  1549661549
// ...
