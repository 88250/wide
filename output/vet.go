// Copyright (c) 2014-2015, b3log.org
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

// GoVetHandler handles request of go vet.
func GoVetHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

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
		data["succ"] = false

		return
	}

	sid := args["sid"].(string)

	filePath := args["file"].(string)
	curDir := filepath.Dir(filePath)

	cmd := exec.Command("go", "vet", ".")
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		logger.Error(err)
		data["succ"] = false

		return
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		logger.Error(err)
		data["succ"] = false

		return
	}

	if !data["succ"].(bool) {
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
		data["succ"] = false

		return
	}

	go func(runningId int) {
		defer util.Recover()

		logger.Debugf("User [%s, %s] is running [go vet] [runningId=%d]", username, sid, runningId)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "go vet"

		// read all
		buf, _ := ioutil.ReadAll(reader)

		// waiting for go vet finished
		cmd.Wait()

		if !cmd.ProcessState.Success() {
			logger.Debugf("User [%s, %s] 's running [go vet] [runningId=%d] has done (with error)", username, sid, runningId)

			channelRet["output"] = "<span class='vet-error'>" + i18n.Get(locale, "vet-error").(string) + "</span>\n" + string(buf)
		} else {
			logger.Debugf("User [%s, %s] 's running [go vet] [runningId=%d] has done", username, sid, runningId)

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
