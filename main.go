package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
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

// Wide 中唯一一个 init 函数.
func init() {
	// TODO:默认启动参数
	flag.Set("logtostderr", "true")
	flag.Set("v", "3")
	flag.Parse()

	// 加载事件处理
	event.Load()

	// 加载配置
	conf.Load()

	// 定时检查运行环境
	conf.FixedTimeCheckEnv()

	// 定时保存配置
	conf.FixedTimeSave()

	// 定时检查无效会话
	session.FixedTimeRelease()
}

// 登录.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	i18n.Load()

	if r.Method == "GET" {
		// 展示登录页面

		model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(r), "locale": i18n.GetLocale(r)}

		t, err := template.ParseFiles("view/login.html")

		if nil != err {
			glog.Error(err)
			http.Error(w, err.Error(), 500)

			return
		}

		t.Execute(w, model)

		return
	}

	// 非 GET 请求当作是登录请求
	succ := false

	data := map[string]interface{}{"succ": &succ}
	defer util.RetJSON(w, r, data)

	args := struct {
		Username string
		Password string
	}{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		succ = true

		return
	}

	for _, user := range conf.Wide.Users {
		if user.Name == args.Username && user.Password == args.Password {
			succ = true
		}
	}

	if !succ {
		return
	}

	// 创建 HTTP 会话
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	httpSession.Values["username"] = args.Username
	httpSession.Values["id"] = strconv.Itoa(rand.Int())
	httpSession.Options.MaxAge = 60 * 60 * 24 // 一天过期
	httpSession.Save(r, w)

	glog.Infof("Created a HTTP session [%s] for user [%s]", httpSession.Values["id"].(string), args.Username)
}

// Wide 首页.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	i18n.Load()

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")

	if httpSession.IsNew {
		http.Redirect(w, r, "/login", http.StatusForbidden)

		return
	}

	httpSession.Save(r, w)

	// 创建一个 Wide 会话
	wideSession := session.WideSessions.New(httpSession)

	username := httpSession.Values["username"].(string)

	wideSessions := session.WideSessions.GetByUsername(username)
	userConf := conf.Wide.GetUser(username)

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(r), "locale": i18n.GetLocale(r),
		"session": wideSession, "latestSessionContent": userConf.LatestSessionContent}

	glog.V(3).Infof("User [%s] has [%d] sessions", username, len(wideSessions))

	t, err := template.ParseFiles("view/index.html")

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, model)
}

// favicon.ico 请求处理.
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: favicon.ico 请求处理
}

// 主程序入口.
func main() {
	runtime.GOMAXPROCS(conf.Wide.MaxProcs)

	defer glog.Flush()

	// 静态资源
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/favicon.ico", handlerWrapper(faviconHandler))

	// 库资源
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("data"))))

	// IDE
	http.HandleFunc("/login", handlerWrapper(loginHandler))
	http.HandleFunc("/", handlerWrapper(indexHandler))

	// 会话
	http.HandleFunc("/session/ws", handlerWrapper(session.WSHandler))
	http.HandleFunc("/session/save", handlerWrapper(session.SaveContent))

	// 运行相关
	http.HandleFunc("/build", handlerWrapper(output.BuildHandler))
	http.HandleFunc("/run", handlerWrapper(output.RunHandler))
	http.HandleFunc("/stop", handlerWrapper(output.StopHandler))
	http.HandleFunc("/go/get", handlerWrapper(output.GoGetHandler))
	http.HandleFunc("/go/install", handlerWrapper(output.GoInstallHandler))
	http.HandleFunc("/output/ws", handlerWrapper(output.WSHandler))

	// 文件树
	http.HandleFunc("/files", handlerWrapper(file.GetFiles))
	http.HandleFunc("/file", handlerWrapper(file.GetFile))
	http.HandleFunc("/file/save", handlerWrapper(file.SaveFile))
	http.HandleFunc("/file/new", handlerWrapper(file.NewFile))
	http.HandleFunc("/file/remove", handlerWrapper(file.RemoveFile))

	// 编辑器
	http.HandleFunc("/editor/ws", handlerWrapper(editor.WSHandler))
	http.HandleFunc("/go/fmt", handlerWrapper(editor.GoFmtHandler))
	http.HandleFunc("/autocomplete", handlerWrapper(editor.AutocompleteHandler))
	http.HandleFunc("/exprinfo", handlerWrapper(editor.GetExprInfoHandler))
	http.HandleFunc("/find/decl", handlerWrapper(editor.FindDeclarationHandler))
	http.HandleFunc("/find/usages", handlerWrapper(editor.FindUsagesHandler))
	http.HandleFunc("/html/fmt", handlerWrapper(editor.HTMLFmtHandler))
	http.HandleFunc("/json/fmt", handlerWrapper(editor.JSONFmtHandler))

	// Shell
	http.HandleFunc("/shell/ws", handlerWrapper(shell.WSHandler))
	http.HandleFunc("/shell", handlerWrapper(shell.IndexHandler))

	// 通知
	http.HandleFunc("/notification/ws", handlerWrapper(notification.WSHandler))

	// 用户
	http.HandleFunc("/user/new", handlerWrapper(session.AddUser))
	http.HandleFunc("/user/repos/init", handlerWrapper(session.InitGitRepos))

	// 文档
	http.Handle("/doc/", http.StripPrefix("/doc/", http.FileServer(http.Dir("doc"))))

	glog.V(0).Infof("Wide is running [%s]", conf.Wide.Server)

	err := http.ListenAndServe(conf.Wide.Server, nil)
	if err != nil {
		glog.Fatal(err)
	}
}

// HTTP Handler 包装，完成共性处理.
//
// 共性处理：
//
//  1. panic recover
//  2. 请求计时
func handlerWrapper(f func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	handler := panicRecover(f)
	handler = stopwatch(handler)

	return handler
}

// Handler 包装请求计时.
func stopwatch(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		defer func() {
			glog.V(5).Infof("[%s] [%s]", r.RequestURI, time.Since(start))
		}()

		// Handler 处理
		handler(w, r)
	}
}

// Handler 包装 recover panic.
func panicRecover(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer util.Recover()

		// Handler 处理
		handler(w, r)
	}
}
