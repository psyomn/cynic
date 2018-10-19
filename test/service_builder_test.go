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
	// maxTime := 1000
	// serviceCount := 10

	cynic.ServiceNew(10)

	// var services []cynic.Service

	// for i := 0; i < serviceCount; i++ {
	// 	service := cynic.ServiceNew(1)
	// 	services := append(services, service)
	// }

	// builder := cynic.ServiceBuilderNew(services)
}
