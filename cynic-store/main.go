/*
Use this to do simple dumps of cynic-storage files.

Copyright 2018 Simon Symeonidis (psyomn)
Copyright 2019 Simon Symeonidis (psyomn)

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
package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/psyomn/cynic"
)

var cmdInFile = ""

func init() {
	flag.StringVar(&cmdInFile, "input", cmdInFile, "the cynic db store to dump")
}

func usage() {
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if cmdInFile == "" {
		usage()
	}

	var buff bytes.Buffer

	dat, err := ioutil.ReadFile(cmdInFile)
	if err != nil {
		log.Fatal("problem opening file: ", cmdInFile)
		os.Exit(1)
	}

	dec := gob.NewDecoder(&buff)
	var snapstore cynic.SnapshotStore
	buff.Write(dat)

	err = dec.Decode(&snapstore)
	if err != nil {
		log.Println("problem decoding store: ", cmdInFile, ", ", err)
		os.Exit(1)
	}

	fmt.Print(snapstore.String())
}
