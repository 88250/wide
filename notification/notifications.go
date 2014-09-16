// 通知.
package notification

import (
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"strconv"
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/event"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/user"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

const (
	Error = "ERROR" // 通知.严重程度：ERROR
	Warn  = "WARN"  // 通知.严重程度：WARN
	Info  = "INFO"  // 通知.严重程度：INFO

	Setup = "Setup" // 通知.类型：安装
)

// 通知结构.
type Notification struct {
	event    *event.Event
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// 一个用户会话的 WebSocket 通道结构.
type WSChannel struct {
	Conn    *websocket.Conn // WebSocket 连接
	Request *http.Request   // 关联的 HTTP 请求
	Time    time.Time       // 该通道最近一次使用时间
}

// 通知通道.
// <sid, WSChannel>
var notificationWSs = map[string]*WSChannel{}

// 用户事件处理：将事件转为通知，并通过通知通道推送给前端.
// 当用户事件队列接收到事件时将会调用该函数进行处理.
func event2Notification(e *event.Event) {
	if nil == notificationWSs[e.Sid] {
		return
	}

	wsChannel := notificationWSs[e.Sid]

	var notification Notification

	switch e.Code {
	case event.EvtCodeGocodeNotFound:
		notification = Notification{event: e, Type: Setup, Severity: Error}
	case event.EvtCodeIDEStubNotFound:
		notification = Notification{event: e, Type: Setup, Severity: Error}
	default:
		glog.Warningf("Can't handle event[code=%d]", e.Code)
		return
	}

	// 消息国际化处理
	notification.Message = i18n.Get(wsChannel.Request, "notification_"+strconv.Itoa(e.Code)).(string)

	wsChannel.Conn.WriteJSON(&notification)
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := user.Session.Get(r, "wide-session")
	sid := session.Values["id"].(string)

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	wsChan := WSChannel{Conn: conn, Request: r, Time: time.Now()}

	notificationWSs[sid] = &wsChan

	ret := map[string]interface{}{"output": "Notification initialized", "cmd": "init-notification"}
	wsChan.Conn.WriteJSON(&ret)

	glog.Infof("Open a new [Notification] with session [%s], %d", sid, len(notificationWSs))

	// 初始化用户事件队列
	event.InitUserQueue(sid, event.HandleFunc(event2Notification))

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

		output := ""

		ret = map[string]interface{}{"output": output, "cmd": "notification-output"}

		if err := wsChan.Conn.WriteJSON(&ret); err != nil {
			glog.Error("Notification WS ERROR: " + err.Error())
			return
		}
	}
}

func pipeCommands(username string, commands ...*exec.Cmd) string {
	for i, command := range commands[:len(commands)-1] {
		setCmdEnv(command, username)

		stdout, err := command.StdoutPipe()
		if nil != err {
			return err.Error()
		}

		command.Start()

		commands[i+1].Stdin = stdout
	}

	last := commands[len(commands)-1]
	setCmdEnv(last, username)

	out, err := last.CombinedOutput()

	if err != nil {
		return err.Error()
	}

	return string(out)
}

func setCmdEnv(cmd *exec.Cmd, username string) {
	userWorkspace := conf.Wide.GetUserWorkspace(username)

	cmd.Env = append(cmd.Env,
		"TERM="+os.Getenv("TERM"),
		"GOPATH="+userWorkspace,
		"GOOS="+runtime.GOOS,
		"GOARCH="+runtime.GOARCH,
		"GOROOT="+runtime.GOROOT(),
		"PATH="+os.Getenv("PATH"))

	cmd.Dir = userWorkspace
}
