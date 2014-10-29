// Event manipulations.
package event

import "github.com/golang/glog"

const (
	EvtCodeGOPATHNotFound      = iota // event code: not found $GOPATH env variable
	EvtCodeGOROOTNotFound             // event code: not found $GOROOT env variable
	EvtCodeGocodeNotFound             // event code: not found gocode
	EvtCodeIDEStubNotFound            // event code: not found ide_stub
	EvtCodeServerInternalError        // event code: server internal error
)

// Max length of queue.
const MaxQueueLength = 10

// Event.
type Event struct {
	Code int         `json:"code"` // event code
	Sid  string      `json:"sid"`  // wide session id related
	Data interface{} `json:"data"` // event data
}

// Global event queue.
//
// Every event in this queue will be dispatched to each user event queue.
var EventQueue = make(chan *Event, MaxQueueLength)

// User event queue.
type UserEventQueue struct {
	Sid      string      // wide session id related
	Queue    chan *Event // queue
	Handlers []Handler   // event handlers
}

type Queues map[string]*UserEventQueue

// User event queues.
//
// <sid, *UserEventQueue>
var UserEventQueues = Queues{}

// Load initializes the event handling.
func Load() {
	go func() {
		for event := range EventQueue {
			glog.V(5).Infof("Received a global event [code=%d]", event.Code)

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
	for _, handler := range handlers {
		uq.Handlers = append(uq.Handlers, handler)
	}
}

// New initializes a user event queue with the specified wide session id.
func (ueqs Queues) New(sid string) *UserEventQueue {
	q := ueqs[sid]
	if nil != q {
		glog.Warningf("Already exist a user queue in session [%s]", sid)

		return q
	}

	q = &UserEventQueue{
		Sid:   sid,
		Queue: make(chan *Event, MaxQueueLength),
	}

	ueqs[sid] = q

	go func() { // start listening
		for evt := range q.Queue {
			glog.V(5).Infof("Session [%s] received a event [%d]", sid, evt.Code)

			// process event by each handlers
			for _, handler := range q.Handlers {
				handler.Handle(evt)
			}
		}
	}()

	return q
}

// Close closes a user event queue with the specified wide session id.
func (ueqs Queues) Close(sid string) {
	q := ueqs[sid]
	if nil == q {
		return
	}

	delete(ueqs, sid)
}

// Type of event handler.
type Handler interface {
	Handle(event *Event)
}

// Type of handler function.
type HandleFunc func(event *Event)

// Default implementation of event handling.
func (fn HandleFunc) Handle(event *Event) {
	fn(event)
}
