// Copyright (c) 2014-present, b3log.org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package output

import (
	"bufio"
	"encoding/json"
	"math/rand"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
)

// RunHandler handles request of executing a binary file.
func RunHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false
	}

	sid := args["sid"].(string)
	wSession := session.WideSessions.Get(sid)
	if nil == wSession {
		result.Succ = false
	}

	filePath := args["executable"].(string)

	var cmd *exec.Cmd
	if conf.Docker {
		fileName := filepath.Base(filePath)
		cmd = exec.Command("docker", "run", "--rm", "--cpus", "0.1", "-v", filePath+":/"+fileName, conf.DockerImageGo, "/"+fileName)
	} else {
		cmd = exec.Command(filePath)
		curDir := filepath.Dir(filePath)
		cmd.Dir = curDir
	}

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		logger.Error(err)
		result.Succ = false
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		logger.Error(err)
		result.Succ = false
	}

	outReader := bufio.NewReader(stdout)
	errReader := bufio.NewReader(stderr)

	if err := cmd.Start(); nil != err {
		logger.Error(err)
		result.Succ = false
	}
	wsChannel := session.OutputWS[sid]
	channelRet := map[string]interface{}{}
	if !result.Succ {
		channelRet["cmd"] = "run-done"
		channelRet["output"] = ""
		wsChannel.WriteJSON(&channelRet)
		wsChannel.Refresh()

		return
	}

	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	channelRet["pid"] = cmd.Process.Pid

	// add the process to user's process set
	Processes.Add(wSession, cmd.Process)

	// push once for front-end to get the 'run' state and pid
	if nil != wsChannel {
		channelRet["cmd"] = "run"
		channelRet["output"] = ""
		if nil != wsChannel {
			wsChannel.WriteJSON(&channelRet)
			wsChannel.Refresh()
		}
	}

	rid := rand.Int()
	go func(runningId int) {
		defer util.Recover()

		logger.Debugf("User [%s, %s] is running [id=%d, file=%s]", wSession.UserId, sid, runningId, filePath)

		go func() {
			defer util.Recover()

			for {
				r, _, err := outReader.ReadRune()
				if nil != err {
					break
				}

				oneRuneStr := string(r)
				oneRuneStr = strings.Replace(oneRuneStr, "<", "&lt;", -1)
				oneRuneStr = strings.Replace(oneRuneStr, ">", "&gt;", -1)
				channelRet["cmd"] = "run"
				channelRet["output"] = oneRuneStr
				wsChannel := session.OutputWS[sid]
				if nil != wsChannel {
					wsChannel.WriteJSON(&channelRet)
					wsChannel.Refresh()
				}
			}
		}()

		for {
			r, _, err := errReader.ReadRune()
			if nil != err {
				break
			}

			oneRuneStr := string(r)
			oneRuneStr = strings.Replace(oneRuneStr, "<", "&lt;", -1)
			oneRuneStr = strings.Replace(oneRuneStr, ">", "&gt;", -1)
			channelRet["cmd"] = "run"
			channelRet["output"] = "<span class='stderr'>" + oneRuneStr + "</span>"
			wsChannel := session.OutputWS[sid]
			if nil != wsChannel {
				wsChannel.WriteJSON(&channelRet)
				wsChannel.Refresh()
			}
		}
	}(rid)

	after := time.After(5 * time.Second)
	channelRet["cmd"] = "run-done"
	select {
	case <-after:
		cmd.Process.Kill()

		channelRet["output"] = "<span class='stderr'>run program timeout in 5s</span>\n"
	case <-done:
		channelRet["output"] = "\n<span class='stderr'>run program complete</span>\n"
	}

	Processes.Remove(wSession, cmd.Process)
	logger.Debugf("User [%s, %s] done running [id=%d, file=%s]", wSession.UserId, sid, rid, filePath)

	if nil != wsChannel {
		wsChannel.WriteJSON(&channelRet)
		wsChannel.Refresh()
	}
}

// StopHandler handles request of stoping a running process.
func StopHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	sid := args["sid"].(string)
	pid := int(args["pid"].(float64))

	wSession := session.WideSessions.Get(sid)
	if nil == wSession {
		result.Succ = false

		return
	}

	Processes.Kill(wSession, pid)
}
