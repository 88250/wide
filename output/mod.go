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

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
)

// GoModHandler handles request of go mod.
func GoModHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

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
		result.Succ = false

		return
	}

	sid := args["sid"].(string)

	filePath := args["file"].(string)
	curDir := filepath.Dir(filePath)
	curDirName := filepath.Base(curDir)

	cmd := exec.Command("go", "mod", "init", curDirName)
	cmd.Dir = curDir

	setCmdEnv(cmd, uid)

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
		// display "START [go mod]" in front-end browser

		channelRet["output"] = "<span class='start-mod'>" + i18n.Get(locale, "start-mod").(string) + "</span>\n"
		channelRet["cmd"] = "start-mod"

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

	runningId := rand.Int()

	logger.Debugf("User [%s, %s] is running [go mod] [runningId=%d]", uid, sid, runningId)
	channelRet = map[string]interface{}{}
	channelRet["cmd"] = "go mod"

	buf, _ := ioutil.ReadAll(reader)
	output := string(buf)
	err = cmd.Wait()
	if nil != err || 0 != cmd.ProcessState.ExitCode() {
		logger.Debugf("User [%s, %s] 's [go mod] [runningId=%d] has done (with error)", uid, sid, runningId)
		channelRet["output"] = "<span class='mod-error'>" + i18n.Get(locale, "mod-error").(string) + "</span>\n" + output
	} else {
		logger.Debugf("User [%s, %s] 's running [go mod] [runningId=%d] has done", uid, sid, runningId)
		channelRet["output"] = "<span class='mod-succ'>" + i18n.Get(locale, "mod-succ").(string) + "</span>\n" + output
	}

	wsChannel := session.OutputWS[sid]
	if nil != wsChannel {
		wsChannel.WriteJSON(&channelRet)
		wsChannel.Refresh()
	}
}
