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

package session

import (
	"crypto/tls"
	"github.com/b3log/wide/conf"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	state := util.Rand.String(16) + referer
	states[state] = state
	path := loginAuthURL + "?client_id=" + clientId + "&state=" + state + "&scope=public_repo,read:user,user:follow"

	logger.Infof("redirect to github [" + path + "]")

	http.Redirect(w, r, path, http.StatusSeeOther)
}

func GithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	logger.Infof("Github callback [" + r.URL.String() + "]")

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
	avatar := githubUser["userAvatar"].(string)

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
	httpSession, _ := HTTPSession.Get(r, "wide-session")
	httpSession.Values["username"] = userName

	httpSession.Values["id"] = strconv.Itoa(rand.Int())
	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	if "" != conf.Wide.Context {
		httpSession.Options.Path = conf.Wide.Context
	}
	httpSession.Save(r, w)

	logger.Debugf("Created a HTTP session [%s] for user [%s]", httpSession.Values["id"].(string), userName)
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
