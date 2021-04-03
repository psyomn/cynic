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
	"time"
)

// AlertFunc defines the hook signature for alert messages.
type AlertFunc = func([]AlertMessage)

// Alerter is an entity that ticks, and if there are alert messages,
// will fire up behavior.
type Alerter struct {
	alerts     []AlertMessage
	Ch         chan AlertMessage
	stopCh     chan int
	waitTime   int
	waitTicker *time.Ticker
	alerterFn  AlertFunc
}

// AlertMessage defines a simple alert structure that can be used by
// users of the library, and decide how to show information about the
// alerts.
type AlertMessage struct {
	Response      interface{} `json:"response_text"`
	Now           string      `json:"now"`
	CynicHostname string      `json:"cynic_hostname"`
}

// AlerterNew creates a new alerter.
func AlerterNew(waitTime int, alerter AlertFunc) Alerter {
	var alerts []AlertMessage
	ch := make(chan AlertMessage)
	stop := make(chan int)
	ticker := time.NewTicker(time.Second * time.Duration(waitTime))

	return Alerter{
		alerts:     alerts,
		Ch:         ch,
		stopCh:     stop,
		waitTime:   waitTime,
		waitTicker: ticker,
		alerterFn:  alerter,
	}
}

// Start begins the alerter.
func (s *Alerter) Start() {
	go s.run()
}

// Stop the alerter.
func (s *Alerter) Stop() {
	s.stopCh <- 0
}

func (s *Alerter) run() {
	defer s.waitTicker.Stop()

	for {
		select {
		case recvAlert := <-s.Ch:
			s.alerts = append(s.alerts, recvAlert)
		case <-s.waitTicker.C:
			if len(s.alerts) > 0 {
				s.alerterFn(s.alerts)
			}
			var clear []AlertMessage
			s.alerts = clear
		case <-s.stopCh:
			return
		}
	}
}
