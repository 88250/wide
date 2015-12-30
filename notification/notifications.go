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

// Package notification includes notification related manipulations.
package notification

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/event"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/log"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"github.com/gorilla/websocket"
)

const (
	error = "ERROR" // notification.severity: ERROR
	warn  = "WARN"  // notification.severity: WARN
	info  = "INFO"  // notification.severity: INFO

	setup  = "Setup"  // notification.type: setup
	server = "Server" // notification.type: server
)

// Logger.
var logger = log.NewLogger(os.Stdout)

// Notification represents a notification.
type Notification struct {
	event    *event.Event
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// event2Notification processes user event by converting the specified event to a notification, and then push it to front
// browser with notification channel.
func event2Notification(e *event.Event) {
	if nil == session.NotificationWS[e.Sid] {
		return
	}

	wsChannel := session.NotificationWS[e.Sid]
	if nil == wsChannel {
		return
	}

	httpSession, _ := session.HTTPSession.Get(wsChannel.Request, "wide-session")
	username := httpSession.Values["username"].(string)
	locale := conf.GetUser(username).Locale

	var notification *Notification

	switch e.Code {
	case event.EvtCodeGocodeNotFound:
		fallthrough
	case event.EvtCodeIDEStubNotFound:
		notification = &Notification{event: e, Type: setup, Severity: error,
			Message: i18n.Get(locale, "notification_"+strconv.Itoa(e.Code)).(string)}
	case event.EvtCodeServerInternalError:
		notification = &Notification{event: e, Type: server, Severity: error,
			Message: i18n.Get(locale, "notification_"+strconv.Itoa(e.Code)).(string) + " [" + e.Data.(string) + "]"}
	default:
		logger.Warnf("Can't handle event[code=%d]", e.Code)

		return
	}

	wsChannel.WriteJSON(notification)

	wsChannel.Refresh()
}

// WSHandler handles request of creating notification channel.
func WSHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query()["sid"][0]

	wSession := session.WideSessions.Get(sid)

	if nil == wSession {
		return
	}

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	wsChan := util.WSChannel{Sid: sid, Conn: conn, Request: r, Time: time.Now()}

	ret := map[string]interface{}{"notification": "Notification initialized", "cmd": "init-notification"}
	err := wsChan.WriteJSON(&ret)
	if nil != err {
		return
	}

	session.NotificationWS[sid] = &wsChan

	logger.Tracef("Open a new [Notification] with session [%s], %d", sid, len(session.NotificationWS))

	// add user event handler
	wSession.EventQueue.AddHandler(event.HandleFunc(event2Notification))

	input := map[string]interface{}{}

	for {
		if err := wsChan.ReadJSON(&input); err != nil {
			return
		}
	}
}
