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

// Shell channel.
//
// <sid, *util.WSChannel>>
var ShellWS = map[string]*util.WSChannel{}

// IndexHandler handles request of Shell index.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	// create a wide session
	rand.Seed(time.Now().UnixNano())
	sid := strconv.Itoa(rand.Int())
	wideSession := session.WideSessions.New(httpSession, sid)

	username := httpSession.Values["username"].(string)
	locale := conf.Wide.GetUser(username).Locale

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale,
		"session": wideSession}

	wideSessions := session.WideSessions.GetByUsername(username)

	glog.V(3).Infof("User [%s] has [%d] sessions", username, len(wideSessions))

	t, err := template.ParseFiles("views/shell.html")

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, model)
}

// WSHandler handles request of creating Shell channel.
func WSHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)

	sid := r.URL.Query()["sid"][0]

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	wsChan := util.WSChannel{Sid: sid, Conn: conn, Request: r, Time: time.Now()}

	ret := map[string]interface{}{"output": "Shell initialized", "cmd": "init-shell"}
	err := wsChan.WriteJSON(&ret)
	if nil != err {
		return
	}

	ShellWS[sid] = &wsChan

	glog.V(4).Infof("Open a new [Shell] with session [%s], %d", sid, len(ShellWS))

	input := map[string]interface{}{}

	for {
		if err := wsChan.ReadJSON(&input); err != nil {
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

		if err := wsChan.WriteJSON(&ret); err != nil {
			glog.Error("Shell WS ERROR: " + err.Error())
			return
		}

		wsChan.Refresh()
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

	// release resources
	for _, command := range commands[:len(commands)-1] {
		command.Wait()
	}

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
