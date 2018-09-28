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
	"log"
)

type timerBuff = []*Service

const (
	wheelMaxDays  = 365
	wheelMaxHours = 24
	wheelMaxMins  = 60
	wheelMaxSecs  = 60

	wheelSecondsInMinute = 60
	wheelSecondsInHour   = wheelSecondsInMinute * 60
	wheelSecondsInDay    = wheelSecondsInHour * 24
	wheelSecondsInYear   = wheelSecondsInDay * 365

	// min interval is one second
	wheelMinInterval = 1
)

// Wheel is a timing wheel, which stores the services to be run at
// different times.
type Wheel struct {
	secs  [wheelMaxSecs]timerBuff
	mins  [wheelMaxMins]timerBuff
	hours [wheelMaxHours]timerBuff
	days  [wheelMaxDays]timerBuff

	secsCnt  int
	minsCnt  int
	hoursCnt int
	daysCnt  int
}

// WheelNew creates a new, empty, timing wheel.
func WheelNew() *Wheel {
	var tw Wheel
	return &tw
}

// Tick moves the cursor of the timing wheel, by one second.
func (s *Wheel) Tick() {
	// TODO: excecution code goes here of expired counters
	for _, service := range s.secs[s.secsCnt] {
		// TODO: worker pool will be much nicer here
		for _, hook := range service.hooks {
			hook(nil, nil)
		}
	}

	s.secsCnt++

	// TODO: wheel rotation should be invoking timer placements here
	if s.secsCnt >= wheelMaxSecs {
		s.rotateMinutes()
		s.secsCnt = 0
		s.minsCnt++
	}

	if s.minsCnt >= wheelMaxMins {
		s.rotateHours()
		s.minsCnt = 0
		s.hoursCnt++
	}

	if s.hoursCnt >= wheelMaxHours {
		s.rotateDays()
		s.hoursCnt = 0
		s.daysCnt++
	}

	if s.daysCnt >= wheelMaxDays {
		// NEW year, wowoowowowowowoowow
		s.daysCnt = 0
	}
}

// Add puts a service in the timing wheel, with respect to its expiry
// time. The expiry time is taken as 'time_now' + service.seconds_to_expire
func (s *Wheel) Add(service *Service) {
	seconds := service.Secs

	days := seconds / wheelSecondsInDay
	if days > 365 {
		log.Fatal("can't assign timers that are >365 days in the future")
	}

	seconds -= wheelSecondsInDay * days
	hours := seconds / wheelSecondsInHour
	seconds -= wheelSecondsInHour * hours
	minutes := seconds / wheelSecondsInMinute
	seconds -= wheelSecondsInMinute * minutes

	if days > 0 {
		index := (days + s.daysCnt) % wheelMaxDays
		log.Println("days index: ", index)
		s.days[index] = append(s.days[index], service)
		return
	}

	if hours > 0 {
		index := (hours + s.hoursCnt) % wheelMaxHours
		log.Println("hors index: ", index)
		s.hours[index] = append(s.hours[index], service)
		return
	}

	if minutes > 0 {
		index := (minutes + s.minsCnt) % wheelMaxMins
		log.Println("minutes index: ", index)
		s.mins[index] = append(s.mins[index], service)
		return
	}

	if seconds > 0 {
		index := seconds - 1
		log.Println("seconds index: ", index)
		s.secs[index] = append(s.secs[index], service)
		return
	}
}

// Seconds get the seconds of the timing wheel
func (s *Wheel) Seconds() int {
	return s.secsCnt
}

// Minutes get the minutes of the timing wheel
func (s *Wheel) Minutes() int {
	return s.minsCnt
}

// Hours get the the hours of the timing wheel
func (s *Wheel) Hours() int {
	return s.hoursCnt
}

// Days get the days of the timing wheel
func (s *Wheel) Days() int {
	return s.daysCnt
}

func (s *Wheel) rotateMinutes() {
	// For everything in 98d:12:34:XX
	// TODO deletion/clearin of counters
}

func (s *Wheel) rotateHours() {
	// for everything in 98d:12:XX:XX
	// TODO deletion/clearin of counters
}

func (s *Wheel) rotateDays() {
	// for everything in 98d:XX:XX:XX
	// TODO deletion/clearin of counters
}
