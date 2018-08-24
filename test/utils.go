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
	"io/ioutil"
	"log"
	"testing"
)

// FixturePathSimple is path to simple fixture for testing.
const FixturePathSimple = "./fixtures/simple.json"

// FixturePathStatus is Path to status fixture for testing.
const FixturePathStatus = "./fixtures/status.json"

// Assert is a simple helper to see if something is true, and if not
// raise failure.
func Assert(t *testing.T, val bool) {
	if !val {
		t.Fail()
	}
}

// ReadFixture simply reads a fixture and returns it as a string.
func ReadFixture(path string) string {
	contents, err := ioutil.ReadFile(path)

	if err != nil {
		log.Fatal("problem opening fixture ", path, ": ", err)
	}

	return string(contents[:])
}
