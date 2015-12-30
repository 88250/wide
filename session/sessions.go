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

// Package session includes session related manipulations.
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
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/event"
	"github.com/b3log/wide/log"
	"github.com/b3log/wide/util"
	"github.com/go-fsnotify/fsnotify"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

const (
	sessionStateActive = iota
	sessionStateClosed // (not used so far)
)

// Logger.
var logger = log.NewLogger(os.Stdout)

var (
	// SessionWS holds all session channels. <sid, *util.WSChannel>
	SessionWS = map[string]*util.WSChannel{}

	// EditorWS holds all editor channels. <sid, *util.WSChannel>
	EditorWS = map[string]*util.WSChannel{}

	// OutputWS holds all output channels. <sid, *util.WSChannel>
	OutputWS = map[string]*util.WSChannel{}

	// NotificationWS holds all notification channels. <sid, *util.WSChannel>
	NotificationWS = map[string]*util.WSChannel{}

	// PlaygroundWS holds all playground channels. <sid, *util.WSChannel>
	PlaygroundWS = map[string]*util.WSChannel{}
)

// HTTP session store.
var HTTPSession = sessions.NewCookieStore([]byte("BEYOND"))

// WideSession represents a session associated with a browser tab.
type WideSession struct {
	ID          string                     // id
	Username    string                     // username
	HTTPSession *sessions.Session          // HTTP session related
	Processes   []*os.Process              // process set
	EventQueue  *event.UserEventQueue      // event queue
	State       int                        // state
	Content     *conf.LatestSessionContent // the latest session content
	FileWatcher *fsnotify.Watcher          // files change watcher
	Created     time.Time                  // create time
	Updated     time.Time                  // the latest use time
}

// Type of wide sessions.
type wSessions []*WideSession

// Wide sessions.
var WideSessions wSessions

// Exclusive lock.
var mutex sync.Mutex

// FixedTimeRelease releases invalid sessions.
//
// In some special cases (such as a browser uninterrupted refresh / refresh in the source code view) will occur
// some invalid sessions, the function checks and removes these invalid sessions periodically (1 hour).
//
// Invalid sessions: sessions that not used within 30 minutes, refers to WideSession.Updated field.
func FixedTimeRelease() {
	go func() {
		defer util.Recover()

		for _ = range time.Tick(time.Hour) {
			hour, _ := time.ParseDuration("-30m")
			threshold := time.Now().Add(hour)

			for _, s := range WideSessions {
				if s.Updated.Before(threshold) {
					logger.Debugf("Removes a invalid session [%s], user [%s]", s.ID, s.Username)

					WideSessions.Remove(s.ID)
				}
			}
		}
	}()
}

// Online user statistic report.
type userReport struct {
	username   string
	sessionCnt int
	processCnt int
	updated    time.Time
}

// report returns a online user statistics in pretty format.
func (u *userReport) report() string {
	return "[" + u.username + "] has [" + strconv.Itoa(u.sessionCnt) + "] sessions and [" + strconv.Itoa(u.processCnt) +
		"] running processes, latest activity [" + u.updated.Format("2006-01-02 15:04:05") + "]"
}

// FixedTimeReport reports the Wide sessions status periodically (10 minutes).
func FixedTimeReport() {
	go func() {
		defer util.Recover()

		for _ = range time.Tick(10 * time.Minute) {
			users := userReports{}
			processSum := 0

			for _, s := range WideSessions {
				processCnt := len(s.Processes)
				processSum += processCnt

				if report, exists := contains(users, s.Username); exists {
					if s.Updated.After(report.updated) {
						report.updated = s.Updated
					}

					report.sessionCnt++
					report.processCnt += processCnt
				} else {
					users = append(users, &userReport{username: s.Username, sessionCnt: 1, processCnt: processCnt, updated: s.Updated})
				}
			}

			var buf bytes.Buffer
			buf.WriteString("\n  [" + strconv.Itoa(len(users)) + "] users, [" + strconv.Itoa(processSum) + "] running processes and [" +
				strconv.Itoa(len(WideSessions)) + "] sessions currently\n")

			sort.Sort(users)

			for _, t := range users {
				buf.WriteString("    " + t.report() + "\n")
			}

			logger.Info(buf.String())
		}
	}()
}

func contains(reports []*userReport, username string) (*userReport, bool) {
	for _, ur := range reports {
		if username == ur.username {
			return ur, true
		}
	}

	return nil, false
}

type userReports []*userReport

func (f userReports) Len() int           { return len(f) }
func (f userReports) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f userReports) Less(i, j int) bool { return f[i].processCnt > f[j].processCnt }

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// WSHandler handles request of creating session channel.
//
// When a channel closed, releases all resources associated with it.
func WSHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query()["sid"][0]

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	wsChan := util.WSChannel{Sid: sid, Conn: conn, Request: r, Time: time.Now()}

	ret := map[string]interface{}{"output": "Session initialized", "cmd": "init-session"}
	err := wsChan.WriteJSON(&ret)
	if nil != err {
		return
	}

	SessionWS[sid] = &wsChan

	wSession := WideSessions.Get(sid)
	if nil == wSession {
		httpSession, _ := HTTPSession.Get(r, "wide-session")

		if httpSession.IsNew {
			return
		}

		httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
		httpSession.Save(r, w)

		wSession = WideSessions.new(httpSession, sid)

		logger.Tracef("Created a wide session [%s] for websocket reconnecting, user [%s]", sid, wSession.Username)
	}

	logger.Tracef("Open a new [Session Channel] with session [%s], %d", sid, len(SessionWS))

	input := map[string]interface{}{}

	wsChan.Conn.SetReadDeadline(time.Now().Add(pongWait))
	wsChan.Conn.SetPongHandler(func(string) error { wsChan.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		WideSessions.Remove(sid)
		ticker.Stop()
		wsChan.Close()
	}()

	// send websocket ping message.
	go func(t *time.Ticker, channel util.WSChannel) {
		for {
			select {
			case <-t.C:
				if err := channel.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					return
				}
			}
		}

	}(ticker, wsChan)

	for {
		if err := wsChan.ReadJSON(&input); err != nil {
			logger.Tracef("[Session Channel] of session [%s] disconnected, releases all resources with it, user [%s]", sid, wSession.Username)

			return
		}

		ret = map[string]interface{}{"output": "", "cmd": "session-output"}

		if err := wsChan.WriteJSON(&ret); err != nil {
			logger.Error("Session WS ERROR: " + err.Error())

			return
		}

		wsChan.Time = time.Now()
	}
}

// SaveContentHandler handles request of session content string.
func SaveContentHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	args := struct {
		Sid string
		*conf.LatestSessionContent
	}{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	wSession := WideSessions.Get(args.Sid)
	if nil == wSession {
		result.Succ = false

		return
	}

	wSession.Content = args.LatestSessionContent

	for _, user := range conf.Users {
		if user.Name == wSession.Username {
			// update the variable in-memory, session.FixedTimeSave() function will persist it periodically
			user.LatestSessionContent = wSession.Content

			user.Lived = time.Now().UnixNano()

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

// GenId generates a wide session id.
func (sessions *wSessions) GenId() string {
	rand.Seed(time.Now().UnixNano())

	return strconv.Itoa(rand.Int())
}

// Get gets a wide session with the specified session id.
func (sessions *wSessions) Get(sid string) *WideSession {
	mutex.Lock()
	defer mutex.Unlock()

	for _, s := range *sessions {
		if s.ID == sid {
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
//  4. file watcher
func (sessions *wSessions) Remove(sid string) {
	mutex.Lock()
	defer mutex.Unlock()

	for i, s := range *sessions {
		if s.ID == sid {
			// remove from session set
			*sessions = append((*sessions)[:i], (*sessions)[i+1:]...)

			// close user event queue
			event.UserEventQueues.Close(sid)

			// kill processes
			for _, p := range s.Processes {
				if err := p.Kill(); nil != err {
					logger.Errorf("Can't kill process [%d] of session [%s], user [%s]", p.Pid, sid, s.Username)
				} else {
					logger.Debugf("Killed a process [%d] of session [%s], user [%s]", p.Pid, sid, s.Username)
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

			if ws, ok := PlaygroundWS[sid]; ok {
				ws.Close()
				delete(PlaygroundWS, sid)
			}

			// file watcher
			if nil != s.FileWatcher {
				s.FileWatcher.Close()
			}

			cnt := 0 // count wide sessions associated with HTTP session
			for _, ses := range *sessions {
				if ses.Username == s.Username {
					cnt++
				}
			}

			logger.Debugf("Removed a session [%s] of user [%s], it has [%d] sessions currently", sid, s.Username, cnt)

			return
		}
	}
}

// GetByUsername gets wide sessions.
func (sessions *wSessions) GetByUsername(username string) []*WideSession {
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

// new creates a wide session.
func (sessions *wSessions) new(httpSession *sessions.Session, sid string) *WideSession {
	mutex.Lock()
	defer mutex.Unlock()

	username := httpSession.Values["username"].(string)
	now := time.Now()

	ret := &WideSession{
		ID:          sid,
		Username:    username,
		HTTPSession: httpSession,
		EventQueue:  nil,
		State:       sessionStateActive,
		Content:     &conf.LatestSessionContent{},
		Created:     now,
		Updated:     now,
	}

	*sessions = append(*sessions, ret)

	if "playground" == username {
		return ret
	}

	// create user event queue
	ret.EventQueue = event.UserEventQueues.New(sid)

	// add a filesystem watcher to notify front-end after the files changed
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error(err)

		return ret
	}

	go func() {
		defer util.Recover()

		for {
			ch := SessionWS[sid]
			if nil == ch {
				return // release this gorutine
			}

			select {
			case event := <-watcher.Events:
				path := filepath.ToSlash(event.Name)
				dir := filepath.ToSlash(filepath.Dir(path))

				ch = SessionWS[sid]
				if nil == ch {
					return // release this gorutine
				}

				logger.Trace(event)

				if event.Op&fsnotify.Create == fsnotify.Create {
					fileType := "f"

					if util.File.IsDir(path) {
						fileType = "d"

						if err = watcher.Add(path); nil != err {
							logger.Warn(err, path)
						}
					}

					cmd := map[string]interface{}{"path": path, "dir": dir,
						"cmd": "create-file", "type": fileType}
					ch.WriteJSON(&cmd)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					cmd := map[string]interface{}{"path": path, "dir": dir,
						"cmd": "remove-file", "type": ""}
					ch.WriteJSON(&cmd)

				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					cmd := map[string]interface{}{"path": path, "dir": dir,
						"cmd": "rename-file", "type": ""}
					ch.WriteJSON(&cmd)
				}
			case err := <-watcher.Errors:
				if nil != err {
					logger.Error("File watcher ERROR: ", err)
				}
			}
		}
	}()

	go func() {
		defer util.Recover()

		workspaces := filepath.SplitList(conf.GetUserWorkspace(username))
		for _, workspace := range workspaces {
			filepath.Walk(filepath.Join(workspace, "src"), func(dirPath string, f os.FileInfo, err error) error {
				if ".git" == f.Name() { // XXX: discard other unconcered dirs
					return filepath.SkipDir
				}

				if f.IsDir() {
					if err = watcher.Add(dirPath); nil != err {
						logger.Error(err, dirPath)
					}

					logger.Tracef("File watcher added a dir [%s]", dirPath)
				}

				return nil
			})
		}

		ret.FileWatcher = watcher
	}()

	return ret
}
