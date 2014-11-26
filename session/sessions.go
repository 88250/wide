// Copyright (c) 2014, B3log
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

// Session manipulations.
//
// Wide server side needs maintain two kinds of sessions:
//
//  1. HTTP session: mainly used for login authentication
//  2. Wide session: browser tab open/refresh will create one, and associates with HTTP session
//
// When a session gone: release all resources associated with it, such as running processes, event queues.
package session

import (
	"bytes"
	"encoding/json"
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
	SessionStateActive = iota // session state: active
	SessionStateClosed        // session state: closed (not used so far)
)

var (
	// Session channels. <sid, *util.WSChannel>
	SessionWS = map[string]*util.WSChannel{}

	// Editor channels. <sid, *util.WSChannel>
	EditorWS = map[string]*util.WSChannel{}

	// Output channels. <sid, *util.WSChannel>
	OutputWS = map[string]*util.WSChannel{}

	// Notification channels. <sid, *util.WSChannel>
	NotificationWS = map[string]*util.WSChannel{}
)

// HTTP session store.
var HTTPSession = sessions.NewCookieStore([]byte("BEYOND"))

// Wide session, associated with a browser tab.
type WideSession struct {
	Id          string                     // id
	Username    string                     // username
	HTTPSession *sessions.Session          // HTTP session related
	Processes   []*os.Process              // process set
	EventQueue  *event.UserEventQueue      // event queue
	State       int                        // state
	Content     *conf.LatestSessionContent // the latest session content
	Created     time.Time                  // create time
	Updated     time.Time                  // the latest use time
}

// Type of wide sessions.
type Sessions []*WideSession

// Wide sessions.
var WideSessions Sessions

// Exclusive lock.
var mutex sync.Mutex

// In some special cases (such as a browser uninterrupted refresh / refresh in the source code view) will occur
// some invalid sessions, the function checks and removes these invalid sessions periodically (1 hour).
//
// Invalid sessions: sessions that not used within 30 minutes, refers to WideSession.Updated field.
func FixedTimeRelease() {
	go func() {
		for _ = range time.Tick(time.Hour) {
			hour, _ := time.ParseDuration("-30m")
			threshold := time.Now().Add(hour)

			for _, s := range WideSessions {
				if s.Updated.Before(threshold) {
					glog.V(3).Infof("Removes a invalid session [%s], user [%s]", s.Id, s.Username)

					WideSessions.Remove(s.Id)
				}
			}
		}
	}()
}

// Online user statistic report.
type userReport struct {
	username   string
	sessionCnt int
	updated    time.Time
}

// report returns a online user statistics in pretty format.
func (u *userReport) report() string {
	return "[" + u.username + "] has [" + strconv.Itoa(u.sessionCnt) + "] sessions, latest activity [" +
		u.updated.Format("2006-01-02 15:04:05") + "]"
}

// FixedTimeReport reports the Wide sessions status periodically (10 minutes).
func FixedTimeReport() {
	go func() {
		for _ = range time.Tick(10 * time.Minute) {
			users := map[string]*userReport{} // <username, *userReport>

			for _, s := range WideSessions {
				if report, exists := users[s.Username]; exists {
					if s.Updated.After(report.updated) {
						users[s.Username].updated = s.Updated
					}

					report.sessionCnt++
				} else {
					users[s.Username] = &userReport{username: s.Username, sessionCnt: 1, updated: s.Updated}
				}
			}

			var buf bytes.Buffer
			buf.WriteString("\n  [" + strconv.Itoa(len(users)) + "] users are online and [" + strconv.Itoa(len(WideSessions)) +
				"] sessions currently\n")

			for _, t := range users {
				buf.WriteString("    " + t.report() + "\n")
			}

			glog.Info(buf.String())
		}
	}()
}

// WSHandler handles request of creating session channel.
//
// When a channel closed, releases all resources associated with it.
func WSHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query()["sid"][0]
	wSession := WideSessions.Get(sid)
	if nil == wSession {
		httpSession, _ := HTTPSession.Get(r, "wide-session")

		if httpSession.IsNew {
			return
		}

		httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
		httpSession.Save(r, w)

		wSession = WideSessions.New(httpSession, sid)

		glog.Infof("Created a wide session [%s] for websocket reconnecting, user [%s]", sid, wSession.Username)
	}

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	wsChan := util.WSChannel{Sid: sid, Conn: conn, Request: r, Time: time.Now()}

	ret := map[string]interface{}{"output": "Session initialized", "cmd": "init-session"}
	err := wsChan.WriteJSON(&ret)
	if nil != err {
		return
	}

	SessionWS[sid] = &wsChan

	glog.V(4).Infof("Open a new [Session Channel] with session [%s], %d", sid, len(SessionWS))

	input := map[string]interface{}{}

	for {
		if err := wsChan.ReadJSON(&input); err != nil {
			glog.V(5).Infof("[Session Channel] of session [%s] disconnected, releases all resources with it, user [%s]",
				sid, wSession.Username)

			WideSessions.Remove(sid)

			return
		}

		ret = map[string]interface{}{"output": "", "cmd": "session-output"}

		if err := wsChan.WriteJSON(&ret); err != nil {
			glog.Error("Session WS ERROR: " + err.Error())

			return
		}

		wsChan.Time = time.Now()
	}
}

// SaveContent handles request of session content storing.
func SaveContent(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	args := struct {
		Sid string
		*conf.LatestSessionContent
	}{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	wSession := WideSessions.Get(args.Sid)
	if nil == wSession {
		data["succ"] = false

		return
	}

	wSession.Content = args.LatestSessionContent

	for _, user := range conf.Wide.Users {
		if user.Name == wSession.Username {
			// update the variable in-memory, conf.FixedTimeSave() function will persist it periodically
			user.LatestSessionContent = wSession.Content

			wSession.Refresh()

			return
		}
	}
}

// SetProcesses binds process set with the wide session.
func (s *WideSession) SetProcesses(ps []*os.Process) {
	s.Processes = ps

	s.Refresh()
}

// Refresh refreshes the channel by updating its use time.
func (s *WideSession) Refresh() {
	s.Updated = time.Now()
}

// New creates a wide session.
func (sessions *Sessions) New(httpSession *sessions.Session, sid string) *WideSession {
	mutex.Lock()
	defer mutex.Unlock()

	now := time.Now()

	// create user event queuselect
	userEventQueue := event.UserEventQueues.New(sid)

	ret := &WideSession{
		Id:          sid,
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

// Get gets a wide session with the specified session id.
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

// Remove removes a wide session specified with the given session id, releases resources associated with it.
//
// Session-related resources:
//
//  1. user event queue
//  2. process set
//  3. websocket channels
func (sessions *Sessions) Remove(sid string) {
	mutex.Lock()
	defer mutex.Unlock()

	for i, s := range *sessions {
		if s.Id == sid {
			// remove from session set
			*sessions = append((*sessions)[:i], (*sessions)[i+1:]...)

			// close user event queue
			event.UserEventQueues.Close(sid)

			// kill processes
			for _, p := range s.Processes {
				if err := p.Kill(); nil != err {
					glog.Errorf("Can't kill process [%d] of session [%s], user [%s]", p.Pid, sid, s.Username)
				} else {
					glog.V(3).Infof("Killed a process [%d] of session [%s], user [%s]", p.Pid, sid, s.Username)
				}
			}

			// close websocket channels
			if ws, ok := OutputWS[sid]; ok {
				ws.Close()
				delete(OutputWS, sid)
			}

			if ws, ok := NotificationWS[sid]; ok {
				ws.Close()
				delete(NotificationWS, sid)
			}

			if ws, ok := SessionWS[sid]; ok {
				ws.Close()
				delete(SessionWS, sid)
			}

			cnt := 0 // count wide sessions associated with HTTP session
			for _, s := range *sessions {
				if s.HTTPSession.ID == s.HTTPSession.ID {
					cnt++
				}
			}

			glog.V(5).Infof("Removed a session [%s] of user [%s], it has [%d] sessions currently", sid, s.Username, cnt)

			return
		}
	}
}

// GetByUsername gets wide sessions.
func (sessions *Sessions) GetByUsername(username string) []*WideSession {
	mutex.Lock()
	defer mutex.Unlock()

	ret := []*WideSession{}

	for _, s := range *sessions {
		if s.Username == username {
			ret = append(ret, s)
		}
	}

	return ret
}
