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
	"github.com/b3log/wide/file"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/output"
	"github.com/b3log/wide/shell"
	"github.com/b3log/wide/user"
	"github.com/golang/glog"
)

// Wide 中唯一一个 init 函数.
func init() {
	flag.Set("logtostderr", "true")
	flag.Set("v", "1")

	flag.Parse()
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	i18n.Load()

	model := map[string]interface{}{"Wide": conf.Wide, "i18n": i18n.GetLangs(r), "locale": i18n.GetLocale(r)}

	session, _ := user.Session.Get(r, "wide-session")

	if session.IsNew {
		// TODO: 写死以 admin 作为用户登录
		name := conf.Wide.Users[0].Name

		session.Values["username"] = name
		session.Values["id"] = strconv.Itoa(rand.Int())
		// 一天过期
		session.Options.MaxAge = 60 * 60 * 24

		glog.Infof("Created a session [%s] for user [%s]", session.Values["id"].(string), name)
	}

	session.Save(r, w)

	t, err := template.ParseFiles("view/index.html")

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, model)
}

func main() {
	conf.Load()

	runtime.GOMAXPROCS(conf.Wide.MaxProcs)

	defer glog.Flush()

	// 静态资源
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

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
	http.HandleFunc("/finddecl", editor.FindDeclarationHandler)
	http.HandleFunc("/html/fmt", editor.HTMLFmtHandler)
	http.HandleFunc("/json/fmt", editor.JSONFmtHandler)

	// Shell
	http.HandleFunc("/shell/ws", shell.WSHandler)
	http.HandleFunc("/shell", shell.IndexHandler)

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
