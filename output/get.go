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
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os/exec"
	"path/filepath"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
)

// GoGetHandler handles request of go get.
func GoGetHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)
	locale := conf.GetUser(username).Locale

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	sid := args["sid"].(string)

	filePath := args["file"].(string)
	curDir := filepath.Dir(filePath)

	cmd := exec.Command("go", "get")
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	if !result.Succ {
		return
	}

	channelRet := map[string]interface{}{}

	if nil != session.OutputWS[sid] {
		// display "START [go get]" in front-end browser

		channelRet["output"] = "<span class='start-get'>" + i18n.Get(locale, "start-get").(string) + "</span>\n"
		channelRet["cmd"] = "start-get"

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
		result.Succ = false

		return
	}

	go func(runningId int) {
		defer util.Recover()
		defer cmd.Wait()

		logger.Debugf("User [%s, %s] is running [go get] [runningId=%d]", username, sid, runningId)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "go get"

		// read all
		buf, _ := ioutil.ReadAll(reader)

		if 0 != len(buf) {
			logger.Debugf("User [%s, %s] 's [go get] [runningId=%d] has done (with error)", username, sid, runningId)

			channelRet["output"] = "<span class='get-error'>" + i18n.Get(locale, "get-error").(string) + "</span>\n" + string(buf)
		} else {
			logger.Debugf("User [%s, %s] 's running [go get] [runningId=%d] has done", username, sid, runningId)

			channelRet["output"] = "<span class='get-succ'>" + i18n.Get(locale, "get-succ").(string) + "</span>\n"
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
