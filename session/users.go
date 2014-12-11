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

package session

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"text/template"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

const (
	// TODO: i18n

	userExists      = "user exists"
	emailExists     = "email exists"
	userCreated     = "user created"
	userCreateError = "user create error"
)

// Exclusive lock for adding user.
var addUserMutex sync.Mutex

// PreferenceHandler handles request of preference page.
func PreferenceHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := HTTPSession.Get(r, "wide-session")

	if httpSession.IsNew {
		http.Redirect(w, r, conf.Wide.Context+"login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	username := httpSession.Values["username"].(string)
	user := conf.Wide.GetUser(username)

	if "GET" == r.Method {
		model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(user.Locale), "user": user,
			"ver": conf.WideVersion, "goos": runtime.GOOS, "goarch": runtime.GOARCH, "gover": runtime.Version(),
			"locales": i18n.GetLocalesNames(), "gofmts": util.Go.GetGoFormats(),
			"themes": conf.GetThemes(), "editorThemes": conf.GetEditorThemes()}

		t, err := template.ParseFiles("views/preference.html")

		if nil != err {
			glog.Error(err)
			http.Error(w, err.Error(), 500)

			return
		}

		t.Execute(w, model)

		return
	}

	// non-GET request as save request

	succ := true
	data := map[string]interface{}{"succ": &succ}
	defer util.RetJSON(w, r, data)

	args := struct {
		FontFamily       string
		FontSize         string
		GoFmt            string
		Workspace        string
		Username         string
		Password         string
		Email            string
		Locale           string
		Theme            string
		EditorFontFamily string
		EditorFontSize   string
		EditorLineHeight string
		EditorTheme      string
		EditorTabSize    string
	}{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		succ = false

		return
	}

	user.FontFamily = args.FontFamily
	user.FontSize = args.FontSize
	user.GoFormat = args.GoFmt
	user.Workspace = args.Workspace
	user.Password = args.Password
	user.Email = args.Email
	user.Locale = args.Locale
	user.Theme = args.Theme
	user.Editor.FontFamily = args.EditorFontFamily
	user.Editor.FontSize = args.EditorFontSize
	user.Editor.LineHeight = args.EditorLineHeight
	user.Editor.Theme = args.EditorTheme
	user.Editor.TabSize = args.EditorTabSize

	conf.UpdateCustomizedConf(username)

	succ = conf.Save()
}

// LoginHandler handles request of user login.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if "GET" == r.Method {
		// show the login page

		model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(conf.Wide.Locale),
			"locale": conf.Wide.Locale, "ver": conf.WideVersion}

		t, err := template.ParseFiles("views/login.html")

		if nil != err {
			glog.Error(err)
			http.Error(w, err.Error(), 500)

			return
		}

		t.Execute(w, model)

		return
	}

	// non-GET request as login request

	succ := true
	data := map[string]interface{}{"succ": &succ}
	defer util.RetJSON(w, r, data)

	args := struct {
		Username string
		Password string
	}{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		succ = false

		return
	}

	succ = false
	for _, user := range conf.Wide.Users {
		if user.Name == args.Username && user.Password == args.Password {
			succ = true

			break
		}
	}

	if !succ {
		return
	}

	// create a HTTP session
	httpSession, _ := HTTPSession.Get(r, "wide-session")
	httpSession.Values["username"] = args.Username
	httpSession.Values["id"] = strconv.Itoa(rand.Int())
	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	glog.Infof("Created a HTTP session [%s] for user [%s]", httpSession.Values["id"].(string), args.Username)
}

// LogoutHandler handles request of user logout (exit).
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	httpSession, _ := HTTPSession.Get(r, "wide-session")

	httpSession.Options.MaxAge = -1
	httpSession.Save(r, w)
}

// SignUpUser handles request of registering user.
func SignUpUser(w http.ResponseWriter, r *http.Request) {
	if "GET" == r.Method {
		// show the user sign up page

		firstUserWorkspace := conf.Wide.GetUserWorkspace(conf.Wide.Users[0].Name)
		dir := filepath.Dir(firstUserWorkspace)

		model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(conf.Wide.Locale),
			"locale": conf.Wide.Locale, "ver": conf.WideVersion, "dir": dir,
			"pathSeparator": conf.PathSeparator}

		t, err := template.ParseFiles("views/sign_up.html")

		if nil != err {
			glog.Error(err)
			http.Error(w, err.Error(), 500)

			return
		}

		t.Execute(w, model)

		return
	}

	// non-GET request as add user request

	succ := true
	data := map[string]interface{}{"succ": &succ}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		succ = false

		return
	}

	username := args["username"].(string)
	password := args["password"].(string)
	email := args["email"].(string)

	msg := addUser(username, password, email)
	if userCreated != msg {
		succ = false
		data["msg"] = msg
	}
}

// addUser add a user with the specified username, password and email.
//
//  1. create the user's workspace
//  2. generate 'Hello, 世界' demo code in the workspace
//  3. update the user customized configurations, such as style.css
//  4. serve files of the user's workspace via HTTP
func addUser(username, password, email string) string {
	addUserMutex.Lock()
	defer addUserMutex.Unlock()

	for _, user := range conf.Wide.Users {
		if user.Name == username {
			return userExists
		}

		if user.Email == email {
			return emailExists
		}
	}

	firstUserWorkspace := conf.Wide.GetUserWorkspace(conf.Wide.Users[0].Name)
	dir := filepath.Dir(firstUserWorkspace)
	workspace := filepath.Join(dir, username)

	newUser := conf.NewUser(username, password, email, workspace)
	conf.Wide.Users = append(conf.Wide.Users, newUser)

	if !conf.Save() {
		return userCreateError
	}

	conf.CreateWorkspaceDir(workspace)
	helloWorld(workspace)
	conf.UpdateCustomizedConf(username)

	http.Handle("/workspace/"+username+"/",
		http.StripPrefix("/workspace/"+username+"/", http.FileServer(http.Dir(newUser.GetWorkspace()))))

	glog.Infof("Created a user [%s]", username)

	return userCreated
}

// helloWorld generates the 'Hello, 世界' source code in workspace/src/hello/main.go.
func helloWorld(workspace string) {
	dir := workspace + conf.PathSeparator + "src" + conf.PathSeparator + "hello"
	if err := os.MkdirAll(dir, 0755); nil != err {
		glog.Error(err)

		return
	}

	fout, err := os.Create(dir + conf.PathSeparator + "main.go")
	if nil != err {
		glog.Error(err)

		os.Exit(-1)
	}

	fout.WriteString(`package main

import "fmt"

func main() {
	fmt.Println("Hello, 世界")
}
`)

	fout.Close()
}
