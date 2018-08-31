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
	"log"
	"reflect"
	"runtime"
)

func getFuncName(fn interface{}) (hookname string) {
	hookname = runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	return
}

func nilOrDie(err error, str string) {
	if err != nil {
		log.Fatal(str, ": ", err)
	}
}

func nilAndOk(err error, str string) {
	if err != nil {
		log.Print(str, ": ", err)
	}
}
