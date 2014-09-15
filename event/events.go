// 事件处理.
package event

import "github.com/golang/glog"

const (
	EvtCodeGOPATHNotFound  = iota // 事件代码：找不到环境变量 $GOPATH
	EvtCodeGOROOTNotFound         // 事件代码：找不到环境变量 $GOROOT
	EvtCodeGocodeNotFound         // 事件代码：找不到 gocode
	EvtCodeIDEStubNotFound        // 事件代码：找不到 IDE stub
)

// 事件队列最大长度.
const MaxQueueLength = 10

// 事件结构.
type Event struct {
	Code int    `json:"code"` // 事件代码
	Sid  string `json:"sid"`  // 用户会话 id
}

// 全局事件队列.
// 入队的事件将分发到每个用户的通知队列.
var EventQueue = make(chan int, MaxQueueLength)

// 用户事件队列.
// <sid, chan>
var UserEventQueues = map[string]chan int{}

// 用户事件处理器集.
// <sid, *Handlers>
var UserEventHandlers = map[string]*Handlers{}

// 加载事件处理.
func Load() {
	go func() {
		for event := range EventQueue {
			glog.V(5).Info("收到全局事件 [%d]", event)

			// 将事件分发到每个用户的事件队列里
			for _, userQueue := range UserEventQueues {
				userQueue <- event
			}
		}
	}()
}

// 初始化一个用户事件队列.
func InitUserQueue(sid string, handlers ...Handler) {
	// FIXME: 会话过期后需要销毁对应的用户事件队列

	q := UserEventQueues[sid]
	if nil != q {
		return
	}

	q = make(chan int, MaxQueueLength)
	UserEventQueues[sid] = q

	if nil == UserEventHandlers[sid] {
		UserEventHandlers[sid] = new(Handlers)
	}

	for _, handler := range handlers {
		UserEventHandlers[sid].add(handler)
	}

	go func() {
		for evtCode := range q {
			glog.V(5).Infof("Session [%s] received a event [%d]", sid, evtCode)

			// 将事件交给事件处理器进行处理
			for _, handler := range *UserEventHandlers[sid] {
				e := Event{Code: evtCode, Sid: sid}
				handler.Handle(&e)
			}
		}
	}()
}

// 事件处理接口.
type Handler interface {
	Handle(event *Event)
}

// 函数指针包装.
type HandleFunc func(event *Event)

// 事件处理默认实现.
func (fn HandleFunc) Handle(event *Event) {
	fn(event)
}

type Handlers []Handler

func (handlers *Handlers) add(handler Handler) {
	*handlers = append(*handlers, handler)
}
