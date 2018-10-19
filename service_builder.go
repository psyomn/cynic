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

type distributionParams struct {
	maxTime int
}

// ServiceBuilder is a helper to set properties to a lot of
// services. For example, if you have 10 services you want to run
// within 100 seconds, you can use this builder in oder to disperse
// everything over 10 seconds.
type ServiceBuilder struct {
	services []Service

	evenDistribute bool
	allRepeatable  bool

	distribution *distributionParams
}

func ServiceBuilderNew(services []Service) ServiceBuilder {
	return ServiceBuilder{
		services:       services,
		evenDistribute: false,
		allRepeatable:  false,
	}
}

// Build takes all the things you gave the builder, puts them
// together, and gives you a session object to do whatever you
// will with it
func (s *ServiceBuilder) Build() Session {
	s.makeRepeatable()

	sess := Session{
		Services:  s.services,
		Alerter:   nil,
		AlertTime: 0,
	}
	return sess
}

// Distribute events over a time
func (s *ServiceBuilder) EventDistribute(maxTime int) {
	s.distribution = &distributionParams{
		maxTime: maxTime,
	}

	serviceCount := len(s.services)
	interval := s.distribution.maxTime / serviceCount
	for index, el := range s.services {
		el.SetSecs(interval)
		el.Offset(interval * index)
	}
}

// Repeatable will mark all services as repeatable
func (s *ServiceBuilder) Repeatable() {
	s.allRepeatable = true
}

func (s *ServiceBuilder) makeRepeatable() {
	for _, el := range s.services {
		el.Repeat(true)
	}
}
