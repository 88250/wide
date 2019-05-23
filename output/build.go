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
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

	httpSession, _ := session.HTTPSession.Get(r, session.CookieName)
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	uid := httpSession.Values["uid"].(string)
	user := conf.GetUser(uid)
	locale := user.Locale

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	sid := args["sid"].(string)
	filePath := args["file"].(string)
	if util.Go.IsAPI(filePath) || !session.CanAccess(uid, filePath) {
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
	if _, err := fout.WriteString(code); nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}
	fout.Close()

	channelRet := map[string]interface{}{}
	if nil != session.OutputWS[sid] {
		// display "START [go build]" in front-end browser

		msg := i18n.Get(locale, "start-build").(string)
		msg = strings.Replace(msg, "build]", "build "+fmt.Sprint(user.BuildArgs(runtime.GOOS))+"]", 1)

		channelRet["output"] = "<span class='start-build'>" + msg + "</span>\n"
		channelRet["cmd"] = "start-build"

		wsChannel := session.OutputWS[sid]
		wsChannel.WriteJSON(&channelRet)
		wsChannel.Refresh()
	}

	var goModCmd *exec.Cmd
	if !util.File.IsExist(filepath.Join(curDir, "go.mod")) {
		curDirName := filepath.Base(curDir)
		goModCmd = exec.Command("go", "mod", "init", curDirName)
	} else {
		goModCmd = exec.Command("go", "mod", "tidy")
	}
	goModCmd.Dir = curDir
	setCmdEnv(goModCmd, uid)
	outputBytes, err := goModCmd.CombinedOutput()
	output := string(outputBytes)
	if nil != err && strings.Contains(output, "go.mod already exists") {
		logger.Error(err.Error() + ": " + output)
		result.Succ = false

		return
	}

	var goBuildArgs []string
	goBuildArgs = append(goBuildArgs, "build")
	goBuildArgs = append(goBuildArgs, user.BuildArgs(runtime.GOOS)...)
	if !util.Str.Contains("-i", goBuildArgs) {
		goBuildArgs = append(goBuildArgs, "-i")
	}

	cmd := exec.Command("go", goBuildArgs...)
	cmd.Dir = curDir
	setCmdEnv(cmd, uid)

	suffix := ""
	if util.OS.IsWindows() {
		suffix = ".exe"
	}
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

	if err := cmd.Start(); nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	channelRet["cmd"] = "build"
	channelRet["executable"] = executable

	outReader := bufio.NewReader(stdout)

	/////////
	go func() {
		defer util.Recover()

		for {
			wsChannel := session.OutputWS[sid]
			if nil == wsChannel {
				break
			}

			line, err := outReader.ReadString('\n')
			if io.EOF == err {
				break
			}

			if nil != err {
				logger.Warn(err)

				break
			}

			channelRet["output"] = line

			err = wsChannel.WriteJSON(&channelRet)
			if nil != err {
				logger.Warn(err)

				break
			}

			wsChannel.Refresh()
		}
	}()

	errReader := bufio.NewReader(stderr)
	var lines []string
	for {
		wsChannel := session.OutputWS[sid]
		if nil == wsChannel {
			break
		}

		line, err := errReader.ReadString('\n')
		if io.EOF == err {
			break
		}

		lines = append(lines, line)

		if nil != err {
			logger.Warn(err)

			break
		}

		// path process
		errOutWithPath := parsePath(curDir, line)
		channelRet["output"] = "<span class='stderr'>" + errOutWithPath + "</span>"

		err = wsChannel.WriteJSON(&channelRet)
		if nil != err {
			logger.Warn(err)
			break
		}

		wsChannel.Refresh()
	}

	if nil == cmd.Wait() {
		channelRet["nextCmd"] = args["nextCmd"]
		channelRet["output"] = "<span class='build-succ'>" + i18n.Get(locale, "build-succ").(string) + "</span>\n"
	} else {
		channelRet["output"] = "<span class='build-error'>" + i18n.Get(locale, "build-error").(string) + "</span>\n"

		// lint process
		if lines[0][0] == '#' {
			lines = lines[1:] // skip the first line
		}

		lints := []*Lint{}

		for _, line := range lines {
			if len(line) < 1 || !strings.Contains(line, ":") {
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
				File:     filepath.ToSlash(filepath.Join(curDir, file)),
				LineNo:   lineNo - 1,
				Severity: lintSeverityError,
				Msg:      msg,
			}

			lints = append(lints, lint)
		}

		channelRet["lints"] = lints
	}

	wsChannel := session.OutputWS[sid]
	if nil == wsChannel {
		return
	}
	err = wsChannel.WriteJSON(&channelRet)
	if nil != err {
		logger.Warn(err)
	}

	wsChannel.Refresh()
}
