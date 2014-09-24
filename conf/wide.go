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
	"time"

	"github.com/b3log/wide/event"
	_ "github.com/b3log/wide/i18n"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

// 最后一次会话内容结构.
type LatestSessionContent struct {
	FileTree    []string `json:"fileTree"`    // 文件树展开的路径集
	Files       []string `json:"files"`       // 编辑器打开的文件路径集
	CurrentFile string   `json:"currentFile"` // 当前编辑器文件路径
}

// 用户结构.
type User struct {
	Name                 string
	Password             string
	Workspace            string // 指定了该用户的 GOPATH 路径
	LatestSessionContent *LatestSessionContent
}

// 配置结构.
type conf struct {
	Server                string
	StaticServer          string
	EditorChannel         string
	OutputChannel         string
	ShellChannel          string
	SessionChannel        string
	StaticResourceVersion string
	MaxProcs              int
	RuntimeMode           string
	Pwd                   string
	Users                 []*User
}

// 配置.
var Wide conf

// 维护非变化部分的配置.
// 只有 Users 是会运行时变化的，保存回写文件时要使用这个变量.
var rawWide conf

// 定时检查 Wide 运行环境.
// 如果是特别严重的问题（比如 $GOPATH 不存在）则退出进程，另一些不太严重的问题（比如 gocode 不存在）则放入全局通知队列.
func FixedTimeCheckEnv() {
	go func() {
		for {
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
				event.EventQueue <- event.EvtCodeGocodeNotFound
				glog.Warningf("Not found gocode [%s]", gocode)
			}

			ide_stub := Wide.GetIDEStub()
			cmd = exec.Command(ide_stub, "version")
			_, err = cmd.Output()
			if nil != err {
				event.EventQueue <- event.EvtCodeIDEStubNotFound
				glog.Warningf("Not found ide_stub [%s]", ide_stub)
			}

			// TODO: 7 分钟进行一次检查环境
			time.Sleep(time.Second * 7)
		}
	}()
}

// 定时（10 分钟）保存配置.
// 主要目的是保存用户会话内容，以备下一次用户打开 Wide 时进行会话还原.
func FixedTimeSave() {
	go func() {
		for {
			Save()

			// 10 分钟进行一次配置保存
			time.Sleep(time.Minute * 10)
		}
	}()
}

// 获取 username 指定的用户的工作空间路径.
// 查找不到时返回空字符串.
func (*conf) GetUserWorkspace(username string) string {
	for _, user := range Wide.Users {
		if user.Name == username {
			ret := strings.Replace(user.Workspace, "{pwd}", Wide.Pwd, 1)
			return filepath.FromSlash(ret)
		}
	}

	return ""
}

// 获取 username 指定的用户配置.
func (*conf) GetUser(username string) *User {
	for _, user := range Wide.Users {
		if user.Name == username {
			return user
		}
	}

	return nil
}

// 获取 gocode 路径.
func (*conf) GetGocode() string {
	return getGOBIN() + "gocode"
}

// 获取 ide_stub 路径.
func (*conf) GetIDEStub() string {
	return getGOBIN() + "ide_stub"
}

// 获取 GOBIN 路径，末尾带路径分隔符.
func getGOBIN() string {
	ret := os.Getenv("GOBIN")
	if "" != ret {
		return ret + string(os.PathSeparator)
	}

	ret = os.Getenv("GOPATH") + string(os.PathSeparator) + "bin" + string(os.PathSeparator) +
		os.Getenv("GOOS") + "_" + os.Getenv("GOARCH")
	if isExist(ret) {
		return ret + string(os.PathSeparator)
	}

	ret = os.Getenv("GOPATH") + string(os.PathSeparator) + "bin" + string(os.PathSeparator) +
		runtime.GOOS + "_" + runtime.GOARCH
	if isExist(ret) {
		return ret + string(os.PathSeparator)
	}

	return os.Getenv("GOPATH") + string(os.PathSeparator) + "bin" + string(os.PathSeparator)
}

// 保存 Wide 配置.
func Save() bool {
	// 只有 Users 是会运行时变化的，其他属性只能手工维护 wide.json 配置文件
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

	// TODO: 弄个反射吧？
	Wide.Server = strings.Replace(Wide.Server, "{IP}", ip, 1)
	Wide.StaticServer = strings.Replace(Wide.StaticServer, "{IP}", ip, 1)
	Wide.EditorChannel = strings.Replace(Wide.EditorChannel, "{IP}", ip, 1)
	Wide.OutputChannel = strings.Replace(Wide.OutputChannel, "{IP}", ip, 1)
	Wide.ShellChannel = strings.Replace(Wide.ShellChannel, "{IP}", ip, 1)
	Wide.SessionChannel = strings.Replace(Wide.SessionChannel, "{IP}", ip, 1)

	// 获取当前执行路径
	file, _ := exec.LookPath(os.Args[0])
	pwd, _ := filepath.Abs(file)
	pwd = pwd[:strings.LastIndex(pwd, string(os.PathSeparator))]
	Wide.Pwd = pwd
	glog.V(3).Infof("pwd [%s]", pwd)

	glog.V(3).Info("Conf: \n" + string(bytes))
}

// 检查文件或目录是否存在.
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false.
func isExist(filename string) bool {
	_, err := os.Stat(filename)

	return err == nil || os.IsExist(err)
}
