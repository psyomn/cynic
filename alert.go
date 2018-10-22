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

// AlertFunc defines the hook signature for alert messages
type AlertFunc = func([]AlertMessage)

// AlertMessage defines a simple alert structure that can be used by
// users of the library, and decide how to show information about the
// alerts.
type AlertMessage struct {
	Response      interface{} `json:"response_text"`
	Endpoint      string      `json:"endpoint"`
	Now           string      `json:"now"`
	CynicHostname string      `json:"cynic_hostname"`
}
