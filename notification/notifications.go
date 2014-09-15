// 通知.
package notification

import (
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/event"
	"github.com/b3log/wide/user"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

// 通知结构.
type Notification struct {
	Event    int
	Type     string
	Severity string // ERROR/WARN/INFO
	Message  string
}

// 一个用户会话的 WebSocket 通道结构.
type WSChannel struct {
	Conn *websocket.Conn // WebSocket 连接
	Time time.Time       // 该通道最近一次使用时间
}

// 通知通道.
// <username, {<sid1, WSChannel1>, <sid2, WSChannel2>}>
var notificationWSs = map[string]map[string]WSChannel{}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := user.Session.Get(r, "wide-session")
	username := session.Values["username"].(string)
	sid := session.Values["id"].(string)

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	wsChan := WSChannel{Conn: conn, Time: time.Now()}

	wsChans := notificationWSs[username]
	if nil == wsChans {
		wsChans = map[string]WSChannel{}
	}

	wsChans[sid] = wsChan

	ret := map[string]interface{}{"output": "Notification initialized", "cmd": "init-notification"}
	wsChan.Conn.WriteJSON(&ret)

	glog.Infof("Open a new [Notification] with session [%s], %d", sid, len(wsChans))

	event.InitUserQueue(sid)

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
