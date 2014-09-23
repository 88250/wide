// 会话操作.
// Wide 服务器端需要维护两种会话：
// 1. HTTP 会话：主要用于验证登录
// 2. Wide 会话：浏览器 tab 打开/刷新会创建一个，并和 HTTP 会话进行关联
//
// 当会话失效时：释放所有和该会话相关的资源，例如运行中的程序进程、事件队列等.
package session

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/event"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

const (
	SessionStateActive = iota // 会话状态：活的
	SessionStateClosed        // 会话状态：已关闭（这个状态目前暂时没有使用到）
)

var (
	// 会话通道. <sid, *util.WSChannel>var
	sessionWS = map[string]*util.WSChannel{}

	// 输出通道. <sid, *util.WSChannel>
	OutputWS = map[string]*util.WSChannel{}

	// 通知通道. <sid, *util.WSChannel>
	NotificationWS = map[string]*util.WSChannel{}
)

// 用户 HTTP 会话，用于验证登录.
var HTTPSession = sessions.NewCookieStore([]byte("BEYOND"))

// Wide 会话，对应一个浏览器 tab.
type WideSession struct {
	Id          string                     // 唯一标识
	Username    string                     // 用户名
	HTTPSession *sessions.Session          // 关联的 HTTP 会话
	Processes   []*os.Process              // 关联的进程集
	EventQueue  *event.UserEventQueue      // 关联的事件队列
	State       int                        // 状态
	Content     *conf.LatestSessionContent // 最近一次会话内容
	Created     time.Time                  // 创建时间
	Updated     time.Time                  // 最近一次使用时间
}

type Sessions []*WideSession

// 所有 Wide 会话集.
var WideSessions Sessions

// 排它锁，防止并发修改.
var mutex sync.Mutex

// 建立会话通道.
// 通道断开时销毁会话状态，回收相关资源.
func WSHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query()["sid"][0]
	wSession := WideSessions.Get(sid)
	if nil == wSession {
		glog.Errorf("Session [%s] not found", sid)

		return
	}

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	wsChan := util.WSChannel{Sid: sid, Conn: conn, Request: r, Time: time.Now()}

	sessionWS[sid] = &wsChan

	ret := map[string]interface{}{"output": "Ouput initialized", "cmd": "init-session"}
	wsChan.Conn.WriteJSON(&ret)

	glog.V(4).Infof("Open a new [Session Channel] with session [%s], %d", sid, len(sessionWS))

	input := map[string]interface{}{}

	for {
		if err := wsChan.Conn.ReadJSON(&input); err != nil {
			glog.V(3).Infof("[Session Channel] of session [%s] disconnected, releases all resources with it", sid)

			for i, s := range WideSessions {
				if s.Id == sid {
					mutex.Lock()

					// 从会话集中移除
					WideSessions = append(WideSessions[:i], WideSessions[i+1:]...)

					// 关闭用户事件队列
					event.UserEventQueues.Close(sid)

					// 杀进程
					for _, p := range s.Processes {
						if err := p.Kill(); nil != err {
							glog.Errorf("Can't kill process [%d] of session [%s]", p.Pid, sid)
						} else {
							glog.V(3).Infof("Killed a process [%d] of session [%s]", p.Pid, sid)
						}
					}

					// 回收所有通道
					OutputWS[sid].Close()
					delete(OutputWS, sid)

					NotificationWS[sid].Close()
					delete(NotificationWS, sid)

					sessionWS[sid].Close()
					delete(sessionWS, sid)

					mutex.Unlock()

					return
				}
			}
		}

		ret = map[string]interface{}{"output": "", "cmd": "session-output"}

		if err := wsChan.Conn.WriteJSON(&ret); err != nil {
			glog.Error("Session WS ERROR: " + err.Error())
			return
		}

		wsChan.Time = time.Now()
	}
}

// 会话内容保存.
func SaveContent(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	args := struct {
		sid string
		*conf.LatestSessionContent
	}{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	wSession := WideSessions.Get(args.sid)
	if nil == wSession {
		data["succ"] = false

		return
	}

	wSession.Content = args.LatestSessionContent

	for _, user := range conf.Wide.Users {
		if user.Name == wSession.Username {
			user.LatestSessionContent = wSession.Content

			// 定时任务会负责持久化

			return
		}
	}
}

// 设置会话关联的进程集.
func (s *WideSession) SetProcesses(ps []*os.Process) {
	s.Processes = ps

	s.Refresh()
}

// 刷新会话最近一次使用时间.
func (s *WideSession) Refresh() {
	s.Updated = time.Now()
}

// 创建一个 Wide 会话.
func (sessions *Sessions) New(httpSession *sessions.Session) *WideSession {
	mutex.Lock()
	defer mutex.Unlock()

	rand.Seed(time.Now().UnixNano())

	id := strconv.Itoa(rand.Int())
	now := time.Now()

	// 创建用户事件队列
	userEventQueue := event.UserEventQueues.New(id)

	ret := &WideSession{
		Id:          id,
		Username:    httpSession.Values["username"].(string),
		HTTPSession: httpSession,
		EventQueue:  userEventQueue,
		State:       SessionStateActive,
		Content:     &conf.LatestSessionContent{},
		Created:     now,
		Updated:     now,
	}

	*sessions = append(*sessions, ret)

	return ret
}

// 获取 Wide 会话.
func (sessions *Sessions) Get(sid string) *WideSession {
	mutex.Lock()
	defer mutex.Unlock()

	for _, s := range *sessions {
		if s.Id == sid {
			return s
		}
	}

	return nil
}

// 移除 Wide 会话.
func (sessions *Sessions) Remove(sid string) {
	mutex.Lock()
	defer mutex.Unlock()

	for i, s := range *sessions {
		if s.Id == sid {
			*sessions = append((*sessions)[:i], (*sessions)[i+1:]...)

			glog.V(3).Infof("Removed a session [%s], has [%d] wide sessions currently", sid, len(*sessions))

			return
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
