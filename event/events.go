// Copyright (c) 2014-2016, b3log.org & hacpai.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package event includes event related manipulations.
package event

import (
	"os"

	"github.com/b3log/wide/log"
	"github.com/b3log/wide/util"
)

const (
	// EvtCodeGOPATHNotFound indicates an event: not found $GOPATH env variable
	EvtCodeGOPATHNotFound = iota
	// EvtCodeGOROOTNotFound indicates an event: not found $GOROOT env variable
	EvtCodeGOROOTNotFound
	// EvtCodeGocodeNotFound indicates an event: not found gocode
	EvtCodeGocodeNotFound
	// EvtCodeIDEStubNotFound indicates an event: not found gotools
	EvtCodeIDEStubNotFound
	// EvtCodeServerInternalError indicates an event: server internal error
	EvtCodeServerInternalError
)

// Max length of queue.
const maxQueueLength = 10

// Logger.
var logger = log.NewLogger(os.Stdout)

// Event represents an event.
type Event struct {
	Code int         `json:"code"` // event code
	Sid  string      `json:"sid"`  // wide session id related
	Data interface{} `json:"data"` // event data
}

// Global event queue.
//
// Every event in this queue will be dispatched to each user event queue.
var EventQueue = make(chan *Event, maxQueueLength)

// UserEventQueue represents a user event queue.
type UserEventQueue struct {
	Sid      string      // wide session id related
	Queue    chan *Event // queue
	Handlers []Handler   // event handlers
}

type queues map[string]*UserEventQueue

// User event queues.
//
// <sid, *UserEventQueue>
var UserEventQueues = queues{}

// Load initializes the event handling.
func Load() {
	go func() {
		defer util.Recover()

		for event := range EventQueue {
			logger.Debugf("Received a global event [code=%d]", event.Code)

			// dispatch the event to each user event queue
			for _, userQueue := range UserEventQueues {
				event.Sid = userQueue.Sid

				userQueue.Queue <- event
			}
		}
	}()
}

// AddHandler adds the specified handlers to user event queues.
func (uq *UserEventQueue) AddHandler(handlers ...Handler) {
	uq.Handlers = append(uq.Handlers, handlers...)
}

// New initializes a user event queue with the specified wide session id.
func (ueqs queues) New(sid string) *UserEventQueue {

	if q, ok := ueqs[sid]; ok {
		logger.Warnf("Already exist a user queue in session [%s]", sid)

		return q
	}

	q := &UserEventQueue{
		Sid:   sid,
		Queue: make(chan *Event, maxQueueLength),
	}

	ueqs[sid] = q

	go func() { // start listening
		defer util.Recover()

		for evt := range q.Queue {
			logger.Debugf("Session [%s] received an event [%d]", sid, evt.Code)

			// process event by each handlers
			for _, handler := range q.Handlers {
				handler.Handle(evt)
			}
		}
	}()

	return q
}

// Close closes a user event queue with the specified wide session id.
func (ueqs queues) Close(sid string) {

	if q, ok := ueqs[sid]; ok {
		close(q.Queue)
		delete(ueqs, sid)
	}
}

// Handler represents an event handler.
type Handler interface {
	Handle(event *Event)
}

// HandleFunc represents a handler function.
type HandleFunc func(event *Event)

// Default implementation of event handling.
func (fn HandleFunc) Handle(event *Event) {
	fn(event)
}
