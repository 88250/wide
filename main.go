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
	"runtime"
	"strings"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/editor"
	"github.com/b3log/wide/event"
	"github.com/b3log/wide/file"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/log"
	"github.com/b3log/wide/notification"
	"github.com/b3log/wide/output"
	"github.com/b3log/wide/playground"
	"github.com/b3log/wide/scm/git"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
)

// Logger
var logger *log.Logger

// The only one init function in Wide.
func init() {
	confPath := flag.String("conf", "conf/wide.json", "path of wide.json")
	confIP := flag.String("ip", "", "this will overwrite Wide.IP if specified")
	confPort := flag.String("port", "", "this will overwrite Wide.Port if specified")
	confServer := flag.String("server", "", "this will overwrite Wide.Server if specified")
	confLogLevel := flag.String("log_level", "", "this will overwrite Wide.LogLevel if specified")
	confStaticServer := flag.String("static_server", "", "this will overwrite Wide.StaticServer if specified")
	confContext := flag.String("context", "", "this will overwrite Wide.Context if specified")
	confChannel := flag.String("channel", "", "this will overwrite Wide.Channel if specified")
	confStat := flag.Bool("stat", false, "whether report statistics periodically")
	confDocker := flag.Bool("docker", false, "whether run in a docker container")
	confPlayground := flag.String("playground", "", "this will overwrite Wide.Playground if specified")

	flag.Parse()

	log.SetLevel("warn")
	logger = log.NewLogger(os.Stdout)

	wd := util.OS.Pwd()
	if strings.HasPrefix(wd, os.TempDir()) {
		logger.Error("Don't run Wide in OS' temp directory or with `go run`")

		os.Exit(-1)
	}

	i18n.Load()

	event.Load()

	conf.Load(*confPath, *confIP, *confPort, *confServer, *confLogLevel, *confStaticServer, *confContext, *confChannel,
		*confPlayground, *confDocker)

	conf.FixedTimeCheckEnv()

	session.FixedTimeSave()
	session.FixedTimeRelease()

	if *confStat {
		session.FixedTimeReport()
	}

	logger.Debug("host ["+runtime.Version()+", "+runtime.GOOS+"_"+runtime.GOARCH+"], cross-compilation ",
		util.Go.GetCrossPlatforms())
}

// Main.
func main() {
	runtime.GOMAXPROCS(conf.Wide.MaxProcs)

	initMime()

	// IDE
	http.HandleFunc(conf.Wide.Context+"/", handlerGzWrapper(indexHandler))
	http.HandleFunc(conf.Wide.Context+"/start", handlerWrapper(startHandler))
	http.HandleFunc(conf.Wide.Context+"/about", handlerWrapper(aboutHandler))
	http.HandleFunc(conf.Wide.Context+"/keyboard_shortcuts", handlerWrapper(keyboardShortcutsHandler))

	// static resources
	http.Handle(conf.Wide.Context+"/static/", http.StripPrefix(conf.Wide.Context+"/static/", http.FileServer(http.Dir("static"))))
	serveSingle("/favicon.ico", "./static/favicon.ico")

	// workspaces
	for _, user := range conf.Users {
		http.Handle(conf.Wide.Context+"/workspace/"+user.Name+"/",
			http.StripPrefix(conf.Wide.Context+"/workspace/"+user.Name+"/", http.FileServer(http.Dir(user.GetWorkspace()))))
	}

	// session
	http.HandleFunc(conf.Wide.Context+"/session/ws", handlerWrapper(session.WSHandler))
	http.HandleFunc(conf.Wide.Context+"/session/save", handlerWrapper(session.SaveContentHandler))

	// run
	http.HandleFunc(conf.Wide.Context+"/build", handlerWrapper(output.BuildHandler))
	http.HandleFunc(conf.Wide.Context+"/run", handlerWrapper(output.RunHandler))
	http.HandleFunc(conf.Wide.Context+"/stop", handlerWrapper(output.StopHandler))
	http.HandleFunc(conf.Wide.Context+"/go/test", handlerWrapper(output.GoTestHandler))
	http.HandleFunc(conf.Wide.Context+"/go/vet", handlerWrapper(output.GoVetHandler))
	http.HandleFunc(conf.Wide.Context+"/go/get", handlerWrapper(output.GoGetHandler))
	http.HandleFunc(conf.Wide.Context+"/go/install", handlerWrapper(output.GoInstallHandler))
	http.HandleFunc(conf.Wide.Context+"/output/ws", handlerWrapper(output.WSHandler))

	// cross-compilation
	http.HandleFunc(conf.Wide.Context+"/cross", handlerWrapper(output.CrossCompilationHandler))

	// file tree
	http.HandleFunc(conf.Wide.Context+"/files", handlerWrapper(file.GetFilesHandler))
	http.HandleFunc(conf.Wide.Context+"/file/refresh", handlerWrapper(file.RefreshDirectoryHandler))
	http.HandleFunc(conf.Wide.Context+"/file", handlerWrapper(file.GetFileHandler))
	http.HandleFunc(conf.Wide.Context+"/file/save", handlerWrapper(file.SaveFileHandler))
	http.HandleFunc(conf.Wide.Context+"/file/new", handlerWrapper(file.NewFileHandler))
	http.HandleFunc(conf.Wide.Context+"/file/remove", handlerWrapper(file.RemoveFileHandler))
	http.HandleFunc(conf.Wide.Context+"/file/rename", handlerWrapper(file.RenameFileHandler))
	http.HandleFunc(conf.Wide.Context+"/file/search/text", handlerWrapper(file.SearchTextHandler))
	http.HandleFunc(conf.Wide.Context+"/file/find/name", handlerWrapper(file.FindHandler))

	// outline
	http.HandleFunc(conf.Wide.Context+"/outline", handlerWrapper(file.GetOutlineHandler))

	// file export/import
	http.HandleFunc(conf.Wide.Context+"/file/zip/new", handlerWrapper(file.CreateZipHandler))
	http.HandleFunc(conf.Wide.Context+"/file/zip", handlerWrapper(file.GetZipHandler))
	http.HandleFunc(conf.Wide.Context+"/file/upload", handlerWrapper(file.UploadHandler))
	http.HandleFunc(conf.Wide.Context+"/file/decompress", handlerWrapper(file.DecompressHandler))

	// editor
	http.HandleFunc(conf.Wide.Context+"/editor/ws", handlerWrapper(editor.WSHandler))
	http.HandleFunc(conf.Wide.Context+"/go/fmt", handlerWrapper(editor.GoFmtHandler))
	http.HandleFunc(conf.Wide.Context+"/autocomplete", handlerWrapper(editor.AutocompleteHandler))
	http.HandleFunc(conf.Wide.Context+"/exprinfo", handlerWrapper(editor.GetExprInfoHandler))
	http.HandleFunc(conf.Wide.Context+"/find/decl", handlerWrapper(editor.FindDeclarationHandler))
	http.HandleFunc(conf.Wide.Context+"/find/usages", handlerWrapper(editor.FindUsagesHandler))

	// shell
	// http.HandleFunc(conf.Wide.Context+"/shell/ws", handlerWrapper(shell.WSHandler))
	// http.HandleFunc(conf.Wide.Context+"/shell", handlerWrapper(shell.IndexHandler))

	// notification
	http.HandleFunc(conf.Wide.Context+"/notification/ws", handlerWrapper(notification.WSHandler))

	// user
	http.HandleFunc(conf.Wide.Context+"/login", handlerWrapper(session.LoginHandler))
	http.HandleFunc(conf.Wide.Context+"/logout", handlerWrapper(session.LogoutHandler))
	http.HandleFunc(conf.Wide.Context+"/signup", handlerWrapper(session.SignUpUserHandler))
	http.HandleFunc(conf.Wide.Context+"/preference", handlerWrapper(session.PreferenceHandler))

	// playground
	http.HandleFunc(conf.Wide.Context+"/playground", handlerWrapper(playground.IndexHandler))
	http.HandleFunc(conf.Wide.Context+"/playground/", handlerWrapper(playground.IndexHandler))
	http.HandleFunc(conf.Wide.Context+"/playground/ws", handlerWrapper(playground.WSHandler))
	http.HandleFunc(conf.Wide.Context+"/playground/save", handlerWrapper(playground.SaveHandler))
	http.HandleFunc(conf.Wide.Context+"/playground/short-url", handlerWrapper(playground.ShortURLHandler))
	http.HandleFunc(conf.Wide.Context+"/playground/build", handlerWrapper(playground.BuildHandler))
	http.HandleFunc(conf.Wide.Context+"/playground/run", handlerWrapper(playground.RunHandler))
	http.HandleFunc(conf.Wide.Context+"/playground/stop", handlerWrapper(playground.StopHandler))
	http.HandleFunc(conf.Wide.Context+"/playground/autocomplete", handlerWrapper(playground.AutocompleteHandler))

	// git
	http.HandleFunc(conf.Wide.Context+"/git/clone", handlerWrapper(git.CloneHandler))

	logger.Infof("Wide is running [%s]", conf.Wide.Server+conf.Wide.Context)

	err := http.ListenAndServe(conf.Wide.Server, nil)
	if err != nil {
		logger.Error(err)
	}
}

// indexHandler handles request of Wide index.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if conf.Wide.Context+"/" != r.RequestURI {
		http.Redirect(w, r, conf.Wide.Context+"/", http.StatusFound)

		return
	}

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Redirect(w, r, conf.Wide.Context+"/login", http.StatusFound)

		return
	}

	username := httpSession.Values["username"].(string)
	if "playground" == username { // reserved user for Playground
		http.Redirect(w, r, conf.Wide.Context+"/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	if "" != conf.Wide.Context {
		httpSession.Options.Path = conf.Wide.Context
	}
	httpSession.Save(r, w)

	user := conf.GetUser(username)
	if nil == user {
		logger.Warnf("Not found user [%s]", username)

		http.Redirect(w, r, conf.Wide.Context+"/login", http.StatusFound)

		return
	}

	locale := user.Locale

	wideSessions := session.WideSessions.GetByUsername(username)

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale,
		"username": username, "sid": session.WideSessions.GenId(), "latestSessionContent": user.LatestSessionContent,
		"pathSeparator": conf.PathSeparator, "codeMirrorVer": conf.CodeMirrorVer,
		"user": user, "editorThemes": conf.GetEditorThemes(), "crossPlatforms": util.Go.GetCrossPlatforms()}

	logger.Debugf("User [%s] has [%d] sessions", username, len(wideSessions))

	t, err := template.ParseFiles("views/index.html")

	if nil != err {
		logger.Error(err)
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
		http.Redirect(w, r, conf.Wide.Context+"/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	if "" != conf.Wide.Context {
		httpSession.Options.Path = conf.Wide.Context
	}
	httpSession.Save(r, w)

	username := httpSession.Values["username"].(string)
	locale := conf.GetUser(username).Locale
	userWorkspace := conf.GetUserWorkspace(username)

	sid := r.URL.Query()["sid"][0]
	wSession := session.WideSessions.Get(sid)
	if nil == wSession {
		logger.Errorf("Session [%s] not found", sid)
	}

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale,
		"username": username, "workspace": userWorkspace, "ver": conf.WideVersion, "sid": sid}

	t, err := template.ParseFiles("views/start.html")

	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, model)
}

// keyboardShortcutsHandler handles request of keyboard shortcuts page.
func keyboardShortcutsHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Redirect(w, r, conf.Wide.Context+"/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	if "" != conf.Wide.Context {
		httpSession.Options.Path = conf.Wide.Context
	}
	httpSession.Save(r, w)

	username := httpSession.Values["username"].(string)
	locale := conf.GetUser(username).Locale

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale}

	t, err := template.ParseFiles("views/keyboard_shortcuts.html")

	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, model)
}

// aboutHandle handles request of about page.
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Redirect(w, r, conf.Wide.Context+"/login", http.StatusFound)

		return
	}

	httpSession.Options.MaxAge = conf.Wide.HTTPSessionMaxAge
	if "" != conf.Wide.Context {
		httpSession.Options.Path = conf.Wide.Context
	}
	httpSession.Save(r, w)

	username := httpSession.Values["username"].(string)
	locale := conf.GetUser(username).Locale

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(locale), "locale": locale,
		"ver": conf.WideVersion, "goos": runtime.GOOS, "goarch": runtime.GOARCH, "gover": runtime.Version()}

	t, err := template.ParseFiles("views/about.html")

	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

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
