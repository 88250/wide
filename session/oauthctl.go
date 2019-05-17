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
	"crypto/tls"
	"github.com/b3log/wide/i18n"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/util"
	"github.com/parnurzeal/gorequest"
)

var states = map[string]string{}

// RedirectGitHubHandler redirects to GitHub auth page.
func RedirectGitHubHandler(w http.ResponseWriter, r *http.Request) {
	requestResult := util.NewResult()
	_, _, errs := gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		Get("https://hacpai.com/oauth/wide/client").
		Set("user-agent", conf.UserAgent).Timeout(10 * time.Second).EndStruct(requestResult)
	if nil != errs {
		logger.Errorf("Get oauth client id failed: %+v", errs)
		http.Error(w, "Get oauth info failed", http.StatusInternalServerError)

		return
	}
	if 0 != requestResult.Code {
		logger.Errorf("get oauth client id failed [code=%d, msg=%s]", requestResult.Code, requestResult.Msg)
		http.Error(w, "Get oauth info failed", http.StatusNotFound)

		return
	}
	data := requestResult.Data.(map[string]interface{})
	clientId := data["clientId"].(string)
	loginAuthURL := data["loginAuthURL"].(string)

	referer := r.URL.Query().Get("referer")
	if "" == referer || !strings.Contains(referer, "://") {
		referer = conf.Wide.Server + referer
	}
	if strings.HasSuffix(referer, "/") {
		referer = referer[:len(referer)-1]
	}
	referer += "__1"
	state := util.Rand.String(16) + referer
	states[state] = state
	path := loginAuthURL + "?client_id=" + clientId + "&state=" + state + "&scope=public_repo,read:user,user:follow"
	http.Redirect(w, r, path, http.StatusSeeOther)
}

func GithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if _, exist := states[state]; !exist {
		http.Error(w, "Get state param failed", http.StatusBadRequest)

		return
	}
	delete(states, state)

	referer := state[16:]
	if strings.Contains(referer, "__0") || strings.Contains(referer, "__1") {
		referer = referer[:len(referer)-len("__0")]
	}
	accessToken := r.URL.Query().Get("ak")
	githubUser := GitHubUserInfo(accessToken)
	if nil == githubUser {
		logger.Warnf("Can not get user info with token [" + accessToken + "]")
		http.Error(w, "Get user info failed", http.StatusUnauthorized)

		return
	}

	githubId := githubUser["userId"].(string)
	userName := githubUser["userName"].(string)
	avatar := githubUser["userAvatarURL"].(string)

	result := util.NewResult()
	defer util.RetResult(w, r, result)

	user := conf.GetUser(githubId)
	if nil == user {
		msg := addUser(githubId, userName, avatar)
		if userCreated != msg {
			result.Succ = false
			result.Msg = msg

			return
		}
	}

	// create a HTTP session
	httpSession, _ := HTTPSession.Get(r, CookieName)
	httpSession.Values["uid"] = githubId
	httpSession.Values["id"] = strconv.Itoa(rand.Int())
	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)
}

// GitHubUserInfo returns GitHub user info specified by the given access token.
func GitHubUserInfo(accessToken string) (ret map[string]interface{}) {
	result := map[string]interface{}{}
	response, data, errors := gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		Get("https://hacpai.com/github/user?ak="+accessToken).Timeout(7*time.Second).
		Set("User-Agent", conf.UserAgent).EndStruct(&result)
	if nil != errors || http.StatusOK != response.StatusCode {
		logger.Errorf("Get github user info failed: %+v, %s", errors, data)

		return nil
	}

	if 0 != result["sc"].(float64) {
		return nil
	}

	return result["data"].(map[string]interface{})
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
	result := util.NewResult()
	defer util.RetResult(w, r, result)

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
