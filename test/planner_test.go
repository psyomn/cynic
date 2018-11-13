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
package cynictesting

import (
	"log"
	"testing"

	"github.com/psyomn/cynic"
)

const (
	second = 1
	minute = 60
	hour   = minute * 60
	day    = 24 * hour
	week   = 7 * day
	month  = 30 * day
	year   = 12 * month
)

func TestAdd(t *testing.T) {
	planner := cynic.PlannerNew()

	// Test most edge cases
	eventSecs := cynic.EventJSONNew("www.google.com", 1)
	eventMaxSecs := cynic.EventJSONNew("www.google.com", 59)

	eventMinute := cynic.EventJSONNew("www.google.com", 60)
	eventMaxMinute := cynic.EventJSONNew("www.google.com", 60*60-1)

	eventHour := cynic.EventJSONNew("www.google.com", 60*60)
	eventMaxHour := cynic.EventJSONNew("www.google.com", 23*60*60+60*59+59) // 23:59:59

	event := cynic.EventJSONNew("www.google.com", 3*60*60+33*60+33)

	events := [...]cynic.Event{
		eventSecs,
		eventMaxSecs,
		eventMinute,
		eventMaxMinute,
		eventHour,
		eventMaxHour,
		event,
	}

	for _, el := range events {
		planner.Add(&el)
	}
}

func TestTickAll(t *testing.T) {
	// take a time and assert that the timer is not expired, up to
	// the n-1 time interval. Test that it is finally expired
	// after the final time interval.
	setupAddTickTest := func(givenTime int) func(t *testing.T) {
		return func(t *testing.T) {
			isExpired := false

			time := givenTime
			event := cynic.EventNew(time)
			event.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
				isExpired = true
				return false, 0
			})

			assert(t, !isExpired)

			planner := cynic.PlannerNew()
			planner.Add(&event)

			for i := 0; i < time; i++ {
				planner.Tick()
				if isExpired {
					log.Println("expired before its time")
				}
				assert(t, !isExpired)
			}

			planner.Tick()
			if !isExpired {
				log.Println(planner)
				log.Println(event)
				log.Println(event.GetAbsExpiry())
			}

			assert(t, isExpired)
		}
	}

	type tickTestCase struct {
		name string
		time int
	}

	cases := [...]tickTestCase{
		tickTestCase{"1 second", 1 * second},
		tickTestCase{"10 seconds", 10 * second},
		tickTestCase{"59 seconds", 59 * second},
		tickTestCase{"just 1 minute", 60 * second},
		tickTestCase{"1 min 1 sec", 1*minute + 1*second},
		tickTestCase{"1 min 30 sec", 1*minute + 30*second},
		tickTestCase{"1 min 59 sec", 1*minute + 59*second},
		tickTestCase{"2 minutes", 2 * minute},
		tickTestCase{"2 minutes 1 second", 2*minute + 1},
		tickTestCase{"3 minutes", 3 * minute},
		tickTestCase{"10 minutes", 10 * minute},
		tickTestCase{"10 minutes 1 second", 10*minute + 1},
		tickTestCase{"1 hour", 1 * hour},
		tickTestCase{"1 hour 1 second", 1*hour + 1*second},
		tickTestCase{"1 hour 1 minute", 1*hour + 1*minute},
		tickTestCase{"1 hour 1 minute 1 second", 1*hour + 1*minute + 1*second},
		tickTestCase{"1 hour 59 second", 1*hour + 59*second},
		tickTestCase{"1 hour 59 minute", 1*hour + 59*minute},
		tickTestCase{"1 hour 59 minute 59 second", 1*hour + 59*minute + 59*second},
		tickTestCase{"23 hour", 23 * hour},
		tickTestCase{"1 day", 1 * day},
		tickTestCase{"1 day 1 second", 1*day + 1*second},
		tickTestCase{"1 day 59 second", 1*day + 59*second},
		tickTestCase{"1 week", 7 * day},
		tickTestCase{"1 week 1 sec", 7*day + 1*second},
		tickTestCase{"1 week 15 minutes", 7*day + 15*minute},
		tickTestCase{"1 month 1 hour", 1*month + 1*hour},
		tickTestCase{"11 months", 11 * month},
	}

	for _, c := range cases {
		t.Run(c.name, setupAddTickTest(c.time))
	}

}

func TestAddRepeatedEvent(t *testing.T) {
	var count int
	var time int
	time = 10

	event := cynic.EventJSONNew("www.google.com", time)
	event.Repeat(true)
	event.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	planner := cynic.PlannerNew()
	planner.Add(&event)

	n := 3
	for i := 0; i < (time*n)+1; i++ {
		planner.Tick()
	}

	assert(t, count == n)
}

func TestAddTickThenAddAgain(t *testing.T) {
	var s1, s2 int
	planner := cynic.PlannerNew()
	event := cynic.EventJSONNew("www.google.com", 10)
	event.AddHook(
		func(_ *cynic.StatusServer) (bool, interface{}) {
			s1 = 1
			return false, 0
		})

	planner.Add(&event)

	planner.Tick()
	planner.Tick()
	planner.Tick()

	assert(t, s1 == 0 && s2 == 0)

	nextEvent := cynic.EventJSONNew("www.HAHAHA.com", 10)
	nextEvent.AddHook(
		func(_ *cynic.StatusServer) (bool, interface{}) {
			s2 = 1
			return false, 0
		})

	planner.Add(&nextEvent)

	for i := 0; i < 8; i++ {
		planner.Tick()
	}

	assert(t, s1 == 1 && s2 == 0)

	for i := 0; i < 4; i++ {
		planner.Tick()
	}

	assert(t, s1 == 1 && s2 == 1)
}

func TestEventOffset(t *testing.T) {
	secs := 3
	offsetTime := 2
	ran := false
	planner := cynic.PlannerNew()

	s := cynic.EventNew(secs)
	s.Offset(offsetTime)

	planner.Add(&s)

	s.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		ran = true
		return false, 0
	})

	assert(t, !ran)

	planner.Tick()
	planner.Tick()
	assert(t, !ran)

	for i := 0; i < secs; i++ {
		planner.Tick()
	}

	assert(t, ran)
}

func TestEventImmediate(t *testing.T) {
	// TODO: Test with immediate and a long long time afterwards, eg:
	//   immediate + 5 hours in the future
	//   immediate + 3 days in the future
	var count int
	time := 12
	s := cynic.EventNew(time)
	s.Immediate(true)
	s.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	w := cynic.PlannerNew()
	w.Add(&s)

	w.Tick()
	w.Tick()
	assert(t, count == 1)

	for i := 0; i < time*10; i++ {
		w.Tick()
	}

	assert(t, count == 1)
}

func TestEventImmediateWithRepeat(t *testing.T) {
	var count int
	time := 12

	s := cynic.EventNew(time)
	s.Immediate(true)
	s.Repeat(true)
	s.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	w := cynic.PlannerNew()
	w.Add(&s)

	w.Tick()
	w.Tick()

	assert(t, count == 1)

	for i := 0; i < time; i++ {
		w.Tick()
	}

	assert(t, count == 2)
}

func TestAddHalfMinute(t *testing.T) {
	var count int

	ser := cynic.EventNew(1)
	ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	w := cynic.PlannerNew()

	countTicks := 0
	for {
		if w.Tick(); countTicks == 30 {
			break
		}
		countTicks++
	}
	w.Add(&ser)

	w.Tick()
	w.Tick()
	assert(t, count == 1)
}

func TestAddLastMinuteSecond(t *testing.T) {
	var count int

	ser := cynic.EventNew(1)
	ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	w := cynic.PlannerNew()

	countTicks := 0
	for {
		w.Tick()
		countTicks++
		if countTicks == 58 {
			break
		}
	}
	w.Add(&ser)

	w.Tick() // expire 58
	w.Tick() // expire 59

	assert(t, count == 1)
}

func TestRepeatedTicks(t *testing.T) {
	var count int
	ser := cynic.EventNew(1)
	ser.Repeat(true)
	ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	w := cynic.PlannerNew()
	w.Add(&ser)

	upto := 30

	// set cursor on top of first event
	w.Tick()

	for i := 0; i < upto; i++ {
		w.Tick()
	}

	assert(t, count == 30)
}

func TestSimpleRepeatedRotation(t *testing.T) {
	var count int
	ser := cynic.EventNew(1)
	label := "simple-repeated-rotation-x3"

	ser.Label = &label
	ser.Repeat(true)
	ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	w := cynic.PlannerNew()
	var totalTicks int
	for {
		totalTicks++
		if w.Tick(); totalTicks == 58 {
			break
		}
	}

	w.Add(&ser)

	// Test first rotation
	w.Tick()
	w.Tick()
	if count != 1 {
		log.Println("failed at first rotation")
	}
	assert(t, count == 1)

	totalTicks = 0
	for {
		totalTicks++
		if w.Tick(); totalTicks == 59 {
			break
		}
	}

	w.Tick()
	if count != 61 {
		log.Println("failed at second rotation")
		log.Println("expected count 61, but got: ", count)
		log.Println(w)
	}
	assert(t, count == 61)

	// Test third rotation
	totalTicks = 0
	for {
		totalTicks++
		if w.Tick(); totalTicks == 59 {
			break
		}
	}

	w.Tick()

	if count != 121 {
		log.Println("failed at third rotation")
		log.Println("expected count 121, but got: ", count)
		log.Println(w)
	}
	assert(t, count == 121)
}

func TestRepeatedRotationTables(t *testing.T) {
	setup := func(interval, timerange int) func(t *testing.T) {
		return func(t *testing.T) {
			var count int
			ser := cynic.EventNew(interval)
			ser.Repeat(true)
			ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
				count++
				return false, 0
			})

			w := cynic.PlannerNew()
			w.Add(&ser)
			w.Tick() // put cursor on top of just inserted timer

			for i := 0; i < timerange-interval; i++ {
				w.Tick()
			}

			expectedCount := (timerange - interval) / interval
			if expectedCount != count {
				log.Println("##### ", t.Name())
				log.Println("interval:       ", interval)
				log.Println("timerange:      ", timerange)
				log.Println("expected ticks: ", expectedCount)
				log.Println("actual ticks:   ", count)
				log.Println("planner: \n", w)
			}
			assert(t, count == expectedCount)
		}
	}

	type testCase struct {
		name      string
		interval  int
		timerange int
	}

	testCases := []testCase{
		testCase{"1 sec within 1 min", 1 * second, 1 * minute},
		testCase{"1 sec within 1 min 1 sec", 1 * second, 1*minute + 1*second},
		testCase{"2 sec within 1 min 1 sec", 2 * second, 1*minute + 1*second},
		testCase{"1 sec within 1 min 30 sec", 1 * second, 1*minute + 30*second},
		testCase{"1 sec within 2 min", 1 * second, 2 * minute},
		testCase{"1 sec within 3 min", 1 * second, 3 * minute},
		testCase{"1 sec within 4 min", 1 * second, 4 * minute},
		testCase{"1 sec within 5 min", 1 * second, 5 * minute},
		testCase{"1 sec within 1 hour", 1 * second, 1 * hour},
		testCase{"59 sec within 10 min", 59 * second, 10 * minute},
		testCase{"60 sec within 10 min", 60 * second, 10 * minute},
		testCase{"1 sec within 3 hour", 1 * second, 3 * hour},

		testCase{"10 sec within 1 min", 10 * second, 1 * minute},
		testCase{"10 sec within 2 min", 10 * second, 2 * minute},
		testCase{"10 sec within 3 min", 10 * second, 3 * minute},
		testCase{"13 sec within 2 min", 13 * second, 2 * minute},

		// days
		testCase{"1 sec within 1 day", 1 * second, 1 * day},
		testCase{"2 sec within 1 day", 2 * second, 1 * day},
		testCase{"33 sec within 1 day", 33 * second, 1 * day},
		testCase{"43 sec within 1 day", 43 * second, 1 * day},
		testCase{"53 sec within 1 day", 53 * second, 1 * day},
		testCase{"10 minutes within 1 day", 10 * minute, 1 * day},
		testCase{"1 hour within 1 week", 1 * hour, 1 * week},

		testCase{"1 hour within 1 day", 1 * hour, 1 * day},
		testCase{"4 hours within 1 day", 4 * hour, 1 * day},

		// weeks
		testCase{"1 day in 1 week", 1 * day, 1 * week},
		testCase{"2 days in 1 week", 2 * day, 1 * week},

		testCase{"1 week in 1 month", 1 * week, 1 * month},
		testCase{"1 month in 1 year", 1 * month, 1 * year},
	}

	for _, tc := range testCases {
		t.Run(tc.name, setup(tc.interval, tc.timerange))
	}
}

func TestPlannerDelete(t *testing.T) {
	var expire1, expire2 bool

	planner := cynic.PlannerNew()
	ser := cynic.EventNew(1)
	ser2 := cynic.EventNew(1)

	ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		expire1 = true
		return false, 0
	})

	ser2.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		expire2 = true
		return false, 0
	})

	planner.Add(&ser)
	planner.Add(&ser2)

	assert(t, planner.Delete(&ser))
	assert(t, ser.IsDeleted())
	assert(t, !ser2.IsDeleted())

	planner.Tick()
	planner.Tick()

	// Make sure that the deleted event does not ever execute,
	// since marked for deletion before tick
	assert(t, expire1 == false)
	assert(t, expire2 == true)
}

func TestSecondsApart(t *testing.T) {
	s1 := cynic.EventNew(1)
	s2 := cynic.EventNew(2)
	s3 := cynic.EventNew(3)
	pl := cynic.PlannerNew()

	run := [...]bool{false, false, false}

	s1.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		run[0] = true
		return false, 0
	})
	s2.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		run[1] = true
		return false, 0
	})
	s3.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		run[2] = true
		return false, 0
	})

	s1.Repeat(true)
	s2.Repeat(true)
	s3.Repeat(true)

	pl.Add(&s1)
	pl.Add(&s2)
	pl.Add(&s3)

	pl.Tick()

	pl.Tick()
	assert(t, run[0] && !run[1] && !run[2])
	run = [...]bool{false, false, false}

	pl.Tick()
	assert(t, run[0] && run[1] && !run[2])
	run = [...]bool{false, false, false}

	pl.Tick()
	assert(t, run[0] && !run[1] && run[2])
}
