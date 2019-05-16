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
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/util"
)

const (
	// TODO: i18n

	userExists       = "user exists"
	userCreated      = "user created"
	userCreateError  = "user create error"
	notAllowRegister = "not allow register"
)

// Exclusive lock for adding user.
var addUserMutex sync.Mutex

// PreferenceHandler handles request of preference page.
func PreferenceHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := HTTPSession.Get(r, "wide-session")

	if httpSession.IsNew {
		http.Redirect(w, r, conf.Wide.Context+"/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	if "" != conf.Wide.Context {
		httpSession.Options.Path = conf.Wide.Context
	}
	httpSession.Save(r, w)

	uid := httpSession.Values["uid"].(string)
	user := conf.GetUser(uid)

	if "GET" == r.Method {
		tmpLinux := user.GoBuildArgsForLinux
		tmpWindows := user.GoBuildArgsForWindows
		tmpDarwin := user.GoBuildArgsForDarwin

		user.GoBuildArgsForLinux = strings.Replace(user.GoBuildArgsForLinux, `"`, `&quot;`, -1)
		user.GoBuildArgsForWindows = strings.Replace(user.GoBuildArgsForWindows, `"`, `&quot;`, -1)
		user.GoBuildArgsForDarwin = strings.Replace(user.GoBuildArgsForDarwin, `"`, `&quot;`, -1)

		model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(user.Locale), "user": user,
			"ver": conf.WideVersion, "goos": runtime.GOOS, "goarch": runtime.GOARCH, "gover": runtime.Version(),
			"locales": i18n.GetLocalesNames(), "gofmts": util.Go.GetGoFormats(),
			"themes": conf.GetThemes(), "editorThemes": conf.GetEditorThemes()}

		t, err := template.ParseFiles("views/preference.html")

		if nil != err {
			logger.Error(err)
			http.Error(w, err.Error(), 500)

			user.GoBuildArgsForLinux = tmpLinux
			user.GoBuildArgsForWindows = tmpWindows
			user.GoBuildArgsForDarwin = tmpDarwin
			return
		}

		t.Execute(w, model)

		user.GoBuildArgsForLinux = tmpLinux
		user.GoBuildArgsForWindows = tmpWindows
		user.GoBuildArgsForDarwin = tmpDarwin

		return
	}

	// non-GET request as save request

	result := util.NewResult()
	defer util.RetResult(w, r, result)

	args := struct {
		FontFamily            string
		FontSize              string
		GoFmt                 string
		GoBuildArgsForLinux   string
		GoBuildArgsForWindows string
		GoBuildArgsForDarwin  string
		Keymap                string
		Workspace             string
		Username              string
		Locale                string
		Theme                 string
		EditorFontFamily      string
		EditorFontSize        string
		EditorLineHeight      string
		EditorTheme           string
		EditorTabSize         string
	}{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	user.FontFamily = args.FontFamily
	user.FontSize = args.FontSize
	user.GoFormat = args.GoFmt
	user.GoBuildArgsForLinux = args.GoBuildArgsForLinux
	user.GoBuildArgsForWindows = args.GoBuildArgsForWindows
	user.GoBuildArgsForDarwin = args.GoBuildArgsForDarwin
	user.Keymap = args.Keymap
	// XXX: disallow change workspace at present
	// user.Workspace = args.Workspace

	user.Locale = args.Locale
	user.Theme = args.Theme
	user.Editor.FontFamily = args.EditorFontFamily
	user.Editor.FontSize = args.EditorFontSize
	user.Editor.LineHeight = args.EditorLineHeight
	user.Editor.Theme = args.EditorTheme
	user.Editor.TabSize = args.EditorTabSize

	conf.UpdateCustomizedConf(uid)

	now := time.Now().UnixNano()
	user.Lived = now
	user.Updated = now

	result.Succ = user.Save()
}

// LogoutHandler handles request of user logout (exit).
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	httpSession, _ := HTTPSession.Get(r, "wide-session")

	httpSession.Options.MaxAge = -1
	httpSession.Save(r, w)
}

// FixedTimeSave saves online users' configurations periodically (1 minute).
//
// Main goal of this function is to save user session content, for restoring session content while user open Wide next time.
func FixedTimeSave() {
	go func() {
		defer util.Recover()

		for _ = range time.Tick(time.Minute) {
			SaveOnlineUsers()
		}
	}()
}

// CanAccess determines whether the user specified by the given user id can access the specified path.
func CanAccess(userId, path string) bool {
	path = filepath.FromSlash(path)

	userWorkspace := conf.GetUserWorkspace(userId)
	workspaces := filepath.SplitList(userWorkspace)

	for _, workspace := range workspaces {
		if strings.HasPrefix(path, workspace) {
			return true
		}
	}

	return false
}

// SaveOnlineUsers saves online users' configurations at once.
func SaveOnlineUsers() {
	users := getOnlineUsers()
	for _, u := range users {
		if u.Save() {
			logger.Tracef("Saved online user [%s]'s configurations", u.Name)
		}
	}
}

func getOnlineUsers() []*conf.User {
	ret := []*conf.User{}

	uids := map[string]string{} // distinct uid
	for _, s := range WideSessions {
		uids[s.UserId] = s.UserId
	}

	for _, uid := range uids {
		u := conf.GetUser(uid)

		if "playground" == uid { // user [playground] is a reserved mock user
			continue
		}

		if nil == u {
			logger.Warnf("Not found user [%s]", uid)

			continue
		}

		ret = append(ret, u)
	}

	return ret
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
	if !conf.Wide.AllowRegister {
		return notAllowRegister
	}

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

	workspace := filepath.Join(conf.Wide.UsersWorkspaces, userId)
	newUser := conf.NewUser(userId, userName, userAvatar, workspace)
	conf.Users = append(conf.Users, newUser)
	if !newUser.Save() {
		return userCreateError
	}

	conf.CreateWorkspaceDir(workspace)
	helloWorld(workspace)
	conf.UpdateCustomizedConf(userId)

	http.Handle("/workspace/"+userId+"/",
		http.StripPrefix("/workspace/"+userId+"/", http.FileServer(http.Dir(newUser.WorkspacePath()))))

	logger.Infof("Created a user [%s]", userId)

	return userCreated
}

// helloWorld generates the 'Hello, 世界' source code.
//  1. src/hello/main.go
//  2. src/web/main.go
func helloWorld(workspace string) {
	consoleHello(workspace)
	webHello(workspace)
}

func consoleHello(workspace string) {
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

func webHello(workspace string) {
	dir := workspace + conf.PathSeparator + "src" + conf.PathSeparator + "web"
	if err := os.MkdirAll(dir, 0755); nil != err {
		logger.Error(err)

		return
	}

	fout, err := os.Create(dir + conf.PathSeparator + "main.go")
	if nil != err {
		logger.Error(err)

		return
	}

	code := `package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, 世界"))
	})

	port := getPort()

	// you may need to change the address
	fmt.Println("Open https://wide.b3log.org:" + port + " in your browser to see the result") 

	if err := http.ListenAndServe(":"+port, nil); nil != err {
		fmt.Println(err)
	}
}

func getPort() string {
	rand.Seed(time.Now().UnixNano())

	return strconv.Itoa(7000 + rand.Intn(8000-7000))
}

`

	fout.WriteString(code)

	fout.Close()
}
