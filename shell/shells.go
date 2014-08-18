package shell

import (
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/session"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var shellWS = map[string]*websocket.Conn{}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "wide-session")
	sid := session.Values["id"].(string)

	shellWS[sid], _ = websocket.Upgrade(w, r, nil, 1024, 1024)

	ret := map[string]interface{}{"output": "Shell initialized", "cmd": "init-shell"}
	shellWS[sid].WriteJSON(&ret)

	glog.Infof("Open a new [Shell] with session [%s], %d", sid, len(shellWS))

	input := map[string]interface{}{}

	for {
		if err := shellWS[sid].ReadJSON(&input); err != nil {
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
			output = pipeCommands(commands...)
		}

		ret = map[string]interface{}{"output": output, "cmd": "shell-output"}

		if err := shellWS[sid].WriteJSON(&ret); err != nil {
			glog.Error("Shell WS ERROR: " + err.Error())
			return
		}
	}
}

func pipeCommands(commands ...*exec.Cmd) string {
	for i, command := range commands[:len(commands)-1] {
		setCmdEnv(command)

		out, err := command.StdoutPipe()

		if nil != err {
			return err.Error()
		}

		command.Start()
		commands[i+1].Stdin = out
	}

	last := commands[len(commands)-1]
	setCmdEnv(last)

	out, err := last.Output()
	if err != nil {
		return err.Error()
	}

	return string(out)
}

func setCmdEnv(cmd *exec.Cmd) {
	cmd.Env = append(cmd.Env, "TERM=xterm", "GOPATH="+conf.Wide.GOPATH,
		"GOROOT="+os.Getenv("GOROOT"))

	cmd.Dir = conf.Wide.GOPATH
}
