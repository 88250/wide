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
//
// 入队的事件将分发到每个用户的事件队列中.
var EventQueue = make(chan int, MaxQueueLength)

// 用户事件队列.
type UserEventQueue struct {
	Sid      string    // 关联的会话 id
	Queue    chan int  // 队列
	Handlers []Handler // 事件处理器集
}

// 事件队列集类型.
type Queues map[string]*UserEventQueue

// 用户事件队列集.
//
// <sid, *UserEventQueue>
var UserEventQueues = Queues{}

// 加载事件处理.
func Load() {
	go func() {
		for event := range EventQueue {
			glog.V(5).Info("收到全局事件 [%d]", event)

			// 将事件分发到每个用户的事件队列里
			for _, userQueue := range UserEventQueues {
				userQueue.Queue <- event
			}
		}
	}()
}

// 为用户队列添加事件处理器.
func (uq *UserEventQueue) AddHandler(handlers ...Handler) {
	for _, handler := range handlers {
		uq.Handlers = append(uq.Handlers, handler)
	}
}

// 初始化一个用户事件队列.
func (ueqs Queues) New(sid string) *UserEventQueue {
	q := ueqs[sid]
	if nil != q {
		glog.Warningf("Already exist a user queue in session [%s]", sid)

		return q
	}

	q = &UserEventQueue{
		Sid:   sid,
		Queue: make(chan int, MaxQueueLength),
	}

	ueqs[sid] = q

	go func() { // 队列开始监听事件
		for evtCode := range q.Queue {
			glog.V(5).Infof("Session [%s] received a event [%d]", sid, evtCode)

			// 将事件交给事件处理器进行处理
			for _, handler := range q.Handlers {
				handler.Handle(&Event{Code: evtCode, Sid: sid})

			}
		}
	}()

	return q
}

// 关闭一个用户事件队列.
func (ueqs Queues) Close(sid string) {
	q := ueqs[sid]
	if nil == q {
		return
	}

	delete(ueqs, sid)
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
