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

// Package output includes build, run and go tool related manipulations.
package output

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/88250/gulu"
	"github.com/88250/wide/conf"
	"github.com/88250/wide/session"
	"github.com/88250/wide/util"
	"github.com/gorilla/websocket"
)

const (
	lintSeverityError = "error"   // lint severity: error
	lintSeverityWarn  = "warning" // lint severity: warning
)

// Logger.
var logger = gulu.Log.NewLogger(os.Stdout)

// Lint represents a code lint.
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

	logger.Tracef("Open a new [Output] with session [%s], %d", sid, len(session.OutputWS))
}

// parsePath parses file path in the specified outputLine, and returns new line with front-end friendly.
func parsePath(curDir, outputLine string) string {
	index := strings.Index(outputLine, " ")
	if -1 == index || index >= len(outputLine) {
		return outputLine
	}

	pathPart := outputLine[:index]
	msgPart := outputLine[index:]

	parts := strings.Split(pathPart, ":")
	if len(parts) < 2 { // no file path info (line & column) found
		return outputLine
	}

	file := parts[0]
	line := parts[1]
	if _, err := strconv.Atoi(line); nil != err {
		return outputLine
	}

	column := "0"
	hasColumn := 4 == len(parts)
	if hasColumn {
		column = parts[2]
	}

	tagStart := `<span class="path" data-path="` + filepath.ToSlash(filepath.Join(curDir, file)) + `" data-line="` + line +
		`" data-column="` + column + `">`
	text := file + ":" + line
	if hasColumn {
		text += ":" + column
	}
	tagEnd := "</span>:"

	return tagStart + text + tagEnd + msgPart
}

func setCmdEnv(cmd *exec.Cmd, uid string) {
	userWorkspace := conf.GetUserWorkspace(uid)
	cache, err := os.UserCacheDir()
	if nil != err {
		logger.Warnf("Get user cache dir failed [" + err.Error() + "]")
		cache = os.TempDir()
	}

	cmd.Env = append(cmd.Env,
		"GOPROXY=https://goproxy.cn",
		"GO111MODULE=on",
		"GOPATH="+userWorkspace,
		"GOOS="+runtime.GOOS,
		"GOARCH="+runtime.GOARCH,
		"GOROOT="+runtime.GOROOT(),
		"GOCACHE="+cache,
		"PATH="+os.Getenv("PATH"))

	if gulu.OS.IsWindows() {
		// FIXME: for some weird issues on Windows, such as: The requested service provider could not be loaded or initialized.
		cmd.Env = append(cmd.Env, os.Environ()...)
	} else {
		// 编译链接时找不到依赖的动态库 https://github.com/b3log/wide/issues/352
		cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH="+os.Getenv("LD_LIBRARY_PATH"))
	}
}
