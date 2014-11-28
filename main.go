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

package main

import (
	"flag"
	"html/template"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/editor"
	"github.com/b3log/wide/event"
	"github.com/b3log/wide/file"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/notification"
	"github.com/b3log/wide/output"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/shell"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

// The only one init function in Wide.
func init() {
	confPath := flag.String("conf", "conf/wide.json", "path of wide.json")
	confIP := flag.String("ip", "", "ip to visit")
	confPort := flag.String("port", "", "port to visit")
	confServer := flag.String("server", "", "this will overwrite Wide.Server if specified")
	confChannel := flag.String("channel", "", "this will overwrite Wide.XXXChannel if specified")
	confStat := flag.Bool("stat", false, "whether report statistics periodically")
	confDocker := flag.Bool("docker", false, "whether run in a docker container")

	flag.Set("alsologtostderr", "true")
	flag.Set("stderrthreshold", "INFO")
	flag.Set("v", "3")

	flag.Parse()

	wd := util.OS.Pwd()
	if strings.HasPrefix(wd, os.TempDir()) {
		glog.Error("Don't run wide in OS' temp directory or with `go run`")

		os.Exit(-1)
	}

	i18n.Load()

	event.Load()

	conf.Load(*confPath, *confIP, *confPort, *confServer, *confChannel, *confDocker)

	conf.FixedTimeCheckEnv()
	conf.FixedTimeSave()

	session.FixedTimeRelease()

	if *confStat {
		session.FixedTimeReport()
	}
}

// indexHandler handles request of Wide index.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	// create a Wide session
	rand.Seed(time.Now().UnixNano())
	sid := strconv.Itoa(rand.Int())
	wideSession := session.WideSessions.New(httpSession, sid)

	username := httpSession.Values["username"].(string)
	user := conf.Wide.GetUser(username)
	if nil == user {
		glog.Warningf("Not found user [%s]", username)

		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	locale := user.Locale

	wideSessions := session.WideSessions.GetByUsername(username)

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale,
		"session": wideSession, "latestSessionContent": user.LatestSessionContent,
		"pathSeparator": conf.PathSeparator, "codeMirrorVer": conf.CodeMirrorVer}

	glog.V(3).Infof("User [%s] has [%d] sessions", username, len(wideSessions))

	t, err := template.ParseFiles("views/index.html")

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, model)
}

// serveSingle registers the handler function for the given pattern and filename.
func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	})
}

// startHandler handles request of start page.
func startHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	username := httpSession.Values["username"].(string)
	locale := conf.Wide.GetUser(username).Locale
	userWorkspace := conf.Wide.GetUserWorkspace(username)

	sid := r.URL.Query()["sid"][0]
	wSession := session.WideSessions.Get(sid)
	if nil == wSession {
		glog.Errorf("Session [%s] not found", sid)
	}

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale,
		"username": username, "workspace": userWorkspace, "ver": conf.WideVersion, "session": wSession}

	t, err := template.ParseFiles("views/start.html")

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, model)
}

// keyboardShortcutsHandler handles request of keyboard shortcuts page.
func keyboardShortcutsHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	username := httpSession.Values["username"].(string)
	locale := conf.Wide.GetUser(username).Locale

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale}

	t, err := template.ParseFiles("views/keyboard_shortcuts.html")

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, model)
}

// aboutHandle handles request of about page.
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	username := httpSession.Values["username"].(string)
	locale := conf.Wide.GetUser(username).Locale

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale,
		"ver": conf.WideVersion, "goos": runtime.GOOS, "goarch": runtime.GOARCH, "gover": runtime.Version()}

	t, err := template.ParseFiles("views/about.html")

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, model)
}

// Main.
func main() {
	runtime.GOMAXPROCS(conf.Wide.MaxProcs)

	initMime()

	defer glog.Flush()

	// IDE
	http.HandleFunc("/", handlerWrapper(indexHandler))
	http.HandleFunc("/start", handlerWrapper(startHandler))
	http.HandleFunc("/about", handlerWrapper(aboutHandler))
	http.HandleFunc("/keyboard_shortcuts", handlerWrapper(keyboardShortcutsHandler))

	// static resources
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	serveSingle("/favicon.ico", "./static/favicon.ico")

	// workspaces
	for _, user := range conf.Wide.Users {
		http.Handle("/workspace/"+user.Name+"/",
			http.StripPrefix("/workspace/"+user.Name+"/", http.FileServer(http.Dir(user.GetWorkspace()))))
	}

	// session
	http.HandleFunc("/session/ws", handlerWrapper(session.WSHandler))
	http.HandleFunc("/session/save", handlerWrapper(session.SaveContent))

	// run
	http.HandleFunc("/build", handlerWrapper(output.BuildHandler))
	http.HandleFunc("/run", handlerWrapper(output.RunHandler))
	http.HandleFunc("/stop", handlerWrapper(output.StopHandler))
	http.HandleFunc("/go/test", handlerWrapper(output.GoTestHandler))
	http.HandleFunc("/go/get", handlerWrapper(output.GoGetHandler))
	http.HandleFunc("/go/install", handlerWrapper(output.GoInstallHandler))
	http.HandleFunc("/output/ws", handlerWrapper(output.WSHandler))

	// file tree
	http.HandleFunc("/files", handlerWrapper(file.GetFiles))
	http.HandleFunc("/file/refresh", handlerWrapper(file.RefreshDirectory))
	http.HandleFunc("/file", handlerWrapper(file.GetFile))
	http.HandleFunc("/file/save", handlerWrapper(file.SaveFile))
	http.HandleFunc("/file/new", handlerWrapper(file.NewFile))
	http.HandleFunc("/file/remove", handlerWrapper(file.RemoveFile))
	http.HandleFunc("/file/rename", handlerWrapper(file.RenameFile))
	http.HandleFunc("/file/search/text", handlerWrapper(file.SearchText))
	http.HandleFunc("/file/find/name", handlerWrapper(file.Find))

	// file export/import
	http.HandleFunc("/file/zip/new", handlerWrapper(file.CreateZip))
	http.HandleFunc("/file/zip", handlerWrapper(file.GetZip))
	http.HandleFunc("/file/upload", handlerWrapper(file.Upload))

	// editor
	http.HandleFunc("/editor/ws", handlerWrapper(editor.WSHandler))
	http.HandleFunc("/go/fmt", handlerWrapper(editor.GoFmtHandler))
	http.HandleFunc("/autocomplete", handlerWrapper(editor.AutocompleteHandler))
	http.HandleFunc("/exprinfo", handlerWrapper(editor.GetExprInfoHandler))
	http.HandleFunc("/find/decl", handlerWrapper(editor.FindDeclarationHandler))
	http.HandleFunc("/find/usages", handlerWrapper(editor.FindUsagesHandler))

	// shell
	http.HandleFunc("/shell/ws", handlerWrapper(shell.WSHandler))
	http.HandleFunc("/shell", handlerWrapper(shell.IndexHandler))

	// notification
	http.HandleFunc("/notification/ws", handlerWrapper(notification.WSHandler))

	// user
	http.HandleFunc("/login", handlerWrapper(session.LoginHandler))
	http.HandleFunc("/logout", handlerWrapper(session.LogoutHandler))
	http.HandleFunc("/signup", handlerWrapper(session.SignUpUser))
	http.HandleFunc("/preference", handlerWrapper(session.PreferenceHandler))

	glog.Infof("Wide is running [%s]", conf.Wide.Server)

	err := http.ListenAndServe(conf.Wide.Server, nil)
	if err != nil {
		glog.Fatal(err)
	}
}

// handlerWrapper wraps the HTTP Handler for some common processes.
//
//  1. panic recover
//  2. request stopwatch
//  3. i18n
func handlerWrapper(f func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	handler := panicRecover(f)
	handler = stopwatch(handler)
	handler = i18nLoad(handler)

	return handler
}

// i18nLoad wraps the i18n process.
func i18nLoad(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		i18n.Load()

		handler(w, r)
	}
}

// stopwatch wraps the request stopwatch process.
func stopwatch(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		defer func() {
			glog.V(5).Infof("[%s] [%s]", r.RequestURI, time.Since(start))
		}()

		handler(w, r)
	}
}

// panicRecover wraps the panic recover process.
func panicRecover(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer util.Recover()

		handler(w, r)
	}
}

// initMime initializes mime types.
//
// We can't get the mime types on some OS (such as Windows XP) by default, so initializes them here.
func initMime() {
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/x-javascript")
	mime.AddExtensionType(".json", "application/json")
}
