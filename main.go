package main

import (
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/editor"
	"github.com/b3log/wide/file"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/output"
	"github.com/b3log/wide/shell"
	"github.com/b3log/wide/user"
	"github.com/golang/glog"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	model := map[string]interface{}{"Wide": conf.Wide, "i18n": i18n.GetLangs(r)}

	session, _ := user.Session.Get(r, "wide-session")

	if session.IsNew {
		// TODO: 以 admin 作为用户登录
		name := conf.Wide.Users[0].Name
		glog.Infof("[%s] logged in", name)

		session.Values["username"] = name
		session.Values["id"] = strconv.Itoa(rand.Int())
	}

	session.Save(r, w)

	t, err := template.ParseFiles("templates/index.html")

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, model)
}

func main() {
	conf.Load()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", indexHandler)

	http.HandleFunc("/build", output.BuildHandler)
	http.HandleFunc("/run", output.RunHandler)
	http.HandleFunc("/output/ws", output.WSHandler)

	http.HandleFunc("/files", file.GetFiles)
	http.HandleFunc("/file", file.GetFile)
	http.HandleFunc("/file/save", file.SaveFile)
	http.HandleFunc("/file/new", file.NewFile)
	http.HandleFunc("/file/remove", file.RemoveFile)

	http.HandleFunc("/editor/ws", editor.WSHandler)
	http.HandleFunc("/fmt", editor.FmtHandler)

	http.HandleFunc("/shell/ws", shell.WSHandler)

	http.HandleFunc("/autocomplete", editor.AutocompleteHandler)

	http.HandleFunc("/user/new", user.AddUser)
	http.HandleFunc("/user/repos/init", user.InitGitRepos)

	glog.Infof("Wide is running [%s]", conf.Wide.Server)

	err := http.ListenAndServe(conf.Wide.Server, nil)
	if err != nil {
		glog.Fatal(err)
	}
}
