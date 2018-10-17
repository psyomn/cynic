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

	s.ticks++
}

// Add puts a service in the timing wheel, with respect to its expiry
// time. The expiry time is taken as 'time_now' +
// service.seconds_to_expire
func (s *Wheel) Add(service *Service) {
	// TODO:
	//   so there should be an absolute here, that is added
	//   with the services that will be added. this is because the
	//   wheels are simple arrays, and don't have any complex
	//   rotations, like dynamic circular buffers

	seconds := s.secsCnt + service.secs
	service.AbsSecs(seconds)

	days := seconds / wheelSecondsInDay
	if days > wheelMaxDays {
		log.Fatal("can't assign timers that are >365 days in the future")
	}

	seconds -= wheelSecondsInDay * days
	hours := seconds / wheelSecondsInHour
	seconds -= wheelSecondsInHour * hours
	minutes := seconds / wheelSecondsInMinute
	seconds -= wheelSecondsInMinute * minutes

	// log.Println(
	// 	"\n### Adding with times: ###########################\n",
	// 	"# seconds: ", seconds, ", seccnt: ", s.secsCnt, "\n",
	// 	"# minutes: ", minutes, ", minscnt: ", s.minsCnt, "\n",
	// 	"# hours: ", hours, " hourscnt: ", s.hoursCnt, "\n",
	// 	"# days: ", days, " dayscnt: ", s.daysCnt, "\n",
	// 	"# ", s,
	// 	"###################################################")

	if service.IsImmediate() {
		seconds = 1
		service.Immediate(false)
		goto secondsAdd
	}

	if days > 0 {
		index := (s.daysCnt+days)%wheelMaxDays - 1
		s.days[index] = append(s.days[index], service)
		return
	}

	if hours > 0 {
		index := (s.hoursCnt+hours)%wheelMaxHours - 1
		s.hours[index] = append(s.hours[index], service)
		return
	}

	if minutes > 0 {
		index := (s.minsCnt+minutes)%wheelMaxMins - 1 // TODO: rethink here
		s.mins[index] = append(s.mins[index], service)
		return
	}

secondsAdd:
	index := seconds
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
	for i := 0; i < len(s.secs); i++ {
		var tb timerBuff
		s.secs[i] = tb
	}

	// For everything in 98d:12:34:XX
	for _, el := range s.mins[s.minsCnt] {
		index := 0
		if el.IsRepeating() {
			index = el.GetAbsSecs() % wheelMaxSecs
		} else {
			index = el.secs % wheelMaxSecs
		}

		s.secs[index] = append(s.secs[index], el)
	}
}

func (s *Wheel) rotateHours() {
	// log.Println("rotate hours") // TODO REMOVE
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
	// log.Println("rotate days") // TODO REMOVE
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
