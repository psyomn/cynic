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
	"strconv"
)

type timerBuff = []*Service

const (
	wheelMaxDays  = 365 // 1..364
	wheelMaxHours = 24  // 1..23
	wheelMaxMins  = 60  // 1..59
	wheelMaxSecs  = 60  // 0..59

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
	for _, service := range s.secs[s.secsCnt] {
		// TODO: worker pool will be much nicer here
		for _, hook := range service.hooks {
			hook(nil, nil)
		}
	}

	s.secsCnt++

	// TODO: wheel rotation should be invoking timer placements here
	if s.secsCnt >= wheelMaxSecs {
		s.secsCnt = 0
		s.rotateMinutes()
		s.minsCnt++
	}

	if s.minsCnt >= wheelMaxMins {
		s.minsCnt = 0
		s.rotateHours()
		s.hoursCnt++
	}

	if s.hoursCnt >= wheelMaxHours {
		s.hoursCnt = 0
		s.rotateDays()
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
		index := (days % wheelMaxDays) - 1
		s.days[index] = append(s.days[index], service)
		return
	}

	if hours > 0 {
		index := (hours % wheelMaxHours) - 1
		s.hours[index] = append(s.hours[index], service)
		return
	}

	if minutes > 0 {
		index := (minutes % wheelMaxMins) - 1
		s.mins[index] = append(s.mins[index], service)
		return
	}

	if seconds > 0 {
		index := seconds
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
	for i := 0; i < len(s.secs); i++ {
		var tb timerBuff
		s.secs[i] = tb
	}

	// For everything in 98d:12:34:XX
	for _, el := range s.mins[s.minsCnt] {
		index := el.Secs % wheelMaxSecs
		s.secs[index] = append(s.secs[index], el)
	}
}

func (s *Wheel) rotateHours() {
	for i := 0; i < len(s.mins); i++ {
		var tb timerBuff
		s.mins[i] = tb
	}

	// for everything in 98d:12:XX:XX
	for _, el := range s.hours[s.hoursCnt] {
		index := ((el.Secs % wheelSecondsInHour) / wheelSecondsInMinute)
		s.mins[index] = append(s.mins[index], el)
	}

	s.rotateMinutes()
}

func (s *Wheel) rotateDays() {
	// for everything in 98d:XX:XX:XX
	for i := 0; i < len(s.hours); i++ {
		// TODO: maybe there's improvements here to be made
		var tb timerBuff
		s.hours[i] = tb
	}

	// for everything in 98d:XX:XX:XX
	for _, el := range s.days[s.daysCnt] {
		index := ((el.Secs % wheelSecondsInDay) / wheelSecondsInHour)
		log.Println("day index: ", index)
		s.mins[index] = append(s.mins[index], el)
	}

	s.rotateHours()
}

// String makes a nice, printable format of the wheel, and timer
// counts
func (s *Wheel) String() string {
	var str string

	str += "s| "
	for _, el := range s.secs {
		str += strconv.Itoa(len(el)) + " "
	}
	str += "\n"

	str += "m| "
	for _, el := range s.mins {
		str += strconv.Itoa(len(el)) + " "
	}
	str += "\n"

	str += "h| "
	for _, el := range s.hours {
		str += strconv.Itoa(len(el))
		str += " "
	}
	str += "\n"

	str += "d| "
	for _, el := range s.days {
		str += strconv.Itoa(len(el))
		str += " "
	}
	str += "\n"

	return str
}
