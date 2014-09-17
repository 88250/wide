// 会话操作.
// Wide 服务器端需要维护两种会话：
// 1. HTTP 会话：主要用于验证登录
// 2. Wide 会话：浏览器 tab 打开/刷新会创建一个，并和 HTTP 会话进行关联
//
// TODO: 当 HTTP 会话失效时，关联的 Wide 会话也会做失效处理：释放所有和该会话相关的资源，例如运行中的程序进程、事件队列等
package session

import (
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/sessions"
)

const (
	SessionStateActive = iota // 会话状态：活的
)

// 用户 HTTP 会话，用于验证登录.
var HTTPSession = sessions.NewCookieStore([]byte("BEYOND"))

// Wide 会话，对应一个浏览器 tab.
type WideSession struct {
	Id          string            // 唯一标识
	HTTPSession *sessions.Session // 关联的 HTTP 会话
	State       int               // 状态
	Created     time.Time         // 创建时间
	Updated     time.Time         // 最近一次使用时间
}

type Sessions []*WideSession

// 所有 Wide 会话集.
var WideSessions Sessions

// 排它锁，防止并发问题.
var mutex sync.Mutex

// 创建一个 Wide 会话.
func (sessions *Sessions) New(httpSession *sessions.Session) *WideSession {
	mutex.Lock()
	defer mutex.Unlock()

	rand.Seed(time.Now().UnixNano())

	id := strconv.Itoa(rand.Int())
	now := time.Now()

	ret := &WideSession{
		Id:          id,
		HTTPSession: httpSession,
		State:       SessionStateActive,
		Created:     now,
		Updated:     now,
	}

	*sessions = append(*sessions, ret)

	return ret
}

// 移除 Wide 会话.
func (sessions *Sessions) Remove(sid string) {
	mutex.Lock()
	defer mutex.Unlock()

	for i, s := range *sessions {
		if s.Id == sid {
			*sessions = append((*sessions)[:i], (*sessions)[i+1:]...)

			glog.V(3).Infof("Has [%d] wide sessions currently", len(*sessions))
		}
	}
}

// 获取 HTTP 会话关联的所有 Wide 会话.
func (sessions *Sessions) GetByHTTPSession(httpSession *sessions.Session) []*WideSession {
	mutex.Lock()
	defer mutex.Unlock()

	ret := []*WideSession{}

	for _, s := range *sessions {
		if s.HTTPSession.ID == httpSession.ID {
			ret = append(ret, s)
		}
	}

	return ret
}

// 移除 HTTP 会话关联的所有 Wide 会话.
func (sessions *Sessions) RemoveByHTTPSession(httpSession *sessions.Session) {
	mutex.Lock()
	defer mutex.Unlock()

	for i, s := range *sessions {
		if s.HTTPSession.ID == httpSession.ID {
			*sessions = append((*sessions)[:i], (*sessions)[i+1:]...)

			glog.V(3).Infof("Has [%d] wide sessions currently", len(*sessions))
		}
	}
}
