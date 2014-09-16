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
	"github.com/b3log/wide/user"
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

	model := map[string]interface{}{"Wide": conf.Wide, "i18n": i18n.GetAll(r), "locale": i18n.GetLocale(r)}

	session, _ := user.Session.Get(r, "wide-session")

	if session.IsNew {
		// TODO: 写死以 admin 作为用户登录
		name := conf.Wide.Users[0].Name

		session.Values["username"] = name
		session.Values["id"] = strconv.Itoa(rand.Int())
		// 一天过期
		session.Options.MaxAge = 60 * 60 * 24

		glog.Infof("Created a session [%s] for user [%s]", session.Values["id"].(string), name)
	}

	session.Save(r, w)

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
	session, _ := user.Session.Get(r, "wide-session")
	username := session.Values["username"].(string)
	sid := session.Values["id"].(string)

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	wsChan := util.WSChannel{Sid: sid, Conn: conn, Request: r, Time: time.Now()}

	shellWS[sid] = &wsChan

	ret := map[string]interface{}{"output": "Shell initialized", "cmd": "init-shell"}
	wsChan.Conn.WriteJSON(&ret)

	glog.Infof("Open a new [Shell] with session [%s], %d", sid, len(shellWS))

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
