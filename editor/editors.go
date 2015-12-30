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

// Package editor includes editor related manipulations.
package editor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/file"
	"github.com/b3log/wide/log"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"github.com/gorilla/websocket"
)

// Logger.
var logger = log.NewLogger(os.Stdout)

// WSHandler handles request of creating editor channel.
// XXX: NOT used at present
func WSHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	sid := httpSession.Values["id"].(string)

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	editorChan := util.WSChannel{Sid: sid, Conn: conn, Request: r, Time: time.Now()}

	ret := map[string]interface{}{"output": "Editor initialized", "cmd": "init-editor"}
	err := editorChan.WriteJSON(&ret)
	if nil != err {
		return
	}

	session.EditorWS[sid] = &editorChan

	logger.Tracef("Open a new [Editor] with session [%s], %d", sid, len(session.EditorWS))

	args := map[string]interface{}{}
	for {
		if err := session.EditorWS[sid].ReadJSON(&args); err != nil {
			return
		}

		code := args["code"].(string)
		line := int(args["cursorLine"].(float64))
		ch := int(args["cursorCh"].(float64))

		offset := getCursorOffset(code, line, ch)

		logger.Tracef("offset: %d", offset)

		gocode := util.Go.GetExecutableInGOBIN("gocode")
		argv := []string{"-f=json", "autocomplete", strconv.Itoa(offset)}

		var output bytes.Buffer

		cmd := exec.Command(gocode, argv...)
		cmd.Stdout = &output

		stdin, _ := cmd.StdinPipe()
		cmd.Start()
		stdin.Write([]byte(code))
		stdin.Close()
		cmd.Wait()

		ret = map[string]interface{}{"output": string(output.Bytes()), "cmd": "autocomplete"}

		if err := session.EditorWS[sid].WriteJSON(&ret); err != nil {
			logger.Error("Editor WS ERROR: " + err.Error())
			return
		}
	}
}

// AutocompleteHandler handles request of code autocompletion.
func AutocompleteHandler(w http.ResponseWriter, r *http.Request) {
	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	session, _ := session.HTTPSession.Get(r, "wide-session")
	if session.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := session.Values["username"].(string)

	path := args["path"].(string)

	fout, err := os.Create(path)

	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	code := args["code"].(string)
	fout.WriteString(code)

	if err := fout.Close(); nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	line := int(args["cursorLine"].(float64))
	ch := int(args["cursorCh"].(float64))

	offset := getCursorOffset(code, line, ch)

	logger.Tracef("offset: %d", offset)

	userWorkspace := conf.GetUserWorkspace(username)
	workspaces := filepath.SplitList(userWorkspace)
	libPath := ""
	for _, workspace := range workspaces {
		userLib := workspace + conf.PathSeparator + "pkg" + conf.PathSeparator +
			runtime.GOOS + "_" + runtime.GOARCH
		libPath += userLib + conf.PathListSeparator
	}

	logger.Tracef("gocode set lib-path [%s]", libPath)

	// FIXME: using gocode set lib-path has some issues while accrossing workspaces
	gocode := util.Go.GetExecutableInGOBIN("gocode")
	exec.Command(gocode, []string{"set", "lib-path", libPath}...).Run()

	argv := []string{"-f=json", "--in=" + path, "autocomplete", strconv.Itoa(offset)}
	cmd := exec.Command(gocode, argv...)

	output, err := cmd.CombinedOutput()
	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

// GetExprInfoHandler handles request of getting expression infomation.
func GetExprInfoHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	session, _ := session.HTTPSession.Get(r, "wide-session")
	username := session.Values["username"].(string)

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	path := args["path"].(string)
	curDir := filepath.Dir(path)
	filename := filepath.Base(path)

	fout, err := os.Create(path)

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

	line := int(args["cursorLine"].(float64))
	ch := int(args["cursorCh"].(float64))

	offset := getCursorOffset(code, line, ch)

	logger.Tracef("offset [%d]", offset)

	ideStub := util.Go.GetExecutableInGOBIN("gotools")
	argv := []string{"types", "-pos", filename + ":" + strconv.Itoa(offset), "-info", "."}
	cmd := exec.Command(ideStub, argv...)
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

	output, err := cmd.CombinedOutput()
	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	exprInfo := strings.TrimSpace(string(output))
	if "" == exprInfo {
		result.Succ = false

		return
	}

	result.Data = exprInfo
}

// FindDeclarationHandler handles request of finding declaration.
func FindDeclarationHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	session, _ := session.HTTPSession.Get(r, "wide-session")
	if session.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := session.Values["username"].(string)

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	path := args["path"].(string)
	curDir := filepath.Dir(path)
	filename := filepath.Base(path)

	fout, err := os.Create(path)

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

	line := int(args["cursorLine"].(float64))
	ch := int(args["cursorCh"].(float64))

	offset := getCursorOffset(code, line, ch)

	logger.Tracef("offset [%d]", offset)

	ideStub := util.Go.GetExecutableInGOBIN("gotools")
	argv := []string{"types", "-pos", filename + ":" + strconv.Itoa(offset), "-def", "."}
	cmd := exec.Command(ideStub, argv...)
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

	output, err := cmd.CombinedOutput()
	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	found := strings.TrimSpace(string(output))
	if "" == found {
		result.Succ = false

		return
	}

	part := found[:strings.LastIndex(found, ":")]
	cursorSep := strings.LastIndex(part, ":")
	path = found[:cursorSep]

	cursorLine, _ := strconv.Atoi(found[cursorSep+1 : strings.LastIndex(found, ":")])
	cursorCh, _ := strconv.Atoi(found[strings.LastIndex(found, ":")+1:])

	data := map[string]interface{}{}
	result.Data = &data

	data["path"] = filepath.ToSlash(path)
	data["cursorLine"] = cursorLine
	data["cursorCh"] = cursorCh
}

// FindUsagesHandler handles request of finding usages.
func FindUsagesHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	session, _ := session.HTTPSession.Get(r, "wide-session")
	if session.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := session.Values["username"].(string)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	filePath := args["path"].(string)
	curDir := filepath.Dir(filePath)
	filename := filepath.Base(filePath)

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

	line := int(args["cursorLine"].(float64))
	ch := int(args["cursorCh"].(float64))

	offset := getCursorOffset(code, line, ch)
	logger.Tracef("offset [%d]", offset)

	ideStub := util.Go.GetExecutableInGOBIN("gotools")
	argv := []string{"types", "-pos", filename + ":" + strconv.Itoa(offset), "-use", "."}
	cmd := exec.Command(ideStub, argv...)
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

	output, err := cmd.CombinedOutput()
	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	out := strings.TrimSpace(string(output))
	if "" == out {
		result.Succ = false

		return
	}

	founds := strings.Split(out, "\n")
	usages := []*file.Snippet{}
	for _, found := range founds {
		found = strings.TrimSpace(found)

		part := found[:strings.LastIndex(found, ":")]
		cursorSep := strings.LastIndex(part, ":")
		path := filepath.ToSlash(found[:cursorSep])
		cursorLine, _ := strconv.Atoi(found[cursorSep+1 : strings.LastIndex(found, ":")])
		cursorCh, _ := strconv.Atoi(found[strings.LastIndex(found, ":")+1:])

		usage := &file.Snippet{Path: path, Line: cursorLine, Ch: cursorCh, Contents: []string{""}}
		usages = append(usages, usage)
	}

	result.Data = usages
}

// getCursorOffset calculates the cursor offset.
//
// line is the line number, starts with 0 that means the first line
// ch is the column number, starts with 0 that means the first column
func getCursorOffset(code string, line, ch int) (offset int) {
	lines := strings.Split(code, "\n")

	// calculate sum length of lines before
	for i := 0; i < line; i++ {
		offset += len(lines[i])
	}

	// calculate length of the current line and column
	curLine := lines[line]
	var buffer bytes.Buffer
	r := []rune(curLine)
	for i := 0; i < ch; i++ {
		buffer.WriteString(string(r[i]))
	}

	offset += len(buffer.String()) // append length of current line
	offset += line                 // append number of '\n'

	return offset
}

func setCmdEnv(cmd *exec.Cmd, username string) {
	userWorkspace := conf.GetUserWorkspace(username)

	cmd.Env = append(cmd.Env,
		"GOPATH="+userWorkspace,
		"GOOS="+runtime.GOOS,
		"GOARCH="+runtime.GOARCH,
		"GOROOT="+runtime.GOROOT(),
		"PATH="+os.Getenv("PATH"))
}
