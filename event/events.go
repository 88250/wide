// 事件处理.
package event

import (
	"github.com/golang/glog"
)

const (
	EvtGOPATHNotFound  = iota // 事件：找不到环境变量 $GOPATH
	EvtGOROOTNotFound         // 事件：找不到环境变量 $GOROOT
	EvtGocodeNotFount         // 事件：找不到 gocode
	EvtIDEStubNotFound        // 事件：找不到 IDE stub
)

// 全局事件队列.
// 入队的事件将分发到每个用户的通知队列.
var EventQueue = make(chan int, 10)

// 用户事件队列.
// 入队的事件将翻译为通知，并通过通知通道推送到前端.
var UserEventQueue map[string]chan int

// 加载事件处理.
func Load() {
	go func() {
		for {
			// 获取事件
			event := <-EventQueue

			glog.V(5).Info("收到全局事件 [%d]", event)

			// 将事件分发到每个用户的事件队列里
			for _, userQueue := range UserEventQueue {
				userQueue <- event
			}
		}
	}()
}
