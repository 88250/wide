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
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os/exec"
	"path/filepath"

	"github.com/b3log/gulu"
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/session"
)

// GoVetHandler handles request of go vet.
func GoVetHandler(w http.ResponseWriter, r *http.Request) {
	result := gulu.Ret.NewResult()
	defer gulu.Ret.RetResult(w, r, result)

	httpSession, _ := session.HTTPSession.Get(r, session.CookieName)
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	uid := httpSession.Values["uid"].(string)
	locale := conf.GetUser(uid).Locale

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Code = -1

		return
	}

	sid := args["sid"].(string)

	filePath := args["file"].(string)
	curDir := filepath.Dir(filePath)

	cmd := exec.Command("go", "vet", ".")
	cmd.Dir = curDir

	setCmdEnv(cmd, uid)

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		logger.Error(err)
		result.Code = -1

		return
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		logger.Error(err)
		result.Code = -1

		return
	}

	if 0 != result.Code {
		return
	}

	channelRet := map[string]interface{}{}

	if nil != session.OutputWS[sid] {
		// display "START [go vet]" in front-end browser

		channelRet["output"] = "<span class='start-vet'>" + i18n.Get(locale, "start-vet").(string) + "</span>\n"
		channelRet["cmd"] = "start-vet"

		wsChannel := session.OutputWS[sid]

		err := wsChannel.WriteJSON(&channelRet)
		if nil != err {
			logger.Warn(err)
			return
		}

		wsChannel.Refresh()
	}

	reader := bufio.NewReader(io.MultiReader(stdout, stderr))

	if err := cmd.Start(); nil != err {
		logger.Error(err)
		result.Code = -1

		return
	}

	go func(runningId int) {
		defer gulu.Panic.Recover()

		logger.Debugf("User [%s, %s] is running [go vet] [runningId=%d]", uid, sid, runningId)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "go vet"

		// read all
		buf, _ := ioutil.ReadAll(reader)

		// waiting for go vet finished
		cmd.Wait()

		if !cmd.ProcessState.Success() {
			logger.Debugf("User [%s, %s] 's running [go vet] [runningId=%d] has done (with error)", uid, sid, runningId)

			channelRet["output"] = "<span class='vet-error'>" + i18n.Get(locale, "vet-error").(string) + "</span>\n" + string(buf)
		} else {
			logger.Debugf("User [%s, %s] 's running [go vet] [runningId=%d] has done", uid, sid, runningId)

			channelRet["output"] = "<span class='vet-succ'>" + i18n.Get(locale, "vet-succ").(string) + "</span>\n" + string(buf)
		}

		if nil != session.OutputWS[sid] {
			wsChannel := session.OutputWS[sid]

			err := wsChannel.WriteJSON(&channelRet)
			if nil != err {
				logger.Warn(err)
			}

			wsChannel.Refresh()
		}
	}(rand.Int())
}
