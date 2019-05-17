// Copyright (c) 2014-2019, b3log.org & hacpai.com
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

// Package shell include playground related mainipulations.
package playground

import (
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/log"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"github.com/gorilla/websocket"
)

// Logger.
var logger = log.NewLogger(os.Stdout)

// IndexHandler handles request of Playground index.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// create a HTTP session
	httpSession, _ := session.HTTPSession.Get(r, session.CookieName)
	if httpSession.IsNew {
		httpSession.Values["id"] = strconv.Itoa(rand.Int())
		httpSession.Values["uid"] = "playground"
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	uid := httpSession.Values["uid"].(string)

	locale := conf.Wide.Locale

	// try to load file
	code := conf.HelloWorld
	fileName := "8b7cc38b4c12e6fde5c4d15a4f2f32e5.go" // MD5 of HelloWorld.go

	if strings.HasSuffix(r.URL.Path, ".go") {
		fileNameArg := r.URL.Path[len("/playground/"):]
		filePath := filepath.Clean(conf.Wide.Data+ "/playground" + fileNameArg)

		bytes, err := ioutil.ReadFile(filePath)
		if nil != err {
			logger.Warn(err)
		} else {
			code = string(bytes)
			fileName = fileNameArg
		}
	}

	query := r.URL.Query()
	embed := false
	embedArg, ok := query["embed"]
	if ok && "true" == embedArg[0] {
		embed = true
	}

	disqus := false
	disqusArg, ok := query["disqus"]
	if ok && "true" == disqusArg[0] {
		disqus = true
	}

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale,
		"sid": session.WideSessions.GenId(), "pathSeparator": conf.PathSeparator,
		"codeMirrorVer": conf.CodeMirrorVer,
		"code":          template.HTML(code), "ver": conf.WideVersion, "year": time.Now().Year(),
		"embed": embed, "disqus": disqus, "fileName": fileName}

	wideSessions := session.WideSessions.GetByUserId(uid)

	logger.Debugf("User [%s] has [%d] sessions", uid, len(wideSessions))

	t, err := template.ParseFiles("views/playground/index.html")

	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	t.Execute(w, model)
}

// WSHandler handles request of creating Playground channel.
func WSHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query()["sid"][0]

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	wsChan := util.WSChannel{Sid: sid, Conn: conn, Request: r, Time: time.Now()}

	ret := map[string]interface{}{"output": "Playground initialized", "cmd": "init-playground"}
	err := wsChan.WriteJSON(&ret)
	if nil != err {
		return
	}

	session.PlaygroundWS[sid] = &wsChan

	logger.Tracef("Open a new [PlaygroundWS] with session [%s], %d", sid, len(session.PlaygroundWS))
}
