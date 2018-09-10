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
	"sync"
	"time"
)

// AddressBook contains all the required services inside a map.
type AddressBook struct {
	entries      map[string]*Service
	statusServer StatusServer
	mutex        *sync.Mutex

	alerter       AlertFunc
	alertTicker   *time.Ticker
	alertMessages []AlertMessage
}

// AddressBookNew creates a new address book
func AddressBookNew(session Session) *AddressBook {
	entries := make(map[string]*Service)
	statusServer := StatusServerNew(session.StatusPort, DefaultStatusEndpoint)

	var alertTicker *time.Ticker
	if session.Alerter != nil {
		alertTicker = time.NewTicker(time.Duration(session.AlertTime) * time.Second)
	}

	alertMessages := make([]AlertMessage, 0)

	addressBook := AddressBook{
		entries:       entries,
		statusServer:  statusServer,
		mutex:         &sync.Mutex{},
		alerter:       session.Alerter,
		alertTicker:   alertTicker,
		alertMessages: alertMessages,
	}

	for i := 0; i < len(session.Services); i++ {
		addressBook.AddService(&session.Services[i])
	}

	addressBook.alerter = session.Alerter

	return &addressBook
}

// AddService adds a service
func (s *AddressBook) AddService(service *Service) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	rawurl := service.URL.String()
	if entry, ok := s.entries[rawurl]; ok {
		if entry.running {
			entry.Stop()
		}
	}
	s.entries[rawurl] = &*service
}

// Get gets a reference to the service with the given rawurl.
func (s *AddressBook) Get(rawurl string) (*Service, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	val, found := s.entries[rawurl]
	return val, found
}

// NumEntries returns the number of entries in the AddressBook
func (s *AddressBook) NumEntries() int {
	return len(s.entries)
}

// Contains checks to see if a url is contained in the service list
func (s *AddressBook) Contains(rawurl string) bool {
	_, ok := s.entries[rawurl]
	return ok
}

// Run will run the address book against given services
func (s *AddressBook) Run(signal chan int) {
	log.Println("starting the query service")
	s.StartTickers()

	go func() { s.statusServer.Start() }()

	if s.alerter != nil {
		go func() { s.startAlerter() }()
	}

commands:
	for {
		code := <-signal

		// TODO going to leave some of the signals here for now, but will
		// probably remove them in the future
		switch code {
		case StopService:
			log.Print("received stop signal")
			break commands
		case AddService:
			log.Printf("adding service")
			// Go over anything that has not been started already
			s.StartTickers()
		case DeleteService:
			log.Printf("removing service")
			// Remove from synced map since we only insert things
			// in the sync map (and deletes would not be updated)
		default:
			log.Printf("signal not supported: %d", code)
		}
	}

	s.stopTickers()
	s.statusServer.Stop()
}

// DeleteService removes a service completely from an address book. It
// is OK to pass non-existant rawurls to delete.
func (s *AddressBook) DeleteService(rawurl string) {
	s.mutex.Lock()
	if service, ok := s.entries[rawurl]; ok {
		service.Stop()
		delete(s.entries, rawurl)
	} else {
		log.Print("no such entry to delete: ", rawurl)
	}
	s.mutex.Unlock()

	s.statusServer.Delete(rawurl)
}

// StartTickers starts the tickers on the associated services. This
// might go away in the future
func (s *AddressBook) StartTickers() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, service := range s.entries {
		if service.running {
			continue
		}

		service.running = true

		go func(service *Service, status *StatusServer) {
			if service.offset > 0 {
				waitSeconds := time.Duration(service.offset) * time.Second
				time.Sleep(waitSeconds)
			}

			if service.immediate {
				// Force first tick if service is immediate
				workerQuery(s, service, status)
			}

			for {
				select {
				case <-service.ticker.C:
					workerQuery(s, service, status)
				case <-service.tickerChan:
					return
				}
			}
		}(service, &s.statusServer)
	}
}

func (s *AddressBook) stopTickers() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, service := range s.entries {
		service.ticker.Stop()
	}
}

func (s *AddressBook) queueAlert(message *AlertMessage) {
	if message == nil {
		log.Fatal("don't queue null message alerts")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.alertMessages = append(s.alertMessages, *message)
}

func (s *AddressBook) startAlerter() {
	for range s.alertTicker.C {
		if len(s.alertMessages) > 0 {
			s.mutex.Lock()
			var messages []AlertMessage
			messages = s.alertMessages
			s.alertMessages = make([]AlertMessage, 0)
			s.mutex.Unlock()

			s.alerter(messages)
		}
	}
}
