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

package session

import (
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"github.com/88250/gulu"
	"github.com/88250/wide/conf"
	"github.com/88250/wide/i18n"
)

var states = map[string]string{}

// LoginRedirectHandler redirects to HacPai auth page.
func LoginRedirectHandler(w http.ResponseWriter, r *http.Request) {
	loginAuthURL := "https://hacpai.com/login?goto=" + conf.Wide.Server + "/login/callback"

	state := gulu.Rand.String(16)
	states[state] = state
	path := loginAuthURL + "?state=" + state
	http.Redirect(w, r, path, http.StatusSeeOther)
}

func LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if _, exist := states[state]; !exist {
		http.Error(w, "Get state param failed", http.StatusBadRequest)

		return
	}
	delete(states, state)

	userId := r.URL.Query().Get("userId")
	userName := r.URL.Query().Get("userName")
	avatar := r.URL.Query().Get("avatar")

	user := conf.GetUser(userId)
	if nil == user {
		msg := addUser(userId, userName, avatar)
		if userCreated != msg {
			result := gulu.Ret.NewResult()
			result.Code = -1
			result.Msg = msg
			gulu.Ret.RetResult(w, r, result)

			return
		}
	}

	// create a HTTP session
	httpSession, _ := HTTPSession.Get(r, CookieName)
	httpSession.Values["uid"] = userId
	httpSession.Values["id"] = strconv.Itoa(rand.Int())
	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// LoginHandler handles request of show login page.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(conf.Wide.Locale),
		"locale": conf.Wide.Locale, "ver": conf.WideVersion, "year": time.Now().Year()}

	t, err := template.ParseFiles("views/login.html")
	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	t.Execute(w, model)
}

// LogoutHandler handles request of user logout (exit).
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	result := gulu.Ret.NewResult()
	defer gulu.Ret.RetResult(w, r, result)

	httpSession, _ := HTTPSession.Get(r, CookieName)

	httpSession.Options.MaxAge = -1
	httpSession.Save(r, w)
}

// addUser add a user with the specified user id, username and avatar.
//
//  1. create the user's workspace
//  2. generate 'Hello, 世界' demo code in the workspace (a console version and a HTTP version)
//  3. update the user customized configurations, such as style.css
//  4. serve files of the user's workspace via HTTP
//
// Note: user [playground] is a reserved mock user
func addUser(userId, userName, userAvatar string) string {
	if "playground" == userId {
		return userExists
	}

	addUserMutex.Lock()
	defer addUserMutex.Unlock()

	for _, user := range conf.Users {
		if strings.ToLower(user.Id) == strings.ToLower(userId) {
			return userExists
		}
	}

	workspace := filepath.Join(conf.Wide.Data, "workspaces", userId)
	newUser := conf.NewUser(userId, userName, userAvatar, workspace)
	conf.Users = append(conf.Users, newUser)
	if !newUser.Save() {
		return userCreateError
	}

	conf.CreateWorkspaceDir(workspace)
	helloWorld(workspace)
	conf.UpdateCustomizedConf(userId)

	logger.Infof("Created a user [%s]", userId)

	return userCreated
}

// helloWorld generates the 'Hello, 世界' source code.
func helloWorld(workspace string) {
	dir := workspace + conf.PathSeparator + "src" + conf.PathSeparator + "hello"
	if err := os.MkdirAll(dir, 0755); nil != err {
		logger.Error(err)

		return
	}

	fout, err := os.Create(dir + conf.PathSeparator + "main.go")
	if nil != err {
		logger.Error(err)

		return
	}

	fout.WriteString(conf.HelloWorld)
	fout.Close()
}
