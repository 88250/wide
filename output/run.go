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

const (
	outputBufMax  = 128 // 128 string(rune)
	outputTimeout = 100 // 100ms
)

type outputBuf struct {
	content     string
	millisecond int64
}

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
	curDir := filepath.Dir(filePath)

	cmd := exec.Command(filePath)
	cmd.Dir = curDir

	if conf.Docker {
		SetNamespace(cmd)
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
		if nil != wsChannel {
			channelRet["cmd"] = "run-done"
			channelRet["output"] = ""

			err := wsChannel.WriteJSON(&channelRet)
			if nil != err {
				logger.Warn(err)
				return
			}

			wsChannel.Refresh()
		}

		return
	}

	channelRet["pid"] = cmd.Process.Pid

	// add the process to user's process set
	Processes.Add(wSession, cmd.Process)

	go func(runningId int) {
		defer util.Recover()
		defer cmd.Wait()

		logger.Debugf("User [%s, %s] is running [id=%d, file=%s]", wSession.Username, sid, runningId, filePath)

		// push once for front-end to get the 'run' state and pid
		if nil != wsChannel {
			channelRet["cmd"] = "run"
			channelRet["output"] = ""
			err := wsChannel.WriteJSON(&channelRet)
			if nil != err {
				logger.Warn(err)
				return
			}

			wsChannel.Refresh()
		}

		go func() {
			defer util.Recover()

			buf := outputBuf{}

			for {
				wsChannel := session.OutputWS[sid]
				if nil == wsChannel {
					break
				}

				r, _, err := outReader.ReadRune()

				if nil != err {
					// remove the exited process from user's process set
					Processes.Remove(wSession, cmd.Process)

					logger.Debugf("User [%s, %s] 's running [id=%d, file=%s] has done [stdout %v], ",
						wSession.Username, sid, runningId, filePath, err)

					channelRet["cmd"] = "run-done"
					channelRet["output"] = buf.content
					err := wsChannel.WriteJSON(&channelRet)
					if nil != err {
						logger.Warn(err)
						break
					}

					wsChannel.Refresh()

					break
				}

				oneRuneStr := string(r)
				oneRuneStr = strings.Replace(oneRuneStr, "<", "&lt;", -1)
				oneRuneStr = strings.Replace(oneRuneStr, ">", "&gt;", -1)

				buf.content += oneRuneStr

				now := time.Now().UnixNano() / int64(time.Millisecond)

				if 0 == buf.millisecond {
					buf.millisecond = now
				}

				if now-outputTimeout >= buf.millisecond || len(buf.content) > outputBufMax || oneRuneStr == "\n" {
					channelRet["cmd"] = "run"
					channelRet["output"] = buf.content

					buf = outputBuf{} // a new buffer

					err = wsChannel.WriteJSON(&channelRet)
					if nil != err {
						logger.Warn(err)
						break
					}

					wsChannel.Refresh()
				}
			}
		}()

		buf := outputBuf{}
		for {
			r, _, err := errReader.ReadRune()

			wsChannel := session.OutputWS[sid]
			if nil != err || nil == wsChannel {
				break
			}

			oneRuneStr := string(r)
			oneRuneStr = strings.Replace(oneRuneStr, "<", "&lt;", -1)
			oneRuneStr = strings.Replace(oneRuneStr, ">", "&gt;", -1)

			buf.content += oneRuneStr

			now := time.Now().UnixNano() / int64(time.Millisecond)

			if 0 == buf.millisecond {
				buf.millisecond = now
			}

			if now-outputTimeout >= buf.millisecond || len(buf.content) > outputBufMax || oneRuneStr == "\n" {
				channelRet["cmd"] = "run"
				channelRet["output"] = "<span class='stderr'>" + buf.content + "</span>"

				buf = outputBuf{} // a new buffer

				err = wsChannel.WriteJSON(&channelRet)
				if nil != err {
					logger.Warn(err)
					break
				}

				wsChannel.Refresh()
			}
		}
	}(rand.Int())
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
