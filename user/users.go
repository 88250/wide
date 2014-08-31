package user

import (
	"encoding/json"
	"github.com/b3log/wide/conf"
	"github.com/golang/glog"
	"net/http"
	"strings"
)

const (
	USER_EXISTS        = "user exists"
	USER_CREATED       = "user created"
	USER_CREATE_FAILED = "user create failed"
)

func AddUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	username := args["username"].(string)
	password := args["password"].(string)

	data := map[string]interface{}{"succ": true}

	msg := addUser(username, password)
	if USER_CREATED != msg {
		data["succ"] = false
		data["msg"] = msg
	}

	ret, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func InitGitRepos(w http.ResponseWriter, r *http.Request) {
	session, _ := Session.Get(r, "wide-session")

	username := session.Values["username"].(string)
	userRepos := strings.Replace(conf.Wide.UserRepos, "{user}", username, -1)

	data := map[string]interface{}{"succ": true}

	// TODO: git clone

	glog.Infof("Git Cloned from [%s] to [%s]", conf.Wide.Repos, userRepos)

	ret, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func addUser(username, password string) string {
	// TODO: https://github.com/b3log/wide/issues/23
	conf.Load()

	// XXX: 新建用户校验增强
	for _, user := range conf.Wide.Users {
		if user.Name == username {
			return USER_EXISTS
		}
	}

	newUser := conf.User{Name: username, Password: password}
	conf.Wide.Users = append(conf.Wide.Users, newUser)

	if !conf.Save() {
		return USER_CREATE_FAILED
	}

	glog.Infof("Created a user [%s]", username)

	return USER_CREATED
}
