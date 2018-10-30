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

	// timestamp is used for calculating in which bucket to put a
	// timer. timestamp only gets updated every minute rotation.
	timestamp int64
}

// WheelNew creates a new, empty, timing wheel.
func WheelNew() *Wheel {
	var tw Wheel
	tw.timestamp = time.Now().Unix()
	return &tw
}

// Tick moves the cursor of the timing wheel, by one second.
func (s *Wheel) Tick() {
	// log.Println("Tick: ", s.ticks)

	for _, service := range s.secs[s.secsCnt] {
		// TODO: this should be a different process, though
		//   getting the tests right in this respect is a little
		//   tricky for now
		// log.Println("Run service/event: ", service.UniqStr())
		service.Execute()
		if service.IsRepeating() {
			// log.Println("Re-Add service")
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

func dayOffset(timestamp int64) int64 {
	return timestamp + int64(wheelSecondsInDay)
}

func hourOffset(timestamp int64) int64 {
	return timestamp + int64(wheelSecondsInHour)
}

func minuteOffset(timestamp int64) int64 {
	return timestamp + int64(wheelSecondsInMinute)
}

// Add a service/event into the time wheel
func (s *Wheel) Add(service *Service) {
	service.SetAbsExpiry(s.timestamp + int64(s.secsCnt))

	seconds := service.GetSecs()
	// timestamp where the event is supposed to occur
	absEventTime := s.timestamp + int64(s.secsCnt) + int64(seconds)

	// Calculate bucket
	var dayIndex, hourIndex, minuteIndex, secondIndex int
	diff := int(absEventTime - s.timestamp) // diff is total time to expiry

	// STUFF TODO
	// if days > wheelMaxDays {
	// 	log.Fatal("can't assign timers that are >365 days in the future")
	// }

	if service.IsImmediate() {
		secondIndex = 1
		service.Immediate(false)
		goto immediate
	}

	{
		// which bucket does it belong to?
		if absEventTime >= dayOffset(s.timestamp) {
			dayIndex = diff / wheelSecondsInDay
			dayIndex += s.daysCnt
		} else if absEventTime >= hourOffset(s.timestamp) {
			hourIndex = diff / wheelSecondsInHour
			hourIndex += s.hoursCnt
		} else if absEventTime >= minuteOffset(s.timestamp) {
			minuteIndex = diff / wheelSecondsInMinute
			minuteIndex += s.minsCnt
		} else /* seconds */ {
			secondIndex = diff // this has s.secsCnt, see diff declaration above
		}

		// Check if we've gone over a minute, and evoked a domino effect
		if secondIndex >= wheelMaxSecs {
			// no s.secsCnt because added before via diff
			secondIndex = secondIndex % wheelMaxSecs
			minuteIndex++
		}
		if minuteIndex >= wheelMaxMins {
			minuteIndex = minuteIndex % wheelMaxMins
			hourIndex++
		}
		if hourIndex >= wheelMaxHours {
			hourIndex = hourIndex % wheelMaxHours
			dayIndex++
		}
		if dayIndex >= wheelMaxDays {
			// TODO recheck me
			dayIndex = dayIndex % wheelMaxDays
		}
	}

	if true {
		log.Println("######################################")
		log.Println("# Insert into indices: ###############")
		log.Println("# timestamp: ", s.timestamp)
		log.Println("# days indx: ", dayIndex)
		log.Println("# hour indx: ", hourIndex)
		log.Println("#   hourcnt : ", s.hoursCnt)
		log.Println("# minute ix: ", minuteIndex)
		log.Println("# second ix: ", secondIndex)
		log.Println("######################################")
		log.Println("# wheel: \n", s)
		log.Println("######################################")
	}

	if dayIndex > 0 {
		index := dayIndex%wheelMaxDays - 1
		s.days[index] = append(s.days[index], service)
		return
	}

	if hourIndex > 0 {
		index := hourIndex%wheelMaxHours - 1
		log.Println("INSERT IN HOUR: ", index)
		s.hours[index] = append(s.hours[index], service)
		return
	}

	if minuteIndex > 0 {
		index := minuteIndex%wheelMaxMins - 1
		log.Println("minute index: ", index)
		s.mins[index] = append(s.mins[index], service)
		return
	}

immediate:
	if secondIndex > 0 {
		index := secondIndex
		s.secs[index] = append(s.secs[index], service)
		// log.Println(s)
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

	// log.Println(">>>>> rotate minutes ")

	// For everything in 98d:12:34:XX
	for _, el := range s.mins[s.minsCnt] {
		index := 0
		if el.IsRepeating() {
			index = int(el.GetAbsExpiry()-s.timestamp) % wheelMaxSecs
			// log.Println("repeating index: ", index)
		} else {
			index = int(el.GetAbsExpiry()-s.timestamp) % wheelMaxSecs
		}
		// log.Println(s.secs)
		s.secs[index] = append(s.secs[index], el)
		// log.Println(s.secs)
	}

	s.timestamp += wheelSecondsInMinute
}

func (s *Wheel) rotateHours() {
	log.Println("rotate hours", s.hoursCnt)
	for i := 0; i < len(s.secs); i++ {
		var tb timerBuff
		s.secs[i] = tb
	}

	for i := 0; i < len(s.mins); i++ {
		var tb timerBuff
		s.mins[i] = tb
	}

	log.Println("contents of hours: ", s.hours)
	// for everything in 98d:12:XX:XX
	for _, el := range s.hours[s.hoursCnt] {
		// May rotate over abs expiry, and it can be quite bigger that way...
		// I think this is wrong...
		index := (el.GetAbsExpiry() - s.timestamp) % wheelSecondsInHour / wheelSecondsInMinute
		log.Println("index:", index)

		if index == 0 {
			// dealing with seconds
			secIndex := (el.GetAbsExpiry() - s.timestamp) % wheelSecondsInMinute
			log.Println("place in seconds", secIndex)
			s.secs[secIndex] = append(s.secs[secIndex], el)
		} else {
			log.Println("place in minutes")
			index--
			s.mins[index] = append(s.mins[index], el)
		}
	}
}

func (s *Wheel) rotateDays() {
	log.Println("rotate days", s.daysCnt)
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
		hourIndex := ((el.GetAbsExpiry() - s.timestamp) % wheelSecondsInDay) / wheelSecondsInHour

		log.Println("hourrotate: place second index")
		if hourIndex > 0 {
			hourIndex--
			s.hours[hourIndex] = append(s.hours[hourIndex], el)
			continue
		}

		// Place in minutes
		log.Println("minuterotate: place second index")
		minuteIndex := ((el.GetAbsExpiry() - s.timestamp) % wheelSecondsInHour) / wheelSecondsInMinute
		if minuteIndex > 0 {
			minuteIndex--
			s.mins[minuteIndex] = append(s.mins[minuteIndex], el)
			continue
		}

		// Place in seconds
		log.Println("dayraotate: place second index")
		secondIndex := (el.GetAbsExpiry() - s.timestamp) % wheelSecondsInMinute
		s.secs[secondIndex] = append(s.secs[secondIndex], el)
	}
}
