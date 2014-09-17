// Shell.
package shell

import (
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

// Shell 通道.
// <sid, *util.WSChannel>>
var shellWS = map[string]*util.WSChannel{}

// Shell 首页.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	i18n.Load()

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")

	if httpSession.IsNew {
		// TODO: 写死以 admin 作为用户登录
		name := conf.Wide.Users[0].Name

		httpSession.Values["username"] = name
		httpSession.Values["id"] = strconv.Itoa(rand.Int())
		// 一天过期
		httpSession.Options.MaxAge = 60 * 60 * 24

		glog.Infof("Created a HTTP session [%s] for user [%s]", httpSession.Values["id"].(string), name)
	}

	httpSession.Save(r, w)

	// 创建一个 Wide 会话
	wideSession := session.WideSessions.New(httpSession)

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(r), "locale": i18n.GetLocale(r),
		"session": wideSession}

	t, err := template.ParseFiles("view/shell.html")

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, model)
}

// 建立 Shell 通道.
func WSHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	username := httpSession.Values["username"].(string)

	// TODO: 会话校验
	sid := r.URL.Query()["sid"][0]

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	wsChan := util.WSChannel{Sid: sid, Conn: conn, Request: r, Time: time.Now()}

	shellWS[sid] = &wsChan

	ret := map[string]interface{}{"output": "Shell initialized", "cmd": "init-shell"}
	wsChan.Conn.WriteJSON(&ret)

	glog.V(4).Infof("Open a new [Shell] with session [%s], %d", sid, len(shellWS))

	input := map[string]interface{}{}

	for {
		if err := wsChan.Conn.ReadJSON(&input); err != nil {
			if err.Error() == "EOF" {
				return
			}

			if err.Error() == "unexpected EOF" {
				return
			}

			glog.Error("Shell WS ERROR: " + err.Error())
			return
		}

		inputCmd := input["cmd"].(string)

		cmds := strings.Split(inputCmd, "|")
		commands := []*exec.Cmd{}
		for _, cmdWithArgs := range cmds {
			cmdWithArgs = strings.TrimSpace(cmdWithArgs)
			cmdWithArgs := strings.Split(cmdWithArgs, " ")
			args := []string{}
			if len(cmdWithArgs) > 1 {
				args = cmdWithArgs[1:]
			}

			cmd := exec.Command(cmdWithArgs[0], args...)
			commands = append(commands, cmd)
		}

		output := ""
		if !strings.Contains(inputCmd, "clear") {
			output = pipeCommands(username, commands...)
		}

		ret = map[string]interface{}{"output": output, "cmd": "shell-output"}

		if err := wsChan.Conn.WriteJSON(&ret); err != nil {
			glog.Error("Shell WS ERROR: " + err.Error())
			return
		}

		// 更新通道最近使用时间
		wsChan.Time = time.Now()
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
