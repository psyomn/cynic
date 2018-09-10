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

import (
	"encoding/json"
	"log"
)

// EndpointJSON is the format that we process when receiving and parse
// json from an enpoint. Basically just an interface to be consumed by
// other things.
type EndpointJSON = interface{}

// JSONPathSpec is a JSONPath string.
type JSONPathSpec = string

// ParseEndpointJSON parses the json returned from an endpoint.
func parseEndpointJSON(raw []byte) EndpointJSON {
	var result interface{}
	error := json.Unmarshal(raw, &result)

	if error != nil {
		log.Println("json decoding failed: ", error)
		return nil
	}

	return result
}
