// Wide 配置相关，所有配置（包括用户配置）都是保存在 wide.json 中.
package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "github.com/b3log/wide/i18n"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

type User struct {
	Name      string
	Password  string
	Workspace string // 指定了该用户的 GOPATH 路径
}

type conf struct {
	Server                string
	StaticServer          string
	EditorChannel         string
	OutputChannel         string
	ShellChannel          string
	StaticResourceVersion string
	ContextPath           string
	StaticPath            string
	MaxProcs              int
	RuntimeMode           string
	Pwd                   string
	Users                 []User
}

var Wide conf
var rawWide conf

// 获取 username 指定的用户的工作空间路径.
func (this *conf) GetUserWorkspace(username string) string {
	for _, user := range Wide.Users {
		if user.Name == username {
			ret := strings.Replace(user.Workspace, "{Pwd}", Wide.Pwd, 1)
			return filepath.FromSlash(ret)
		}
	}

	return ""
}

func Save() bool {
	// 只有 Users 是可以通过界面修改的，其他属性只能手工维护 wide.json 配置文件
	rawWide.Users = Wide.Users

	// 原始配置文件内容
	bytes, err := json.MarshalIndent(rawWide, "", "    ")

	if nil != err {
		glog.Error(err)

		return false
	}

	if err = ioutil.WriteFile("conf/wide.json", bytes, 0644); nil != err {
		glog.Error(err)

		return false
	}

	return true
}

func Load() {
	bytes, _ := ioutil.ReadFile("conf/wide.json")

	err := json.Unmarshal(bytes, &Wide)
	if err != nil {
		glog.Error(err)

		os.Exit(-1)
	}

	// 保存未经变量替换处理的原始配置文件，用于写回时
	json.Unmarshal(bytes, &rawWide)

	ip, err := util.Net.LocalIP()
	if err != nil {
		glog.Error(err)

		os.Exit(-1)
	}

	glog.V(3).Infof("IP [%s]", ip)
	Wide.Server = strings.Replace(Wide.Server, "{IP}", ip, 1)
	Wide.StaticServer = strings.Replace(Wide.StaticServer, "{IP}", ip, 1)
	Wide.EditorChannel = strings.Replace(Wide.EditorChannel, "{IP}", ip, 1)
	Wide.OutputChannel = strings.Replace(Wide.OutputChannel, "{IP}", ip, 1)
	Wide.ShellChannel = strings.Replace(Wide.ShellChannel, "{IP}", ip, 1)

	// 获取当前执行路径
	file, _ := exec.LookPath(os.Args[0])
	pwd, _ := filepath.Abs(file)
	pwd = pwd[:strings.LastIndex(pwd, string(os.PathSeparator))]
	Wide.Pwd = pwd
	glog.V(3).Infof("pwd [%s]", pwd)

	glog.V(3).Info("Conf: \n" + string(bytes))
}
