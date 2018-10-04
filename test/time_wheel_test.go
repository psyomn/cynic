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
	serviceSecs := cynic.ServiceNew("www.google.com", 1)
	serviceMaxSecs := cynic.ServiceNew("www.google.com", 59)

	serviceMinute := cynic.ServiceNew("www.google.com", 60)
	serviceMaxMinute := cynic.ServiceNew("www.google.com", 60*60-1)

	serviceHour := cynic.ServiceNew("www.google.com", 60*60)
	serviceMaxHour := cynic.ServiceNew("www.google.com", 23*60*60+60*59+59) // 23:59:59

	service := cynic.ServiceNew("www.google.com", 3*60*60+33*60+33)

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
			service := cynic.ServiceNew("www.google.com", time)
			service.AddHook(func(_ *cynic.AddressBook, _ interface{}) (_ bool, _ interface{}) {
				isExpired = true
				return false, 0
			})

			assert(t, !isExpired)

			wheel := cynic.WheelNew()
			wheel.Add(&service)

			for i := 0; i < time; i++ {
				wheel.Tick()
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
		// TODO: eventually this should be supported. This is
		//   panicking because this inits a ticker with value 0,
		//   which makes no sense.
		// tickTestCase{"0 seconds", 0 * second},
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
		tickTestCase{"1 hour 1 minute", 1*hour + 1*minute},
		tickTestCase{"1 hour 1 second", 1*hour + 1*second},
		tickTestCase{"1 hour 59 second", 1*hour + 59*second},
		tickTestCase{"1 hour 59 minute", 1*hour + 59*minute},
		tickTestCase{"1 hour 59 minute 59 second", 1*hour + 59*minute + 59*second},
		tickTestCase{"23 hour", 23 * hour},
		tickTestCase{"1 day", 1 * day},
		tickTestCase{"1 day 1 second", 1*day + 1*second},
		tickTestCase{"1 day 59 second", 1*day + 59*second},
		tickTestCase{"1 week", 7 * day},
	}

	for _, c := range cases {
		t.Run(c.name, setupAddTickTest(c.time))
	}

}

func TestAddRepeatedService(t *testing.T) {
	var count int
	var time int
	time = 10

	service := cynic.ServiceNew("www.google.com", time)
	service.Repeat(true)
	service.AddHook(func(_ *cynic.AddressBook, _ interface{}) (_ bool, _ interface{}) {
		count++
		return false, 0
	})

	wheel := cynic.WheelNew()
	wheel.Add(&service)

	n := 3
	for i := 0; i < time*n; i++ {
		wheel.Tick()
	}

	assert(t, count == n)
}

type wheelTestCase struct {
	name  string
	total int
	sec   int
	min   int
	hour  int
	day   int
}

func setupTimeTest(totalSecs, sec, min, hour, day int) func(t *testing.T) {
	return func(t *testing.T) {
		tw := cynic.WheelNew()

		for i := 1; i <= totalSecs; i++ {
			tw.Tick()
		}

		assert(t, tw.Seconds() == sec)
		assert(t, tw.Minutes() == min)
		assert(t, tw.Hours() == hour)
		assert(t, tw.Days() == day)
	}
}

func TestTickSeconds(t *testing.T) {
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

func TestMinuteRotation(t *testing.T) {
	t.Fatal("implement")
}

func TestHourRotation(t *testing.T) {
	t.Fatal("implement")
}

func TestDayRotation(t *testing.T) {
	t.Fatal("implement")
}