/*
Package cynic_testing tests that it can monitor you from the ceiling.

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
	"testing"

	"github.com/psyomn/cynic"
)

func TestSimpleBuilder(t *testing.T) {
	setup := func(serviceCount, maxTime int) func(t *testing.T) {
		return func(t *testing.T) {
			var services []cynic.Service

			for i := 0; i < serviceCount; i++ {
				service := cynic.ServiceNew(1)
				services = append(services, service)
			}

			builder := cynic.ServiceBuilderNew(services)
			builder.DistributeEvents(maxTime)

			session, ok := builder.Build()
			assert(t, ok)

			for _, el := range session.Services {
				assert(t, el.GetSecs() == (maxTime/serviceCount))
			}
		}
	}

	type testCase struct {
		name     string
		serCount int
		maxTime  int
	}

	testCases := [...]testCase{
		testCase{"maxtime 5, service count 5", 5, 5},
		testCase{"maxtime 1000 service count 100", 100, 1000},
		testCase{"maxtime 999 service count 100", 100, 999},
	}

	for _, c := range testCases {
		t.Run(c.name, setup(c.serCount, c.maxTime))
	}
}

func TestSimpleErrorCases(t *testing.T) {
	setup := func(serviceCount, maxTime int) func(t *testing.T) {
		return func(t *testing.T) {
			var services []cynic.Service
			builder := cynic.ServiceBuilderNew(services)
			_, ok := builder.Build()
			assert(t, !ok)
		}
	}

	type testCase struct {
		name     string
		serCount int
		maxTime  int
	}

	tests := [...]testCase{
		testCase{"maxtime -10 service count 1", 1, -10},
		testCase{"maxtime 0 service count 10", 0, 10},
		testCase{"maxtime 10 service count 11", 10, 11},
	}

	for _, c := range tests {
		t.Run(c.name, setup(c.serCount, c.maxTime))
	}
}
