package session

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

const (
	UserExists      = "user exists"
	UserCreated     = "user created"
	UserCreateError = "user create error"
)

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

	msg := addUser(username, password)
	if UserCreated != msg {
		succ = false
		data["msg"] = msg
	}
}

func addUser(username, password string) string {
	// XXX: validate

	for _, user := range conf.Wide.Users {
		if user.Name == username {
			return UserExists
		}
	}

	firstUserWorkspace := conf.Wide.GetUserWorkspace(conf.Wide.Users[0].Name)
	dir := filepath.Dir(firstUserWorkspace)
	workspace := filepath.Join(dir, username)

	newUser := &conf.User{Name: username, Password: password, Workspace: workspace,
		Locale: conf.Wide.Locale, GoFormat: "gofmt", Editor: &conf.Editor{FontSize: "10px"}}
	conf.Wide.Users = append(conf.Wide.Users, newUser)

	if !conf.Save() {
		return UserCreateError
	}

	conf.CreateWorkspaceDir(workspace)
	conf.UpdateCustomizedConf(username)

	glog.Infof("Created a user [%s]", username)

	return UserCreated
}
