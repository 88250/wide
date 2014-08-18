package main

// TODO:
// 1. 编辑锁
//

import (
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/editor"
	"github.com/b3log/wide/file"
	"github.com/b3log/wide/output"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/shell"
	"github.com/golang/glog"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "wide-session")

	if session.IsNew {
		session.Values["id"] = strconv.Itoa(rand.Int())
	}

	session.Save(r, w)

	t, err := template.ParseFiles("templates/index.html")

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	t.Execute(w, map[string]interface{}{"Wide": conf.Wide})
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", indexHandler)

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

	err := http.ListenAndServe(conf.Wide.Server, nil)
	if err != nil {
		glog.Fatal(err)
	}
}
