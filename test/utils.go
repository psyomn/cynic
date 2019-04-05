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
)

// Assert is a simple helper to see if something is true, and if not
// raise failure.
func assert(t *testing.T, ok bool, args ...interface{}) {
	if len(args) == 0 && !ok {
		t.Fail()
	}

	if len(args) == 1 && !ok {
		t.Fatalf("%s", args[0])
	}

	if len(args) > 1 && !ok {
		format, ok := args[0].(string)
		if !ok {
			panic("what do you think you're doing")
		}

		t.Fatalf(format, args[1:]...)
	}
}
