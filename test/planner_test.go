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
	"sync"
	"sync/atomic"
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

	eventSecs := cynic.EventNew(1 * second)
	eventMaxSecs := cynic.EventNew(59 * second)

	eventMinute := cynic.EventNew(1 * minute)
	eventMaxMinute := cynic.EventNew(1*hour - 1)

	eventHour := cynic.EventNew(1 * hour)
	eventMaxHour := cynic.EventNew(23*hour + 59*minute + 59*second) // 23:59:59

	event := cynic.EventNew(3*hour + 33*minute + 33*second)

	events := [...]cynic.Event{
		eventSecs,
		eventMaxSecs,
		eventMinute,
		eventMaxMinute,
		eventHour,
		eventMaxHour,
		event,
	}

	for i := 0; i < len(events); i++ {
		planner.Add(&events[i])
	}

	assert(t, len(events) == planner.Len())
}

func TestTickAll(t *testing.T) {
	setupAddTickTest := func(givenTime int) func(t *testing.T) {
		// take a time and assert that the timer is not expired, up to
		// the n-1 time interval. Test that it is finally expired
		// after the final time interval.
		return func(t *testing.T) {
			var wg sync.WaitGroup

			isExpired := false

			time := givenTime
			event := cynic.EventNew(time)

			wg.Add(1)
			event.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
				defer wg.Done()

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
			wg.Wait()

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
	var wg sync.WaitGroup
	var count int
	time := 10
	n := 3

	event := cynic.EventNew(time)
	event.Repeat(true)
	wg.Add(n)

	event.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg.Done()

		count++
		return false, 0
	})

	planner := cynic.PlannerNew()
	planner.Add(&event)

	for i := 0; i < (time*n)+1; i++ {
		planner.Tick()
	}

	wg.Wait()
	assert(t, count == n)
}

func TestAddTickThenAddAgain(t *testing.T) {
	var s1, s2 int
	var wg1, wg2 sync.WaitGroup

	planner := cynic.PlannerNew()
	event := cynic.EventNew(10)

	wg1.Add(1)
	event.AddHook(
		func(_ *cynic.HookParameters) (bool, interface{}) {
			defer wg1.Done()

			s1 = 1
			return false, 0
		})

	planner.Add(&event)

	planner.Tick()
	planner.Tick()
	planner.Tick()

	assert(t, s1 == 0 && s2 == 0)

	nextEvent := cynic.EventNew(10)

	wg2.Add(1)
	nextEvent.AddHook(
		func(_ *cynic.HookParameters) (bool, interface{}) {
			defer wg2.Done()

			s2 = 1
			return false, 0
		})

	planner.Add(&nextEvent)

	for i := 0; i < 8; i++ {
		planner.Tick()
	}

	wg1.Wait()
	assert(t, s1 == 1 && s2 == 0)

	for i := 0; i < 4; i++ {
		planner.Tick()
	}

	wg2.Wait()
	assert(t, s1 == 1 && s2 == 1)
}

func TestEventOffset(t *testing.T) {
	var wg sync.WaitGroup
	secs := 3
	offsetTime := 2
	ran := false

	s := cynic.EventNew(secs)
	s.SetOffset(offsetTime)

	wg.Add(1)
	s.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg.Done()

		ran = true
		return false, 0
	})

	planner := cynic.PlannerNew()
	planner.Add(&s)
	planner.Tick()

	assert(t, !ran)

	planner.Tick()
	planner.Tick()

	assert(t, !ran)

	for i := 0; i < secs; i++ {
		planner.Tick()
	}

	wg.Wait()
	assert(t, ran)
}

func TestEventImmediate(t *testing.T) {
	setup := func(givenTime int) func(t *testing.T) {
		return func(t *testing.T) {
			var count int
			var wg sync.WaitGroup
			time := givenTime
			s := cynic.EventNew(time)

			s.Immediate(true)
			wg.Add(1)
			s.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
				defer wg.Done()

				count++
				return false, 0
			})

			w := cynic.PlannerNew()
			w.Add(&s)

			w.Tick()
			w.Tick()
			wg.Wait()
			assert(t, count == 1)

			for i := 0; i < time*10; i++ {
				w.Tick()
			}

			assert(t, count == 1)
		}
	}

	type testCase struct {
		name string
		time int
	}

	testCases := [...]testCase{
		{"3 seconds", 3 * second},
		{"3 hours", 3 * hour},
		{"3 days", 3 * day},
	}

	for _, tc := range testCases {
		t.Run(tc.name, setup(tc.time))
	}
}

func TestEventImmediateWithRepeat(t *testing.T) {
	var wg sync.WaitGroup
	var count int
	time := 12

	s := cynic.EventNew(time)
	s.Immediate(true)
	s.Repeat(true)
	wg.Add(1)
	s.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg.Done()

		count++
		return false, 0
	})

	w := cynic.PlannerNew()
	w.Add(&s)

	w.Tick()
	w.Tick()
	wg.Wait()
	assert(t, count == 1)

	wg.Add(1) // due to repeat
	for i := 0; i < time; i++ {
		w.Tick()
	}

	wg.Wait()
	assert(t, count == 2)
}

func TestAddHalfMinute(t *testing.T) {
	var wg sync.WaitGroup
	var count int

	ser := cynic.EventNew(1)
	wg.Add(1)
	ser.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg.Done()

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
	wg.Wait()

	assert(t, count == 1)
}

func TestAddLastMinuteSecond(t *testing.T) {
	var count int
	var wg sync.WaitGroup

	ser := cynic.EventNew(1)
	wg.Add(1)
	ser.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg.Done()
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
	wg.Wait()

	assert(t, count == 1)
}

func TestRepeatedTicks(t *testing.T) {
	var count int
	var wg sync.WaitGroup
	ser := cynic.EventNew(1)
	upto := 30

	wg.Add(upto)
	ser.Repeat(true)
	ser.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg.Done()

		count++
		return false, 0
	})

	w := cynic.PlannerNew()
	w.Add(&ser)

	// set cursor on top of first event
	w.Tick()

	for i := 0; i < upto; i++ {
		w.Tick()
	}

	wg.Wait()
	assert(t, count == 30)
}

func TestSimpleRepeatedRotation(t *testing.T) {
	var wg sync.WaitGroup
	var count uint32
	ser := cynic.EventNew(1)

	ser.Repeat(true)
	ser.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg.Done()
		atomic.AddUint32(&count, 1)
		return false, 0
	})

	w := cynic.PlannerNew()

	{
		for i := 0; i < 59; i++ {
			w.Tick()
		}
		wg.Add(1)
		w.Add(&ser)

		w.Tick() // place on top of event and ...
		w.Tick() // ... execute event
		wg.Wait()

		assert(t, count == 1, "first rotation: %d", count)
	}

	{
		wg.Add(60)
		for i := 0; i < 59; i++ {
			w.Tick()
		}
		w.Tick()
		wg.Wait()

		assert(
			t, count == 61,
			"second rotation: expected count 61, but got: %d, \n\nPlanner info: %v",
			count, w,
		)
	}

	{
		// Test third rotation
		wg.Add(60)
		for i := 0; i < 60; i++ {
			w.Tick()
		}
		wg.Wait()

		assert(t, count == 121,
			"third rotation: expected count 121, but got: %d\n\nPlanner: %v\n",
			count,
			w,
		)
	}
}

func TestRepeatedRotationTables(t *testing.T) {
	setup := func(interval, timerange int) func(t *testing.T) {
		return func(t *testing.T) {
			var wg sync.WaitGroup
			var count uint32

			ser := cynic.EventNew(interval)
			ser.Repeat(true)

			ser.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
				defer wg.Done()

				atomic.AddUint32(&count, 1)
				return false, 0
			})

			w := cynic.PlannerNew()
			w.Add(&ser) // TODO this has to be on top
			w.Tick()    // place position in the inclusive time ranxge

			expectedCount := (timerange - interval) / interval
			wg.Add(expectedCount)
			for i := 0; i < timerange-interval; i++ {
				w.Tick()
			}

			// wg.Wait()
			log.Println(&wg)
			if expectedCount != int(count) {
				log.Println("##### ", t.Name())
				log.Println("interval:       ", interval)
				log.Println("timerange:      ", timerange)
				log.Println("expected ticks: ", expectedCount)
				log.Println("actual ticks:   ", count)
				log.Println("planner: \n", w)
			}

			assert(t, int(count) == expectedCount)
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
	var wg1, wg2 sync.WaitGroup

	planner := cynic.PlannerNew()
	ser := cynic.EventNew(1)
	ser2 := cynic.EventNew(1)

	ser.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg1.Done()
		expire1 = true
		return false, 0
	})

	ser2.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg2.Done()
		expire2 = true
		return false, 0
	})

	wg1.Add(1)
	planner.Add(&ser)

	wg2.Add(1)
	planner.Add(&ser2)

	assert(t, planner.Delete(&ser))
	assert(t, ser.IsDeleted())
	assert(t, !ser2.IsDeleted())

	planner.Tick()
	planner.Tick()
	wg2.Wait()

	// Make sure that the deleted event does not ever execute,
	// since marked for deletion before tick
	assert(t, !expire1)
	assert(t, expire2)
}

func TestSecondsApart(t *testing.T) {
	var wg1, wg2, wg3 sync.WaitGroup
	s1 := cynic.EventNew(1)
	s2 := cynic.EventNew(2)
	s3 := cynic.EventNew(3)
	pl := cynic.PlannerNew()

	run := [...]bool{false, false, false}

	s1.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg1.Done()
		run[0] = true
		return false, 0
	})
	s2.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg2.Done()
		run[1] = true
		return false, 0
	})
	s3.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg3.Done()
		run[2] = true
		return false, 0
	})

	s1.Repeat(true)
	s2.Repeat(true)
	s3.Repeat(true)

	wg1.Add(1)
	wg2.Add(1)
	wg3.Add(1)

	pl.Add(&s1)
	pl.Add(&s2)
	pl.Add(&s3)

	pl.Tick()

	pl.Tick()
	wg1.Wait()
	assert(t, run[0] && !run[1] && !run[2])
	run = [...]bool{false, false, false}

	pl.Tick()
	wg2.Wait()
	assert(t, run[0] && run[1] && !run[2])
	run = [...]bool{false, false, false}

	pl.Tick()
	wg3.Wait()
	assert(t, run[0] && !run[1] && run[2])
}

func TestChainAddition(t *testing.T) {
	var wg sync.WaitGroup
	s1 := cynic.EventNew(1)
	s2 := cynic.EventNew(1)
	s3 := cynic.EventNew(1)
	s4 := cynic.EventNew(1)
	run := [...]bool{false, false, false, false}

	hook := func(e *cynic.Event, r *bool) cynic.HookSignature {
		return func(params *cynic.HookParameters) (bool, interface{}) {
			defer wg.Done()
			log.Println("ASDF")

			if params == nil {
				t.Fatal("hook params are nil")
			}

			if params.Planner == nil {
				t.Fatal("planner should not be nil")
			}

			if e != nil {
				log.Println("add event")
				params.Planner.Add(e)
			}

			*r = true

			return false, 0
		}
	}

	s1.AddHook(hook(&s2, &run[0]))
	s2.AddHook(hook(&s3, &run[1]))
	s3.AddHook(hook(&s4, &run[2]))
	s4.AddHook(hook(nil, &run[3]))

	planner := cynic.PlannerNew()

	wg.Add(1)
	planner.Add(&s1)
	planner.Tick()
	assert(t, !(run[0] || run[1] || run[2] || run[3]))

	for i := 0; i < 4; i++ {
		log.Println("tick")
		planner.Tick()
	}

	assert(t, (run[0] && run[1] && run[2] && run[3]))
}

func TestMultipleEventsAndHooks(t *testing.T) {
	var wg sync.WaitGroup
	var count uint32
	const max = 10

	hk := func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg.Done()

		atomic.AddUint32(&count, 1)
		return false, 0
	}

	planner := cynic.PlannerNew()
	for i := 0; i < max; i++ {
		newEvent := cynic.EventNew(1)

		// Add the hook twice, for realsies
		newEvent.AddHook(hk)
		newEvent.AddHook(hk)
		wg.Add(2)

		planner.Add(&newEvent)
	}

	planner.Tick() // place cursor
	planner.Tick() // should execute
	wg.Wait()

	assert(t, count == 20)
}

func TestImmediateWithOffset(t *testing.T) {
	var wg sync.WaitGroup
	var count int
	offset := 5
	eventTime := 10

	event := cynic.EventNew(eventTime)
	event.Immediate(true)
	event.SetOffset(offset)
	event.Repeat(true)
	event.AddHook(func(_ *cynic.HookParameters) (bool, interface{}) {
		defer wg.Done()

		count++
		return false, 0
	})

	planner := cynic.PlannerNew()
	wg.Add(1)
	planner.Add(&event)

	// This means that it should tick:
	// - at first tick (seconds = 1 + 5) -> due to offset
	// - after 10 seconds (absolute time = 16 seconds)

	// should not have counted yet
	assert(t, count == 0)

	// Everything upto the offset is zero
	for i := 0; i < offset; i++ {
		planner.Tick()
		assert(t, count == 0)
	}
	planner.Tick()
	wg.Wait()
	assert(t, count == 1)

	wg.Add(1)
	for i := 0; i < eventTime; i++ {
		planner.Tick()
	}
	wg.Wait()
	assert(t, count == 2)
}
