// 事件处理.
package event

import "github.com/golang/glog"

const (
	EvtGOPATHNotFound  = iota // 事件：找不到环境变量 $GOPATH
	EvtGOROOTNotFound         // 事件：找不到环境变量 $GOROOT
	EvtGocodeNotFount         // 事件：找不到 gocode
	EvtIDEStubNotFound        // 事件：找不到 IDE stub

)

const MaxQueueLength = 10

// 全局事件队列.
// 入队的事件将分发到每个用户的通知队列.
var EventQueue = make(chan int, MaxQueueLength)

// 用户事件队列.
// 入队的事件将翻译为通知，并通过通知通道推送到前端.
var UserEventQueues = map[string]chan int{}

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

// 添加一个用户事件队列.
func InitUserQueue(sid string) {
	// FIXME: 会话过期后需要销毁对应的用户事件队列

	q := UserEventQueues[sid]
	if nil != q {
		close(q)
	}

	q = make(chan int, MaxQueueLength)
	UserEventQueues[sid] = q

	go func() {
		for event := range q {
			glog.Infof("Session [%s] received a event [%d]", sid, event)
		}
	}()
}
