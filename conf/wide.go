// Wide 配置相关，所有配置（包括用户配置）都是保存在 wide.json 中.
package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/b3log/wide/event"
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
	MaxProcs              int
	RuntimeMode           string
	Pwd                   string
	Users                 []User
}

var Wide conf
var rawWide conf

// 检查 Wide 运行环境.
// 如果是特别严重的问题（比如 $GOPATH 不存在）则退出进程。另一些不太严重的问题（比如 gocode 不存在）则放入全局通知队列。
func (*conf) CheckEnv() {
	if "" == os.Getenv("GOPATH") {
		glog.Fatal("Not found $GOPATH")
		os.Exit(-1)
	}

	if "" == os.Getenv("GOROOT") {
		glog.Fatal("Not found $GOROOT")

		os.Exit(-1)
	}

	gocode := Wide.GetGocode()
	cmd := exec.Command(gocode, "close")
	_, err := cmd.Output()
	if nil != err {
		event.EventQueue <- event.EvtGocodeNotFount
		glog.Warning("Not found gocode")
	}

	ide_stub := Wide.GetIDEStub()
	cmd = exec.Command(ide_stub, "version")
	_, err = cmd.Output()
	if nil != err {
		glog.Info(err)
		event.EventQueue <- event.EvtIDEStubNotFound
		glog.Warning("Not found ide_stub")
	}
}

// 获取 username 指定的用户的工作空间路径.
func (*conf) GetUserWorkspace(username string) string {
	for _, user := range Wide.Users {
		if user.Name == username {
			ret := strings.Replace(user.Workspace, "{Pwd}", Wide.Pwd, 1)
			return filepath.FromSlash(ret)
		}
	}

	return ""
}

// 获取 gocode 路径.
func (*conf) GetGocode() string {
	if "" != os.Getenv("GOARCH") {
		return os.Getenv("GOPATH") + string(os.PathSeparator) + "bin" + string(os.PathSeparator) +
			runtime.GOOS + "_" + os.Getenv("GOARCH") + string(os.PathSeparator) + "gocode"
	} else {
		return os.Getenv("GOPATH") + string(os.PathSeparator) + "bin" + string(os.PathSeparator) +
			"gocode"
	}
}

// 获取 ide_stub 路径.
func (*conf) GetIDEStub() string {
	if "" != os.Getenv("GOARCH") {
		return os.Getenv("GOPATH") + string(os.PathSeparator) + "bin" + string(os.PathSeparator) +
			runtime.GOOS + "_" + os.Getenv("GOARCH") + string(os.PathSeparator) + "ide_stub"
	} else {
		return os.Getenv("GOPATH") + string(os.PathSeparator) + "bin" + string(os.PathSeparator) +
			"ide_stub"
	}
}

// 保存 Wide 配置.
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

// 加载 Wide 配置.
func Load() {
	// 检查 Wide 运行环境
	Wide.CheckEnv()

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
