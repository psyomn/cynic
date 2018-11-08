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
	wheel := cynic.WheelNew()

	// Test most edge cases
	serviceSecs := cynic.ServiceJSONNew("www.google.com", 1)
	serviceMaxSecs := cynic.ServiceJSONNew("www.google.com", 59)

	serviceMinute := cynic.ServiceJSONNew("www.google.com", 60)
	serviceMaxMinute := cynic.ServiceJSONNew("www.google.com", 60*60-1)

	serviceHour := cynic.ServiceJSONNew("www.google.com", 60*60)
	serviceMaxHour := cynic.ServiceJSONNew("www.google.com", 23*60*60+60*59+59) // 23:59:59

	service := cynic.ServiceJSONNew("www.google.com", 3*60*60+33*60+33)

	services := [...]cynic.Service{
		serviceSecs,
		serviceMaxSecs,
		serviceMinute,
		serviceMaxMinute,
		serviceHour,
		serviceMaxHour,
		service,
	}

	for _, el := range services {
		wheel.Add(&el)
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
			service := cynic.ServiceNew(time)
			service.AddHook(func(_ *cynic.StatusServer) (_ bool, _ interface{}) {
				isExpired = true
				return false, 0
			})

			assert(t, !isExpired)

			wheel := cynic.WheelNew()
			wheel.Add(&service)

			for i := 0; i < time; i++ {
				wheel.Tick()
				if isExpired {
					log.Println("expired before its time")
				}
				assert(t, !isExpired)
			}

			wheel.Tick()
			if !isExpired {
				log.Println(wheel)
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

func TestAddRepeatedService(t *testing.T) {
	var count int
	var time int
	time = 10

	service := cynic.ServiceJSONNew("www.google.com", time)
	service.Repeat(true)
	service.AddHook(func(_ *cynic.StatusServer) (_ bool, _ interface{}) {
		count++
		return false, 0
	})

	wheel := cynic.WheelNew()
	wheel.Add(&service)

	n := 3
	for i := 0; i < (time*n)+1; i++ {
		wheel.Tick()
	}

	assert(t, count == n)
}

func TestTickSeconds(t *testing.T) {
	setupTimeTest := func(totalSecs, sec, min, hour, day int) func(t *testing.T) {
		return func(t *testing.T) {
			tw := cynic.WheelNew()

			for i := 1; i <= totalSecs; i++ {
				tw.Tick()
			}

			// assert(t, tw.Seconds() == sec)
			// assert(t, tw.Minutes() == min)
			// assert(t, tw.Hours() == hour)
			// assert(t, tw.Days() == day)
		}
	}

	type wheelTestCase struct {
		name  string
		total int
		sec   int
		min   int
		hour  int
		day   int
	}

	cases := []wheelTestCase{
		wheelTestCase{"15 seconds", 15, 15, 0, 0, 0},
		wheelTestCase{"1 minute", 60, 0, 1, 0, 0},
		wheelTestCase{"1 minute 1 second", 61, 1, 1, 0, 0},
		wheelTestCase{"10 minutes", 60 * 10, 0, 10, 0, 0},
		wheelTestCase{"61 minutes", 60 * 61, 0, 1, 1, 0},
		wheelTestCase{"120 minutes", 60 * 60 * 2, 0, 0, 2, 0},
		wheelTestCase{"121 minutes", 2*hour + 1*minute, 0, 1, 2, 0},
	}

	for _, c := range cases {
		t.Run(c.name, setupTimeTest(c.total, c.sec, c.min, c.hour, c.day))
	}
}

func TestAddTickThenAddAgain(t *testing.T) {
	var s1, s2 int
	wheel := cynic.WheelNew()
	service := cynic.ServiceJSONNew("www.google.com", 10)
	service.AddHook(
		func(_ *cynic.StatusServer) (_ bool, _ interface{}) {
			s1 = 1
			return false, 0
		})

	wheel.Add(&service)

	wheel.Tick()
	wheel.Tick()
	wheel.Tick()

	assert(t, s1 == 0 && s2 == 0)

	nextService := cynic.ServiceJSONNew("www.HAHAHA.com", 10)
	nextService.AddHook(
		func(_ *cynic.StatusServer) (_ bool, _ interface{}) {
			s2 = 1
			return false, 0
		})

	wheel.Add(&nextService)

	for i := 0; i < 8; i++ {
		wheel.Tick()
	}

	assert(t, s1 == 1 && s2 == 0)

	for i := 0; i < 4; i++ {
		wheel.Tick()
	}

	assert(t, s1 == 1 && s2 == 1)
}

func TestServiceOffset(t *testing.T) {
	secs := 3
	offsetTime := 2
	ran := false
	wheel := cynic.WheelNew()

	s := cynic.ServiceNew(secs)
	s.Offset(offsetTime)

	wheel.Add(&s)

	s.AddHook(func(_ *cynic.StatusServer) (_ bool, _ interface{}) {
		ran = true
		return false, 0
	})

	assert(t, !ran)

	wheel.Tick()
	wheel.Tick()
	assert(t, !ran)

	for i := 0; i < secs; i++ {
		wheel.Tick()
	}

	assert(t, ran)
}

func TestServiceImmediate(t *testing.T) {
	// TODO: Test with immediate and a long long time afterwards, eg:
	//   immediate + 5 hours in the future
	//   immediate + 3 days in the future
	var count int
	time := 12
	s := cynic.ServiceNew(time)
	s.Immediate(true)
	s.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	w := cynic.WheelNew()
	w.Add(&s)

	w.Tick()
	w.Tick()
	assert(t, count == 1)

	for i := 0; i < time*10; i++ {
		w.Tick()
	}

	assert(t, count == 1)
}

func TestServiceImmediateWithRepeat(t *testing.T) {
	var count int
	time := 12

	s := cynic.ServiceNew(time)
	s.Immediate(true)
	s.Repeat(true)
	s.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	w := cynic.WheelNew()
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

	ser := cynic.ServiceNew(1)
	ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	w := cynic.WheelNew()

	// for {
	// 	 if w.Tick(); w.Seconds() == 30 {
	// 	 	break
	// 	 }
	// }
	w.Add(&ser)

	w.Tick()
	w.Tick()
	assert(t, count == 1)
}

func TestAddLastMinuteSecond(t *testing.T) {
	var count int

	ser := cynic.ServiceNew(1)
	ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	w := cynic.WheelNew()

	// for {
	// 	w.Tick()
	// 	if w.Seconds() == 58 {
	// 		break
	// 	}
	// }
	w.Add(&ser)

	w.Tick() // expire 58
	w.Tick() // expire 59

	assert(t, count == 1)
}

func TestRepeatedTicks(t *testing.T) {
	var count int
	ser := cynic.ServiceNew(1)
	ser.Repeat(true)
	ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	w := cynic.WheelNew()
	w.Add(&ser)

	upto := 30

	// set cursor on top of first service
	w.Tick()

	for i := 0; i < upto; i++ {
		w.Tick()
	}

	assert(t, count == 30)
}

func TestSimpleRepeatedRotation(t *testing.T) {
	var count int
	ser := cynic.ServiceNew(1)
	label := "simple-repeated-rotation-x3"

	ser.Label = &label
	ser.Repeat(true)
	ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
		count++
		return false, 0
	})

	w := cynic.WheelNew()

	// for {
	// 	 if w.Tick(); w.Seconds() == 58 {
	// 	 	break
	// 	 }
	// }

	w.Add(&ser)
	log.Println(w)
	log.Println("==========================")

	// Test first rotation
	w.Tick()
	w.Tick()
	if count != 1 {
		log.Println("failed at first rotation")
	}
	assert(t, count == 1)

	// Test second rotation
	// for {
	// 	if w.Tick(); w.Seconds() == 59 {
	// 	 	break
	// 	 }
	// }

	w.Tick()
	if count != 61 {
		log.Println("failed at second rotation")
		log.Println("expected count 61, but got: ", count)
		log.Println(w)
	}
	assert(t, count == 61)

	// Test third rotation
	// for {
	// 	if w.Tick(); w.Seconds() == 59 {
	// 		break
	// 	}
	// }

	log.Println("count: ", count)
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
			ser := cynic.ServiceNew(interval)
			ser.Repeat(true)
			ser.AddHook(func(_ *cynic.StatusServer) (bool, interface{}) {
				// log.Print(".")
				count++
				return false, 0
			})

			w := cynic.WheelNew()
			w.Add(&ser)
			w.Tick() // put cursor on top of just inserted timer

			for i := 0; i < timerange-interval; i++ {
				// log.Println("Tick : ", i)
				w.Tick()
			}

			expectedCount := (timerange - interval) / interval
			if expectedCount != count {
				log.Println("##### ", t.Name())
				log.Println("interval:       ", interval)
				log.Println("timerange:      ", timerange)
				log.Println("expected ticks: ", expectedCount)
				log.Println("actual ticks:   ", count)
				log.Println("abs secs:       ", ser.GetAbsSecs())
				log.Println("wheel: \n", w)
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
		// testCase{"1 sec within 3 hour", 1 * second, 3 * hour},

		testCase{"10 sec within 1 min", 10 * second, 1 * minute},
		testCase{"10 sec within 2 min", 10 * second, 2 * minute},
		testCase{"10 sec within 3 min", 10 * second, 3 * minute},
		testCase{"13 sec within 2 min", 13 * second, 2 * minute},

		// days
		// testCase{"1 sec within 1 day", 1 * second, 1 * day},
		// testCase{"2 sec within 1 day", 2 * second, 1 * day},
		// testCase{"33 sec within 1 day", 33 * second, 1 * day},
		// testCase{"43 sec within 1 day", 43 * second, 1 * day},
		// testCase{"53 sec within 1 day", 53 * second, 1 * day},
		// testCase{"10 minutes within 1 day", 10 * minute, 1 * day},
		// testCase{"1 hour within 1 week", 1 * hour, 1 * week},

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
