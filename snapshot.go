/*
Package cynic monitors you from the ceiling.

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
package cynic

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	storeMagic   = 0x43594E4943535452
	storeVersion = 1
)

// SnapshotConfig is the configuration for the snapshots to be taken
type SnapshotConfig struct {
	Enabled   bool
	Interval  time.Duration
	DumpEvery time.Duration
	Path      string
}

// Snapshot is a copy of the state of the map currently being
// monitored.
type snapshot struct {
	Timestamp int64  // unix timestamp
	Data      string // json
}

// SnapshotStore is storage of states of the map at different times
type SnapshotStore struct {
	Magic     uint64
	Version   uint8 // storage version
	Snapshots []*snapshot
}

var snapshotMutex sync.Mutex

// String stringifies the contents of SnapshotStore
func (s *SnapshotStore) String() string {
	snapshotMutex.Lock()
	defer snapshotMutex.Unlock()

	var build strings.Builder

	versionDescription := fmt.Sprintf("version: %d\n", s.Version)
	build.WriteString(versionDescription)

	for _, snap := range s.Snapshots {
		row := fmt.Sprintf("%d:%s\n", snap.Timestamp, snap.Data)
		build.WriteString(row)
	}

	return build.String()
}

func snapshotStoreNew() SnapshotStore {
	snps := make([]*snapshot, 0)
	return SnapshotStore{
		Magic:     storeMagic,
		Version:   storeVersion,
		Snapshots: snps,
	}
}

func (s *SnapshotStore) add(snapshot *snapshot) {
	snapshotMutex.Lock()
	defer snapshotMutex.Unlock()

	s.Snapshots = append(s.Snapshots, snapshot)
}

func (s *SnapshotStore) encode() (bytes.Buffer, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	err := enc.Encode(*s)
	if err != nil {
		log.Println("problem encoding cynic store file: ", err)
	}

	return buffer, err
}

func (s *SnapshotStore) encodeToFile(path string) error {
	snapshotMutex.Lock()
	defer snapshotMutex.Unlock()

	buffer, err := s.encode()
	if err != nil {
		log.Println(err)
		return err
	}

	return ioutil.WriteFile(path, buffer.Bytes(), 0644)
}

func (s *SnapshotStore) clear() {
	snapshotMutex.Lock()
	defer snapshotMutex.Unlock()

	snp := make([]*snapshot, 0)
	s.Snapshots = snp
}
