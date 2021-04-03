/*
Copyright 2018-2021 Simon Symeonidis (psyomn)

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
package test

import (
	"context"
	"net/http"
	"testing"
)

// Assert is a simple helper to see if something is true, and if not
// raise failure.
func assert(t *testing.T, val bool) {
	if !val {
		t.Fail()
	}
}

func makeBackgroundRequest(url string) (*http.Request, error) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	return req, err
}
