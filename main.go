package main

import (
	"flag"
	"html/template"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/editor"
	"github.com/b3log/wide/event"
	"github.com/b3log/wide/file"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/notification"
	"github.com/b3log/wide/output"
	"github.com/b3log/wide/shell"
	"github.com/b3log/wide/user"
	"github.com/golang/glog"
)

// Wide 中唯一一个 init 函数.
func init() {
	// 默认启动参数
	flag.Set("logtostderr", "true")
	flag.Set("v", "3")
	flag.Parse()

	// 加载事件处理
	event.Load()

	// 加载配置
	conf.Load()

	// 定时检查 Wide 运行环境
	conf.CheckEnv()
}

// Wide 首页.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// 创建一个 Wide 会话
	wideSession := user.WideSessions.New()

	i18n.Load()

	model := map[string]interface{}{"conf": conf.Wide, "i18n": i18n.GetAll(r), "locale": i18n.GetLocale(r),
		"session": wideSession}

	httpSession, _ := user.HTTPSession.Get(r, "wide-session")

	httpSessionId := httpSession.Values["id"].(string)
	// TODO: 写死以 admin 作为用户登录
	username := conf.Wide.Users[0].Name
	if httpSession.IsNew {

		httpSession.Values["username"] = username
		httpSessionId = strconv.Itoa(rand.Int())
		httpSession.Values["id"] = httpSessionId
		// 一天过期
		httpSession.Options.MaxAge = 60 * 60 * 24

		glog.Infof("Created a HTTP session [%s] for user [%s]", httpSession.Values["id"].(string), username)
	}

	httpSession.Save(r, w)

	// Wide 会话关联 HTTP 会话
	wideSession.HTTPSessionId = httpSession.Values["id"].(string)

	wideSessions := user.WideSessions.GetByHTTPSid(httpSessionId)
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
	// TODO: favicon.ico 请求处理.
}

// 主程序入口.
func main() {
	runtime.GOMAXPROCS(conf.Wide.MaxProcs)

	defer glog.Flush()

	// 静态资源
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/favicon.ico", faviconHandler)

	// 库资源
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("data"))))

	// IDE 首页
	http.HandleFunc("/", indexHandler)

	// 运行相关
	http.HandleFunc("/build", output.BuildHandler)
	http.HandleFunc("/run", output.RunHandler)
	http.HandleFunc("/go/get", output.GoGetHandler)
	http.HandleFunc("/go/install", output.GoInstallHandler)
	http.HandleFunc("/output/ws", output.WSHandler)

	// 文件树
	http.HandleFunc("/files", file.GetFiles)
	http.HandleFunc("/file", file.GetFile)
	http.HandleFunc("/file/save", file.SaveFile)
	http.HandleFunc("/file/new", file.NewFile)
	http.HandleFunc("/file/remove", file.RemoveFile)

	// 编辑器
	http.HandleFunc("/editor/ws", editor.WSHandler)
	http.HandleFunc("/go/fmt", editor.GoFmtHandler)
	http.HandleFunc("/autocomplete", editor.AutocompleteHandler)
	http.HandleFunc("/find/decl", editor.FindDeclarationHandler)
	http.HandleFunc("/find/usages", editor.FindUsagesHandler)
	http.HandleFunc("/html/fmt", editor.HTMLFmtHandler)
	http.HandleFunc("/json/fmt", editor.JSONFmtHandler)

	// Shell
	http.HandleFunc("/shell/ws", shell.WSHandler)
	http.HandleFunc("/shell", shell.IndexHandler)

	// 通知
	http.HandleFunc("/notification/ws", notification.WSHandler)

	// 用户
	http.HandleFunc("/user/new", user.AddUser)
	http.HandleFunc("/user/repos/init", user.InitGitRepos)

	// 文档
	http.Handle("/doc/", http.StripPrefix("/doc/", http.FileServer(http.Dir("doc"))))

	glog.V(0).Infof("Wide is running [%s]", conf.Wide.Server)

	err := http.ListenAndServe(conf.Wide.Server, nil)
	if err != nil {
		glog.Fatal(err)
	}
}
