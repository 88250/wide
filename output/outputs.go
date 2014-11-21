// Copyright (c) 2014, B3log
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

// Build, run and go tool manipulations.
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
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

const (
	lintSeverityError = "error"   // lint severity: error
	lintSeverityWarn  = "warning" // lint severity: warning
)

// Code lint.
type Lint struct {
	File     string `json:"file"`
	LineNo   int    `json:"lineNo"`
	Severity string `json:"severity"`
	Msg      string `json:"msg"`
}

// WSHandler handles request of creating output channel.
func WSHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query()["sid"][0]

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	wsChan := util.WSChannel{Sid: sid, Conn: conn, Request: r, Time: time.Now()}

	ret := map[string]interface{}{"output": "Ouput initialized", "cmd": "init-output"}
	err := wsChan.WriteJSON(&ret)
	if nil != err {
		return
	}

	session.OutputWS[sid] = &wsChan

	glog.V(4).Infof("Open a new [Output] with session [%s], %d", sid, len(session.OutputWS))
}

// RunHandler handles request of executing a binary file.
func RunHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false
	}

	sid := args["sid"].(string)
	wSession := session.WideSessions.Get(sid)
	if nil == wSession {
		data["succ"] = false
	}

	filePath := args["executable"].(string)
	curDir := filepath.Dir(filePath)

	cmd := exec.Command(filePath)
	cmd.Dir = curDir

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false
	}

	//outReader := bufio.NewReader(stdout)
	outReader := util.NewReader(stdout)
	errReader := bufio.NewReader(stderr)

	if err := cmd.Start(); nil != err {
		glog.Error(err)
		data["succ"] = false
	}

	wsChannel := session.OutputWS[sid]

	channelRet := map[string]interface{}{}

	if !data["succ"].(bool) {
		if nil != wsChannel {
			channelRet["cmd"] = "run-done"
			channelRet["output"] = ""

			err := wsChannel.WriteJSON(&channelRet)
			if nil != err {
				glog.Error(err)
				return
			}

			wsChannel.Refresh()
		}

		return
	}

	channelRet["pid"] = cmd.Process.Pid

	// add the process to user's process set
	processes.add(wSession, cmd.Process)

	go func(runningId int) {
		defer util.Recover()
		defer cmd.Wait()

		glog.V(5).Infof("Session [%s] is running [id=%d, file=%s]", sid, runningId, filePath)

		// push once for front-end to get the 'run' state and pid
		if nil != wsChannel {
			channelRet["cmd"] = "run"
			channelRet["output"] = ""
			err := wsChannel.WriteJSON(&channelRet)
			if nil != err {
				glog.Error(err)
				return
			}

			wsChannel.Refresh()
		}

		go func() {
			for {
				buf, err := outReader.ReadData()
				//buf, err := outReader.ReadString('\n')
				buf = strings.Replace(buf, "<", "&lt;", -1)
				buf = strings.Replace(buf, ">", "&gt;", -1)

				// TODO: fix the duplicated error

				if nil != err {
					// remove the exited process from user process set
					processes.remove(wSession, cmd.Process)

					glog.V(5).Infof("Session [%s] 's running [id=%d, file=%s] has done [stdout err]", sid, runningId, filePath)

					if nil != wsChannel {
						channelRet["cmd"] = "run-done"
						channelRet["output"] = buf
						err := wsChannel.WriteJSON(&channelRet)
						if nil != err {
							glog.Error(err)
							break
						}

						wsChannel.Refresh()
					}

					break
				} else {
					if nil != wsChannel {
						channelRet["cmd"] = "run"
						channelRet["output"] = buf
						err := wsChannel.WriteJSON(&channelRet)
						if nil != err {
							glog.Error(err)
							break
						}

						wsChannel.Refresh()
					}
				}
			}
		}()

		for {
			buf, err := errReader.ReadString('\n')
			buf = strings.Replace(buf, "<", "&lt;", -1)
			buf = strings.Replace(buf, ">", "&gt;", -1)

			if nil != err {
				// remove the exited process from user process set
				processes.remove(wSession, cmd.Process)

				glog.V(5).Infof("Session [%s] 's running [id=%d, file=%s] has done [stderr err]", sid, runningId, filePath)

				if nil != session.OutputWS[sid] {
					wsChannel := session.OutputWS[sid]

					channelRet["cmd"] = "run-done"
					channelRet["output"] = buf
					err := wsChannel.WriteJSON(&channelRet)
					if nil != err {
						glog.Error(err)
						break
					}

					wsChannel.Refresh()
				}

				break
			} else {
				if nil != session.OutputWS[sid] {
					wsChannel := session.OutputWS[sid]

					channelRet["cmd"] = "run"
					channelRet["output"] = buf
					err := wsChannel.WriteJSON(&channelRet)
					if nil != err {
						glog.Error(err)
						break
					}

					wsChannel.Refresh()
				}
			}
		}
	}(rand.Int())
}

// BuildHandler handles request of building.
func BuildHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	username := httpSession.Values["username"].(string)
	locale := conf.Wide.GetUser(username).Locale

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	sid := args["sid"].(string)

	filePath := args["file"].(string)
	curDir := filepath.Dir(filePath)

	fout, err := os.Create(filePath)

	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	code := args["code"].(string)

	fout.WriteString(code)

	if err := fout.Close(); nil != err {
		glog.Error(err)
		data["succ"] = false

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
	glog.V(5).Infof("go build for [%s]", executable)

	executable = filepath.Join(curDir, executable)

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	if !data["succ"].(bool) {
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
			glog.Error(err)
			return
		}

		wsChannel.Refresh()
	}

	reader := bufio.NewReader(io.MultiReader(stdout, stderr))

	if err := cmd.Start(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	go func(runningId int) {
		defer util.Recover()
		defer cmd.Wait()

		glog.V(5).Infof("Session [%s] is building [id=%d, dir=%s]", sid, runningId, curDir)

		// read all
		buf, _ := ioutil.ReadAll(reader)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "build"
		channelRet["executable"] = executable

		if 0 == len(buf) { // build success
			channelRet["nextCmd"] = args["nextCmd"]
			channelRet["output"] = "<span class='build-succ'>" + i18n.Get(locale, "build-succ").(string) + "</span>\n"

			go func() { // go install, for subsequent gocode lib-path
				cmd := exec.Command("go", "install")
				cmd.Dir = curDir

				setCmdEnv(cmd, username)

				out, _ := cmd.CombinedOutput()
				if len(out) > 0 {
					glog.Warning(string(out))
				}
			}()
		} else { // build error
			// build gutter lint

			errOut := string(buf)
			channelRet["output"] = "<span class='build-error'>" + i18n.Get(locale, "build-error").(string) + "</span>\n" + errOut

			lines := strings.Split(errOut, "\n")

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
					lineNo, _ = strconv.Atoi(left[:index])
					msg = left[index+2:]
				}

				lint := &Lint{
					File:     file,
					LineNo:   lineNo - 1,
					Severity: lintSeverityError,
					Msg:      msg,
				}

				lints = append(lints, lint)
			}

			channelRet["lints"] = lints
		}

		if nil != session.OutputWS[sid] {
			glog.V(5).Infof("Session [%s] 's build [id=%d, dir=%s] has done", sid, runningId, curDir)

			wsChannel := session.OutputWS[sid]
			err := wsChannel.WriteJSON(&channelRet)
			if nil != err {
				glog.Error(err)
			}

			wsChannel.Refresh()
		}

	}(rand.Int())
}

// GoTestHandler handles request of go test.
func GoTestHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	username := httpSession.Values["username"].(string)
	locale := conf.Wide.GetUser(username).Locale

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	sid := args["sid"].(string)

	filePath := args["file"].(string)
	curDir := filepath.Dir(filePath)

	cmd := exec.Command("go", "test", "-v")
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	if !data["succ"].(bool) {
		return
	}

	channelRet := map[string]interface{}{}

	if nil != session.OutputWS[sid] {
		// display "START [go test]" in front-end browser

		channelRet["output"] = "<span class='start-test'>" + i18n.Get(locale, "start-test").(string) + "</span>\n"
		channelRet["cmd"] = "start-test"

		wsChannel := session.OutputWS[sid]

		err := wsChannel.WriteJSON(&channelRet)
		if nil != err {
			glog.Error(err)
			return
		}

		wsChannel.Refresh()
	}

	reader := bufio.NewReader(io.MultiReader(stdout, stderr))

	if err := cmd.Start(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	go func(runningId int) {
		defer util.Recover()

		glog.V(5).Infof("Session [%s] is running [go test] [runningId=%d]", sid, runningId)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "go test"

		// read all
		buf, _ := ioutil.ReadAll(reader)

		// waiting for go test finished
		cmd.Wait()

		if !cmd.ProcessState.Success() {
			glog.V(5).Infof("Session [%s] 's running [go test] [runningId=%d] has done (with error)", sid, runningId)

			channelRet["output"] = "<span class='test-error'>" + i18n.Get(locale, "test-error").(string) + "</span>\n" + string(buf)
		} else {
			glog.V(5).Infof("Session [%s] 's running [go test] [runningId=%d] has done", sid, runningId)

			channelRet["output"] = "<span class='test-succ'>" + i18n.Get(locale, "test-succ").(string) + "</span>\n" + string(buf)
		}

		if nil != session.OutputWS[sid] {
			wsChannel := session.OutputWS[sid]

			err := wsChannel.WriteJSON(&channelRet)
			if nil != err {
				glog.Error(err)
			}

			wsChannel.Refresh()
		}
	}(rand.Int())
}

// GoInstallHandler handles request of go install.
func GoInstallHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	username := httpSession.Values["username"].(string)
	locale := conf.Wide.GetUser(username).Locale

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	sid := args["sid"].(string)

	filePath := args["file"].(string)
	curDir := filepath.Dir(filePath)

	cmd := exec.Command("go", "install")
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

	glog.V(5).Infof("go install %s", curDir)

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	if !data["succ"].(bool) {
		return
	}

	channelRet := map[string]interface{}{}

	if nil != session.OutputWS[sid] {
		// display "START [go install]" in front-end browser

		channelRet["output"] = "<span class='start-install'>" + i18n.Get(locale, "start-install").(string) + "</span>\n"
		channelRet["cmd"] = "start-install"

		wsChannel := session.OutputWS[sid]

		err := wsChannel.WriteJSON(&channelRet)
		if nil != err {
			glog.Error(err)
			return
		}

		wsChannel.Refresh()
	}

	reader := bufio.NewReader(io.MultiReader(stdout, stderr))

	if err := cmd.Start(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	go func(runningId int) {
		defer util.Recover()
		defer cmd.Wait()

		glog.V(5).Infof("Session [%s] is running [go install] [id=%d, dir=%s]", sid, runningId, curDir)

		// read all
		buf, _ := ioutil.ReadAll(reader)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "go install"

		if 0 != len(buf) { // build error
			// build gutter lint

			errOut := string(buf)
			lines := strings.Split(errOut, "\n")

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
					lineNo, _ = strconv.Atoi(left[:index])
					msg = left[index+2:]
				}

				lint := &Lint{
					File:     file,
					LineNo:   lineNo - 1,
					Severity: lintSeverityError,
					Msg:      msg,
				}

				lints = append(lints, lint)
			}

			channelRet["lints"] = lints

			channelRet["output"] = "<span class='install-error'>" + i18n.Get(locale, "install-error").(string) + "</span>\n" + errOut
		} else {
			channelRet["output"] = "<span class='install-succ'>" + i18n.Get(locale, "install-succ").(string) + "</span>\n"
		}

		if nil != session.OutputWS[sid] {
			glog.V(5).Infof("Session [%s] 's running [go install] [id=%d, dir=%s] has done", sid, runningId, curDir)

			wsChannel := session.OutputWS[sid]
			err := wsChannel.WriteJSON(&channelRet)
			if nil != err {
				glog.Error(err)
			}

			wsChannel.Refresh()
		}

	}(rand.Int())
}

// GoGetHandler handles request of go get.
func GoGetHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	username := httpSession.Values["username"].(string)
	locale := conf.Wide.GetUser(username).Locale

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

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
		glog.Error(err)
		data["succ"] = false

		return
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	if !data["succ"].(bool) {
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
			glog.Error(err)
			return
		}

		wsChannel.Refresh()
	}

	reader := bufio.NewReader(io.MultiReader(stdout, stderr))

	if err := cmd.Start(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	go func(runningId int) {
		defer util.Recover()
		defer cmd.Wait()

		glog.V(5).Infof("Session [%s] is running [go get] [runningId=%d]", sid, runningId)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "go get"

		// read all
		buf, _ := ioutil.ReadAll(reader)

		if 0 != len(buf) {
			glog.V(5).Infof("Session [%s] 's running [go get] [runningId=%d] has done (with error)", sid, runningId)

			channelRet["output"] = "<span class='get-error'>" + i18n.Get(locale, "get-error").(string) + "</span>\n" + string(buf)
		} else {
			glog.V(5).Infof("Session [%s] 's running [go get] [runningId=%d] has done", sid, runningId)

			channelRet["output"] = "<span class='get-succ'>" + i18n.Get(locale, "get-succ").(string) + "</span>\n"

		}

		if nil != session.OutputWS[sid] {
			wsChannel := session.OutputWS[sid]

			err := wsChannel.WriteJSON(&channelRet)
			if nil != err {
				glog.Error(err)
			}

			wsChannel.Refresh()
		}
	}(rand.Int())
}

// StopHandler handles request of stoping a running process.
func StopHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	sid := args["sid"].(string)
	pid := int(args["pid"].(float64))

	wSession := session.WideSessions.Get(sid)
	if nil == wSession {
		data["succ"] = false

		return
	}

	processes.kill(wSession, pid)
}

func setCmdEnv(cmd *exec.Cmd, username string) {
	userWorkspace := conf.Wide.GetUserWorkspace(username)

	cmd.Env = append(cmd.Env,
		"GOPATH="+userWorkspace,
		"GOOS="+runtime.GOOS,
		"GOARCH="+runtime.GOARCH,
		"GOROOT="+runtime.GOROOT(),
		"PATH="+os.Getenv("PATH"))
}
