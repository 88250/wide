// Notification manipulations.
package notification

import (
	"net/http"
	"time"

	"strconv"
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/event"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

const (
	Error = "ERROR" // notification.severity: ERROR
	Warn  = "WARN"  // notification.severity: WARN
	Info  = "INFO"  // notification.severity: INFO

	Setup  = "Setup"  // notification.type: setup
	Server = "Server" // notification.type: server
)

// Notification.
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
	locale := conf.Wide.GetUser(username).Locale

	var notification *Notification

	switch e.Code {
	case event.EvtCodeGocodeNotFound:
		fallthrough
	case event.EvtCodeIDEStubNotFound:
		notification = &Notification{event: e, Type: Setup, Severity: Error,
			Message: i18n.Get(locale, "notification_"+strconv.Itoa(e.Code)).(string)}
	case event.EvtCodeServerInternalError:
		notification = &Notification{event: e, Type: Server, Severity: Error,
			Message: i18n.Get(locale, "notification_"+strconv.Itoa(e.Code)).(string) + " [" + e.Data.(string) + "]"}
	default:
		glog.Warningf("Can't handle event[code=%d]", e.Code)

		return
	}

	wsChannel.Conn.WriteJSON(notification)

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

	session.NotificationWS[sid] = &wsChan

	glog.V(4).Infof("Open a new [Notification] with session [%s], %d", sid, len(session.NotificationWS))

	// add user event handler
	wSession.EventQueue.AddHandler(event.HandleFunc(event2Notification))

	input := map[string]interface{}{}

	for {
		if err := wsChan.Conn.ReadJSON(&input); err != nil {
			if err.Error() == "EOF" {
				return
			}

			if err.Error() == "unexpected EOF" {
				return
			}

			glog.Error("Notification WS ERROR: " + err.Error())

			return
		}
	}
}
