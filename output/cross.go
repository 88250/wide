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
	"strconv"
	"strings"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
)

// CrossCompilationHandler handles request of cross compilation.
func CrossCompilationHandler(w http.ResponseWriter, r *http.Request) {
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
	filePath := args["path"].(string)

	if util.Go.IsAPI(filePath) || !session.CanAccess(uid, filePath) {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	platform := args["platform"].(string)
	goos := strings.Split(platform, "_")[0]
	goarch := strings.Split(platform, "_")[1]

	curDir := filepath.Dir(filePath)

	suffix := ""
	if "windows" == goos {
		suffix = ".exe"
	}

	user := conf.GetUser(uid)
	goBuildArgs := []string{}
	goBuildArgs = append(goBuildArgs, "build")
	goBuildArgs = append(goBuildArgs, user.BuildArgs(goos)...)

	cmd := exec.Command("go", goBuildArgs...)
	cmd.Dir = curDir

	setCmdEnv(cmd, uid)

	for i, env := range cmd.Env {
		if strings.HasPrefix(env, "GOOS=") {
			cmd.Env[i] = "GOOS=" + goos

			continue
		}

		if strings.HasPrefix(env, "GOARCH=") {
			cmd.Env[i] = "GOARCH=" + goarch

			continue
		}
	}

	executable := filepath.Base(curDir) + suffix
	executable = filepath.Join(curDir, executable)
	name := filepath.Base(curDir) + "-" + goos + "-" + goarch

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

		// read all
		buf, _ := ioutil.ReadAll(reader)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "cross-build"
		channelRet["executable"] = executable
		channelRet["name"] = name

		if 0 == len(buf) { // build success
			channelRet["output"] = "<span class='build-succ'>" + i18n.Get(locale, "build-succ").(string) + "</span>\n"
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
			wsChannel := session.OutputWS[sid]
			err := wsChannel.WriteJSON(&channelRet)
			if nil != err {
				logger.Warn(err)
			}

			wsChannel.Refresh()
		}

	}(rand.Int())
}
