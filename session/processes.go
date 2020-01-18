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

package session

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/88250/gulu"
	"github.com/88250/wide/conf"
	"github.com/88250/wide/util"
)

// Type of process set.
type procs map[string][]*os.Process

// Processse of all users.
//
// <sid, []*os.Process>
var Processes = procs{}

// Exclusive lock.
var procMutex sync.Mutex

// RunHandler handles request of executing a binary file.
func RunHandler(w http.ResponseWriter, r *http.Request, channel map[string]*util.WSChannel) {
	result := gulu.Ret.NewResult()
	defer gulu.Ret.RetResult(w, r, result)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Code = -1
	}

	sid := args["sid"].(string)
	wSession := WideSessions.Get(sid)
	if nil == wSession {
		result.Code = -1
	}

	filePath := args["executable"].(string)

	randInt := rand.Int()
	rid := strconv.Itoa(randInt)
	var cmd *exec.Cmd
	if conf.Docker {
		fileName := filepath.Base(filePath)
		//cmdStr := "docker run --rm --cpus 0.5 --name " + rid + " -v " + filePath + ":/" + fileName + " " + conf.DockerImageGo + " /" + fileName
		//cmd = exec.Command("sh", "-c", cmdStr)
		cmd = exec.Command("docker", "run", "--rm", "--cpus", "0.5", "--name", rid, "-v", filePath+":/"+fileName, conf.DockerImageGo, "/"+fileName)
	} else {
		cmd = exec.Command(filePath)
		curDir := filepath.Dir(filePath)
		cmd.Dir = curDir
	}

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	if err := cmd.Start(); nil != err {
		logger.Error(err)
		result.Code = -1
	}
	wsChannel := channel[sid]
	channelRet := map[string]interface{}{}
	if 0 != result.Code {
		channelRet["cmd"] = "run-done"
		channelRet["output"] = ""
		wsChannel.WriteJSON(&channelRet)
		wsChannel.Refresh()

		return
	}

	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	channelRet["pid"] = cmd.Process.Pid
	Processes.Add(wSession, cmd.Process)
	shouldExitBuf := false

	// push once for front-end to get the 'run' state and pid
	if nil != wsChannel {
		channelRet["cmd"] = "run"
		channelRet["output"] = ""
		if nil != wsChannel {
			wsChannel.WriteJSON(&channelRet)
			wsChannel.Refresh()
		}
	}

	go func() {
		defer gulu.Panic.Recover(nil)
		logger.Debugf("User [%s, %s] is running [id=%s, file=%s]", wSession.UserId, sid, rid, filePath)

		go func() {
			defer gulu.Panic.Recover(nil)
			for {
				if shouldExitBuf {
					break
				}

				if 1 > outBuf.Len() {
					time.Sleep(7 * time.Millisecond)
					logger.Info("out buf is empty")
					continue
				}

				r, _, err := outBuf.ReadRune()
				if nil != err {
					time.Sleep(7 * time.Millisecond)
					continue
				}

				oneRuneStr := string(r)
				oneRuneStr = strings.Replace(oneRuneStr, "<", "&lt;", -1)
				oneRuneStr = strings.Replace(oneRuneStr, ">", "&gt;", -1)
				channelRet["cmd"] = "run"
				channelRet["output"] = oneRuneStr
				wsChannel := channel[sid]
				if nil != wsChannel {
					wsChannel.WriteJSON(&channelRet)
					wsChannel.Refresh()
				}
			}
		}()

		for {
			if shouldExitBuf {
				break
			}

			if 1 > errBuf.Len() {
				time.Sleep(7 * time.Millisecond)
				continue
			}

			r, _, err := errBuf.ReadRune()
			if nil != err {
				time.Sleep(7 * time.Millisecond)
				continue
			}

			oneRuneStr := string(r)
			oneRuneStr = strings.Replace(oneRuneStr, "<", "&lt;", -1)
			oneRuneStr = strings.Replace(oneRuneStr, ">", "&gt;", -1)
			channelRet["cmd"] = "run"
			channelRet["output"] = "<span class='stderr'>" + oneRuneStr + "</span>"
			wsChannel := channel[sid]
			if nil != wsChannel {
				wsChannel.WriteJSON(&channelRet)
				wsChannel.Refresh()
			}
		}
	}()

	after := time.After(5 * time.Second)
	kill := false
	select {
	case <-after:
		if conf.Docker {
			killCmd := exec.Command("docker", "rm", "-f", rid)
			if err := killCmd.Run(); nil != err {
				logger.Errorf("executes [docker rm -f " + rid + "] failed [" + err.Error() + "], this will cause resource leaking")
			}
		} else {
			cmd.Process.Kill()
		}

		channelRet["output"] = "\n<span class='stderr'>run program timeout in 5s</span>\n"
		kill = true
	case <-done:
		channelRet["output"] = "\n<span class='stderr'>run program complete</span>\n"
	}

	shouldExitBuf = true
	Processes.Remove(wSession, cmd.Process)
	logger.Debugf("User [%s, %s] done running [id=%s, file=%s, kill=%v]", wSession.UserId, sid, rid, filePath, kill)

	if nil != wsChannel {
		channelRet["cmd"] = "run-done"
		wsChannel.WriteJSON(&channelRet)
		wsChannel.Refresh()
	}
}

// StopHandler handles request of stopping a running process.
func StopHandler(w http.ResponseWriter, r *http.Request) {
	result := gulu.Ret.NewResult()
	defer gulu.Ret.RetResult(w, r, result)

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Code = -1

		return
	}

	sid := args["sid"].(string)
	pid := int(args["pid"].(float64))

	wSession := WideSessions.Get(sid)
	if nil == wSession {
		result.Code = -1

		return
	}

	Processes.Kill(wSession, pid)
}

// Add adds the specified process to the user process set.
func (procs *procs) Add(wSession *WideSession, proc *os.Process) {
	procMutex.Lock()
	defer procMutex.Unlock()

	sid := wSession.ID
	userProcesses := (*procs)[sid]

	userProcesses = append(userProcesses, proc)
	(*procs)[sid] = userProcesses

	// bind process with wide session
	wSession.SetProcesses(userProcesses)

	logger.Tracef("Session [%s] has [%d] processes", sid, len((*procs)[sid]))
}

// Remove removes the specified process from the user process set.
func (procs *procs) Remove(wSession *WideSession, proc *os.Process) {
	procMutex.Lock()
	defer procMutex.Unlock()

	sid := wSession.ID

	userProcesses := (*procs)[sid]

	var newProcesses []*os.Process
	for i, p := range userProcesses {
		if p.Pid == proc.Pid {
			newProcesses = append(userProcesses[:i], userProcesses[i+1:]...) // remove it
			(*procs)[sid] = newProcesses

			// bind process with wide session
			wSession.SetProcesses(newProcesses)

			logger.Tracef("Session [%s] has [%d] processes", sid, len((*procs)[sid]))

			return
		}
	}
}

// Kill kills a process specified by the given pid.
func (procs *procs) Kill(wSession *WideSession, pid int) {
	procMutex.Lock()
	defer procMutex.Unlock()

	sid := wSession.ID

	userProcesses := (*procs)[sid]

	for i, p := range userProcesses {
		if p.Pid == pid {
			if err := p.Kill(); nil != err {
				logger.Errorf("Kill a process [pid=%d] of user [%s, %s] failed [error=%v]", pid, wSession.UserId, sid, err)
			} else {
				var newProcesses []*os.Process

				newProcesses = append(userProcesses[:i], userProcesses[i+1:]...)
				(*procs)[sid] = newProcesses

				// bind process with wide session
				wSession.SetProcesses(newProcesses)

				logger.Debugf("Killed a process [pid=%d] of user [%s, %s]", pid, wSession.UserId, sid)
			}

			return
		}
	}
}
