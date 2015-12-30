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
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
)

// BuildHandler handles request of building.
func BuildHandler(w http.ResponseWriter, r *http.Request) {
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

	if util.Go.IsAPI(filePath) || !session.CanAccess(username, filePath) {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	curDir := filepath.Dir(filePath)

	fout, err := os.Create(filePath)

	if nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	code := args["code"].(string)

	fout.WriteString(code)

	if err := fout.Close(); nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	suffix := ""
	if util.OS.IsWindows() {
		suffix = ".exe"
	}

	cmd := exec.Command("go", "build")
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

	executable := filepath.Base(curDir) + suffix
	executable = filepath.Join(curDir, executable)

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
		// display "START [go build]" in front-end browser

		channelRet["output"] = "<span class='start-build'>" + i18n.Get(locale, "start-build").(string) + "</span>\n"
		channelRet["cmd"] = "start-build"

		wsChannel := session.OutputWS[sid]

		err := wsChannel.WriteJSON(&channelRet)
		if nil != err {
			logger.Error(err)
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

		// logger.Debugf("User [%s, %s] is building [id=%d, dir=%s]", username, sid, runningId, curDir)

		// read all
		buf, _ := ioutil.ReadAll(reader)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "build"
		channelRet["executable"] = executable

		if 0 == len(buf) { // build success
			channelRet["nextCmd"] = args["nextCmd"]
			channelRet["output"] = "<span class='build-succ'>" + i18n.Get(locale, "build-succ").(string) + "</span>\n"

			go func() { // go install, for subsequent gocode lib-path
				defer util.Recover()

				cmd := exec.Command("go", "install")
				cmd.Dir = curDir

				setCmdEnv(cmd, username)

				out, _ := cmd.CombinedOutput()
				if len(out) > 0 {
					logger.Warn(string(out))
				}
			}()
		} else { // build error
			// build gutter lint

			errOut := string(buf)
			lines := strings.Split(errOut, "\n")

			// path process
			var errOutWithPath string
			for _, line := range lines {
				errOutWithPath += parsePath(curDir, line) + "\n"
			}

			channelRet["output"] = "<span class='build-error'>" + i18n.Get(locale, "build-error").(string) + "</span>\n" +
				"<span class='stderr'>" + errOutWithPath + "</span>"

			// lint process

			if lines[0][0] == '#' {
				lines = lines[1:] // skip the first line
			}

			lints := []*Lint{}

			for _, line := range lines {
				if len(line) < 1 {
					continue
				}

				if line[0] == '\t' {
					// append to the last lint
					last := len(lints)
					msg := lints[last-1].Msg
					msg += line

					lints[last-1].Msg = msg

					continue
				}

				file := line[:strings.Index(line, ":")]
				left := line[strings.Index(line, ":")+1:]
				index := strings.Index(left, ":")
				lineNo := 0
				msg := left
				if index >= 0 {
					lineNo, err = strconv.Atoi(left[:index])

					if nil != err {
						continue
					}

					msg = left[index+2:]
				}

				lint := &Lint{
					File:     filepath.Join(curDir, file),
					LineNo:   lineNo - 1,
					Severity: lintSeverityError,
					Msg:      msg,
				}

				lints = append(lints, lint)
			}

			channelRet["lints"] = lints
		}

		if nil != session.OutputWS[sid] {
			// logger.Debugf("User [%s, %s] 's build [id=%d, dir=%s] has done", username, sid, runningId, curDir)

			wsChannel := session.OutputWS[sid]
			err := wsChannel.WriteJSON(&channelRet)
			if nil != err {
				logger.Warn(err)
			}

			wsChannel.Refresh()
		}

	}(rand.Int())
}
