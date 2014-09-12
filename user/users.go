// 用户操作.
package user

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

const (
	USER_EXISTS        = "user exists"
	USER_CREATED       = "user created"
	USER_CREATE_FAILED = "user create failed"
)

// 添加用户.
func AddUser(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	username := args["username"].(string)
	password := args["password"].(string)

	msg := addUser(username, password)
	if USER_CREATED != msg {
		data["succ"] = false
		data["msg"] = msg
	}
}

// 初始化用户 git 仓库.
func InitGitRepos(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	session, _ := Session.Get(r, "wide-session")

	username := session.Values["username"].(string)
	userRepos := conf.Wide.UserWorkspaces + string(os.PathSeparator) + username + string(os.PathSeparator) + "src"

	// TODO: git clone

	glog.Infof("Git Cloned from [%s] to [%s]", conf.Wide.Workspace+string(os.PathSeparator)+"src", userRepos)
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
