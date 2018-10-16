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
	"time"
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

	ticks int
}

// WheelNew creates a new, empty, timing wheel.
func WheelNew() *Wheel {
	var tw Wheel
	return &tw
}

// Tick moves the cursor of the timing wheel, by one second.
func (s *Wheel) Tick() {
	for _, service := range s.secs[s.secsCnt] {
		// TODO: this should be a different process, though
		//   getting the tests right in this respect is a little
		//   tricky for now
		service.Execute()
		if service.IsRepeating() {
			log.Println("> re-add secscnt: ", s.secsCnt)
			log.Println("> wheel-tick: ", s.ticks)
			log.Println("> service data: ", service)
			s.Add(service)
		}
	}

	s.secsCnt++

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

	log.Println("end-tick")
	log.Println("")
	s.ticks++
}

// Add puts a service in the timing wheel, with respect to its expiry
// time. The expiry time is taken as 'time_now' +
// service.seconds_to_expire
func (s *Wheel) Add(service *Service) {
	// TODO: verify
	seconds := s.secsCnt +
		s.minsCnt*wheelSecondsInMinute +
		s.hoursCnt*wheelSecondsInHour +
		s.daysCnt*wheelSecondsInDay +
		service.secs

	log.Println("### total-seconds", seconds)

	days := seconds / wheelSecondsInDay
	if days > 365 {
		log.Fatal("can't assign timers that are >365 days in the future")
	}

	seconds -= wheelSecondsInDay * days
	hours := seconds / wheelSecondsInHour
	seconds -= wheelSecondsInHour * hours
	minutes := seconds / wheelSecondsInMinute
	seconds -= wheelSecondsInMinute * minutes

	// TODO REMOVE ME
	log.Println("s:", seconds, "m:", minutes, "h:", hours, "d:", days)

	if service.IsImmediate() {
		seconds = 1
		service.Immediate(false)
		goto secondsAdd
	}

	if days > 0 {
		log.Println("wheel: adding in days") // TODO REMOVE
		index := (days % wheelMaxDays) - 1
		s.days[index] = append(s.days[index], service)
		return
	}

	if hours > 0 {
		index := (hours % wheelMaxHours) - 1
		log.Println("wheel: adding in hours: ", index) // TODO REMOVE
		s.hours[index] = append(s.hours[index], service)
		return
	}

	if minutes > 0 {
		index := (minutes % wheelMaxMins) // - 1 // TODO: rethink here

		log.Println("wheel: adding in minutes: ", index, " actual ", minutes) // TODO REMOVE
		s.mins[index] = append(s.mins[index], service)
		return
	}

secondsAdd:
	index := seconds
	log.Println("wheel: adding in seconds: ", index) // TODO REMOVE
	s.secs[index] = append(s.secs[index], service)
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

// Run runs the wheel, with a 1s tick
func (s *Wheel) Run() {
	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			s.Tick()
		}
	}()
	defer ticker.Stop()
}

//// -- Private --

func (s *Wheel) rotateMinutes() {
	log.Println("rotate minutes")
	for i := 0; i < len(s.secs); i++ {
		var tb timerBuff
		s.secs[i] = tb
	}

	log.Println("minutes: ", s.mins)

	// For everything in 98d:12:34:XX
	for _, el := range s.mins[s.minsCnt] {
		index := el.secs % wheelMaxSecs
		log.Println("placing from min to sec in index: ", index)
		s.secs[index] = append(s.secs[index], el)
	}
	log.Println("seconds: ", s.secs)
}

func (s *Wheel) rotateHours() {
	log.Println("rotate hours")
	for i := 0; i < len(s.secs); i++ {
		var tb timerBuff
		s.secs[i] = tb
	}

	for i := 0; i < len(s.mins); i++ {
		var tb timerBuff
		s.mins[i] = tb
	}

	// for everything in 98d:12:XX:XX
	for _, el := range s.hours[s.hoursCnt] {
		index := ((el.secs % wheelSecondsInHour) / wheelSecondsInMinute)
		if index == 0 {
			// dealing with seconds
			secIndex := (el.secs % wheelSecondsInHour)
			s.secs[secIndex] = append(s.secs[secIndex], el)
		} else {
			index--
			s.mins[index] = append(s.mins[index], el)
		}
	}
}

func (s *Wheel) rotateDays() {
	log.Println("rotate days")
	for i := 0; i < len(s.secs); i++ {
		var tb timerBuff
		s.secs[i] = tb
	}

	for i := 0; i < len(s.mins); i++ {
		var tb timerBuff
		s.mins[i] = tb
	}

	for i := 0; i < len(s.hours); i++ {
		var tb timerBuff
		s.hours[i] = tb
	}

	// for everything in 98d:XX:XX:XX
	for _, el := range s.days[s.daysCnt] {
		// Place in hours
		hourIndex := ((el.secs % wheelSecondsInDay) / wheelSecondsInHour)
		if hourIndex > 0 {
			hourIndex--
			s.hours[hourIndex] = append(s.hours[hourIndex], el)
			continue
		}

		// Place in minutes
		minuteIndex := ((el.secs % wheelSecondsInHour) / wheelSecondsInMinute)
		if minuteIndex > 0 {
			minuteIndex--
			s.mins[minuteIndex] = append(s.mins[minuteIndex], el)
			continue
		}

		// Place in seconds
		secondIndex := (el.secs % wheelSecondsInMinute)
		s.secs[secondIndex] = append(s.secs[secondIndex], el)
	}
}
