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

package main

import (
	"compress/gzip"
	"flag"
	"html/template"
	"io"
	"mime"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/b3log/gulu"
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/editor"
	"github.com/b3log/wide/event"
	"github.com/b3log/wide/file"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/notification"
	"github.com/b3log/wide/output"
	"github.com/b3log/wide/playground"
	"github.com/b3log/wide/session"
)

// Logger
var logger *gulu.Logger

// The only one init function in Wide.
func init() {
	confPath := flag.String("conf", "conf/wide.json", "path of wide.json")
	confData := flag.String("data", "", "path of data dir")
	confServer := flag.String("server", "", "this will overwrite Wide.Server if specified")
	confLogLevel := flag.String("log_level", "", "this will overwrite Wide.LogLevel if specified")
	confSiteStatCode := flag.String("site_stat_code", "", "this will overrite Wide.SiteStatCode if specified")

	flag.Parse()

	gulu.Log.SetLevel("warn")
	logger = gulu.Log.NewLogger(os.Stdout)

	//wd := gulu.OS.Pwd()
	//if strings.HasPrefix(wd, os.TempDir()) {
	//	logger.Error("Don't run Wide in OS' temp directory or with `go run`")
	//
	//	os.Exit(-1)
	//}

	i18n.Load()
	event.Load()
	conf.Load(*confPath, *confData, *confServer, *confLogLevel, template.HTML(*confSiteStatCode))

	conf.FixedTimeCheckEnv()
	session.FixedTimeSave()
	session.FixedTimeRelease()
	session.FixedTimeReport()

	logger.Debug("host ["+runtime.Version()+", "+runtime.GOOS+"_"+runtime.GOARCH+"], cross-compilation ", gulu.Go.GetCrossPlatforms())
}

// Main.
func main() {
	initMime()
	handleSignal()

	// IDE
	http.HandleFunc("/", handlerGzWrapper(indexHandler))
	http.HandleFunc("/start", handlerWrapper(startHandler))
	http.HandleFunc("/about", handlerWrapper(aboutHandler))
	http.HandleFunc("/keyboard_shortcuts", handlerWrapper(keyboardShortcutsHandler))

	// static resources
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/static/users/", http.StripPrefix("/static/", http.FileServer(http.Dir(conf.Wide.Data+"/static"))))
	serveSingle("/favicon.ico", "./static/images/favicon.png")

	// oauth
	http.HandleFunc("/oauth/github/redirect", session.RedirectGitHubHandler)
	http.HandleFunc("/oauth/github/callback", session.GithubCallbackHandler)

	// session
	http.HandleFunc("/session/ws", handlerWrapper(session.WSHandler))
	http.HandleFunc("/session/save", handlerWrapper(session.SaveContentHandler))

	// run
	http.HandleFunc("/build", handlerWrapper(output.BuildHandler))
	http.HandleFunc("/run", handlerWrapper(output.RunHandler))
	http.HandleFunc("/stop", handlerWrapper(output.StopHandler))
	http.HandleFunc("/go/test", handlerWrapper(output.GoTestHandler))
	http.HandleFunc("/go/vet", handlerWrapper(output.GoVetHandler))
	http.HandleFunc("/go/install", handlerWrapper(output.GoInstallHandler))
	http.HandleFunc("/output/ws", handlerWrapper(output.WSHandler))

	// cross-compilation
	http.HandleFunc("/cross", handlerWrapper(output.CrossCompilationHandler))

	// file tree
	http.HandleFunc("/files", handlerWrapper(file.GetFilesHandler))
	http.HandleFunc("/file/refresh", handlerWrapper(file.RefreshDirectoryHandler))
	http.HandleFunc("/file", handlerWrapper(file.GetFileHandler))
	http.HandleFunc("/file/save", handlerWrapper(file.SaveFileHandler))
	http.HandleFunc("/file/new", handlerWrapper(file.NewFileHandler))
	http.HandleFunc("/file/remove", handlerWrapper(file.RemoveFileHandler))
	http.HandleFunc("/file/rename", handlerWrapper(file.RenameFileHandler))
	http.HandleFunc("/file/search/text", handlerWrapper(file.SearchTextHandler))
	http.HandleFunc("/file/find/name", handlerWrapper(file.FindHandler))

	// outline
	http.HandleFunc("/outline", handlerWrapper(file.GetOutlineHandler))

	// file export
	http.HandleFunc("/file/zip/new", handlerWrapper(file.CreateZipHandler))
	http.HandleFunc("/file/zip", handlerWrapper(file.GetZipHandler))

	// editor
	http.HandleFunc("/editor/ws", handlerWrapper(editor.WSHandler))
	http.HandleFunc("/go/fmt", handlerWrapper(editor.GoFmtHandler))
	http.HandleFunc("/autocomplete", handlerWrapper(editor.AutocompleteHandler))
	http.HandleFunc("/exprinfo", handlerWrapper(editor.GetExprInfoHandler))
	http.HandleFunc("/find/decl", handlerWrapper(editor.FindDeclarationHandler))
	http.HandleFunc("/find/usages", handlerWrapper(editor.FindUsagesHandler))

	// notification
	http.HandleFunc("/notification/ws", handlerWrapper(notification.WSHandler))

	// user
	http.HandleFunc("/login", handlerWrapper(session.LoginHandler))
	http.HandleFunc("/logout", handlerWrapper(session.LogoutHandler))
	http.HandleFunc("/preference", handlerWrapper(session.PreferenceHandler))

	// playground
	http.HandleFunc("/playground", handlerWrapper(playground.IndexHandler))
	http.HandleFunc("/playground/", handlerWrapper(playground.IndexHandler))
	http.HandleFunc("/playground/ws", handlerWrapper(playground.WSHandler))
	http.HandleFunc("/playground/save", handlerWrapper(playground.SaveHandler))
	http.HandleFunc("/playground/build", handlerWrapper(playground.BuildHandler))
	http.HandleFunc("/playground/run", handlerWrapper(playground.RunHandler))
	http.HandleFunc("/playground/stop", handlerWrapper(playground.StopHandler))
	http.HandleFunc("/playground/autocomplete", handlerWrapper(playground.AutocompleteHandler))

	logger.Infof("Wide is running [%s]", conf.Wide.Server)

	err := http.ListenAndServe("127.0.0.1:7070", nil)
	if err != nil {
		logger.Error(err)
	}
}

// indexHandler handles request of Wide index.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if "/" != r.RequestURI {
		http.Redirect(w, r, "/", http.StatusFound)

		return
	}

	httpSession, _ := session.HTTPSession.Get(r, session.CookieName)
	if httpSession.IsNew {
		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	uid := httpSession.Values["uid"].(string)
	if "playground" == uid { // reserved user for Playground
		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	user := conf.GetUser(uid)
	if nil == user {
		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	locale := user.Locale

	wideSessions := session.WideSessions.GetByUserId(uid)

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale,
		"uid": uid, "sid": session.WideSessions.GenId(), "latestSessionContent": user.LatestSessionContent,
		"pathSeparator": conf.PathSeparator, "codeMirrorVer": conf.CodeMirrorVer,
		"user": user, "editorThemes": conf.GetEditorThemes(), "crossPlatforms": gulu.Go.GetCrossPlatforms()}

	logger.Debugf("User [%s] has [%d] sessions", uid, len(wideSessions))

	t, err := template.ParseFiles("views/index.html")
	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	t.Execute(w, model)
}

// handleSignal handles system signal for graceful shutdown.
func handleSignal() {
	go func() {
		c := make(chan os.Signal)

		signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
		s := <-c
		logger.Tracef("Got signal [%s]", s)

		session.SaveOnlineUsers()
		logger.Tracef("Saved all online user, exit")

		os.Exit(0)
	}()
}

// serveSingle registers the handler function for the given pattern and filename.
func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	})
}

// startHandler handles request of start page.
func startHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, session.CookieName)
	if httpSession.IsNew {
		http.Redirect(w, r, "/s", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	uid := httpSession.Values["uid"].(string)
	locale := conf.GetUser(uid).Locale
	userWorkspace := conf.GetUserWorkspace(uid)

	sid := r.URL.Query()["sid"][0]
	wSession := session.WideSessions.Get(sid)
	if nil == wSession {
		logger.Errorf("Session [%s] not found", sid)
	}

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale,
		"uid": uid, "workspace": userWorkspace, "ver": conf.WideVersion, "sid": sid}

	t, err := template.ParseFiles("views/start.html")

	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	t.Execute(w, model)
}

// keyboardShortcutsHandler handles request of keyboard shortcuts page.
func keyboardShortcutsHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, session.CookieName)
	if httpSession.IsNew {
		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	uid := httpSession.Values["uid"].(string)
	locale := conf.GetUser(uid).Locale

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale}

	t, err := template.ParseFiles("views/keyboard_shortcuts.html")

	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	t.Execute(w, model)
}

// aboutHandle handles request of about page.
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, session.CookieName)
	if httpSession.IsNew {
		http.Redirect(w, r, "/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	httpSession.Save(r, w)

	uid := httpSession.Values["uid"].(string)
	locale := conf.GetUser(uid).Locale

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale,
		"ver": conf.WideVersion, "goos": runtime.GOOS, "goarch": runtime.GOARCH, "gover": runtime.Version()}

	t, err := template.ParseFiles("views/about.html")

	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	t.Execute(w, model)
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

// handlerGzWrapper wraps the HTTP Handler for some common processes.
//
//  1. panic recover
//  2. gzip response
//  3. request stopwatch
//  4. i18n
func handlerGzWrapper(f func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	handler := panicRecover(f)
	handler = gzipWrapper(handler)
	handler = stopwatch(handler)
	handler = i18nLoad(handler)

	return handler
}

// gzipWrapper wraps the process with response gzip.
func gzipWrapper(f func(http.ResponseWriter, *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			f(w, r)

			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}

		f(gzr, r)
	}
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
			logger.Tracef("[%s, %s, %s]", r.Method, r.RequestURI, time.Since(start))
		}()

		handler(w, r)
	}
}

// panicRecover wraps the panic recover process.
func panicRecover(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer gulu.Panic.Recover()

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

// gzipResponseWriter represents a gzip response writer.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// Write writes response with appropriate 'Content-Type'.
func (w gzipResponseWriter) Write(b []byte) (int, error) {
	if "" == w.Header().Get("Content-Type") {
		// If no content type, apply sniffing algorithm to un-gzipped body.
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}

	return w.Writer.Write(b)
}
