package conf

import (
	"encoding/json"
	"flag"
	"github.com/88250/wide/util"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Conf struct {
	Server                string
	StaticServer          string
	EditorChannel         string
	OutputChannel         string
	ShellChannel          string
	StaticResourceVersion string
	ContextPath           string
	StaticPath            string
	RuntimeMode           string
	GOPATH                string
}

var Wide Conf

func init() {
	flag.Set("logtostderr", "true")

	flag.Parse()

	bytes, _ := ioutil.ReadFile("conf/wide.json")

	err := json.Unmarshal(bytes, &Wide)
	if err != nil {
		glog.Error(err)

		os.Exit(-1)
	}

	ip, err := util.Net.LocalIP()
	if err != nil {
		glog.Error(err)

		os.Exit(-1)
	}

	glog.Infof("IP [%s]", ip)
	Wide.Server = strings.Replace(Wide.Server, "{IP}", ip, 1)
	Wide.StaticServer = strings.Replace(Wide.StaticServer, "{IP}", ip, 1)

	Wide.EditorChannel = strings.Replace(Wide.EditorChannel, "{IP}", ip, 1)
	Wide.OutputChannel = strings.Replace(Wide.OutputChannel, "{IP}", ip, 1)
	Wide.ShellChannel = strings.Replace(Wide.ShellChannel, "{IP}", ip, 1)

	file, _ := exec.LookPath(os.Args[0])
	pwd, _ := filepath.Abs(file)
	pwd = pwd[:strings.LastIndex(pwd, string(os.PathSeparator))]
	glog.Infof("pwd [%s]", pwd)
	Wide.GOPATH = strings.Replace(Wide.GOPATH, "{pwd}", pwd, 1)

	glog.Info("Conf: \n" + string(bytes))
}
