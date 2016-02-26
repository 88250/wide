// Copyright (c) 2014-2016, b3log.org & hacpai.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package conf includes configurations related manipulations.
package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/b3log/wide/event"
	"github.com/b3log/wide/log"
	"github.com/b3log/wide/util"
)

const (
	// PathSeparator holds the OS-specific path separator.
	PathSeparator = string(os.PathSeparator)
	// PathListSeparator holds the OS-specific path list separator.
	PathListSeparator = string(os.PathListSeparator)

	// WideVersion holds the current Wide's version.
	WideVersion = "1.5.0"
	// CodeMirrorVer holds the current editor version.
	CodeMirrorVer = "5.1"

	HelloWorld = `package main

import "fmt"

func main() {
	fmt.Println("Hello, 世界")
}
`
)

// Configuration.
type conf struct {
	IP                    string // server ip, ${ip}
	Port                  string // server port
	Context               string // server context
	Server                string // server host and port ({IP}:{Port})
	StaticServer          string // static resources server scheme, host and port (http://{IP}:{Port})
	LogLevel              string // logging level: trace/debug/info/warn/error
	Channel               string // channel (ws://{IP}:{Port})
	HTTPSessionMaxAge     int    // HTTP session max age (in seciond)
	StaticResourceVersion string // version of static resources
	MaxProcs              int    // Go max procs
	RuntimeMode           string // runtime mode (dev/prod)
	WD                    string // current working direcitory, ${pwd}
	Locale                string // default locale
	Playground            string // playground directory
	AllowRegister         bool   // allow register or not
	Autocomplete          bool   // default autocomplete
}

// Logger.
var logger = log.NewLogger(os.Stdout)

// Wide configurations.
var Wide *conf

// configurations of users.
var Users []*User

// Indicates whether runs via Docker.
var Docker bool

// Load loads the Wide configurations from wide.json and users' configurations from users/{username}.json.
func Load(confPath, confIP, confPort, confServer, confLogLevel, confStaticServer, confContext, confChannel,
	confPlayground string, confDocker bool) {
	// XXX: ugly args list....

	initWide(confPath, confIP, confPort, confServer, confLogLevel, confStaticServer, confContext, confChannel,
		confPlayground, confDocker)
	initUsers()
}

func initUsers() {
	f, err := os.Open("conf/users")
	if nil != err {
		logger.Error(err)

		os.Exit(-1)
	}

	names, err := f.Readdirnames(-1)
	if nil != err {
		logger.Error(err)

		os.Exit(-1)
	}
	f.Close()

	for _, name := range names {
		if strings.HasPrefix(name, ".") { // hiden files that not be created by Wide
			continue
		}

		if ".json" != filepath.Ext(name) { // such as backup (*.json~) not be created by Wide
			continue
		}

		user := &User{}

		bytes, _ := ioutil.ReadFile("conf/users/" + name)

		err := json.Unmarshal(bytes, user)
		if err != nil {
			logger.Errorf("Parses [%s] error: %v, skip loading this user", name, err)

			continue
		}

		// Compatibility upgrade (1.3.0): https://github.com/b3log/wide/issues/83
		if "" == user.Keymap {
			user.Keymap = "wide"
		}

		Users = append(Users, user)
	}

	initWorkspaceDirs()
	initCustomizedConfs()
}

func initWide(confPath, confIP, confPort, confServer, confLogLevel, confStaticServer, confContext, confChannel,
	confPlayground string, confDocker bool) {
	bytes, err := ioutil.ReadFile(confPath)
	if nil != err {
		logger.Error(err)

		os.Exit(-1)
	}

	Wide = &conf{}

	err = json.Unmarshal(bytes, Wide)
	if err != nil {
		logger.Error("Parses [wide.json] error: ", err)

		os.Exit(-1)
	}

	// Logging Level
	log.SetLevel(Wide.LogLevel)
	if "" != confLogLevel {
		Wide.LogLevel = confLogLevel
		log.SetLevel(confLogLevel)
	}

	logger.Debug("Conf: \n" + string(bytes))

	// Working Directory
	Wide.WD = util.OS.Pwd()
	logger.Debugf("${pwd} [%s]", Wide.WD)

	// User Home
	home, err := util.OS.Home()
	if nil != err {
		logger.Error("Can't get user's home, please report this issue to developer", err)

		os.Exit(-1)
	}

	logger.Debugf("${user.home} [%s]", home)

	// Playground Directory
	Wide.Playground = strings.Replace(Wide.Playground, "${home}", home, 1)
	if "" != confPlayground {
		Wide.Playground = confPlayground
	}

	if !util.File.IsExist(Wide.Playground) {
		if err := os.Mkdir(Wide.Playground, 0775); nil != err {
			logger.Errorf("Create Playground [%s] error", err)

			os.Exit(-1)
		}
	}

	// IP
	if "" != confIP {
		Wide.IP = confIP
	} else {
		ip, err := util.Net.LocalIP()
		if nil != err {
			logger.Error(err)

			os.Exit(-1)
		}

		logger.Debugf("${ip} [%s]", ip)
		Wide.IP = strings.Replace(Wide.IP, "${ip}", ip, 1)
	}

	if "" != confPort {
		Wide.Port = confPort
	}

	// Docker flag
	Docker = confDocker

	// Server
	Wide.Server = strings.Replace(Wide.Server, "{IP}", Wide.IP, 1)
	Wide.Server = strings.Replace(Wide.Server, "{Port}", Wide.Port, 1)
	if "" != confServer {
		Wide.Server = confServer
	}

	// Static Server
	Wide.StaticServer = strings.Replace(Wide.StaticServer, "{IP}", Wide.IP, 1)
	Wide.StaticServer = strings.Replace(Wide.StaticServer, "{Port}", Wide.Port, 1)
	if "" != confStaticServer {
		Wide.StaticServer = confStaticServer
	}

	// Context
	if "" != confContext {
		Wide.Context = confContext
	}

	time := strconv.FormatInt(time.Now().UnixNano(), 10)
	logger.Debugf("${time} [%s]", time)
	Wide.StaticResourceVersion = strings.Replace(Wide.StaticResourceVersion, "${time}", time, 1)

	// Channel
	Wide.Channel = strings.Replace(Wide.Channel, "{IP}", Wide.IP, 1)
	Wide.Channel = strings.Replace(Wide.Channel, "{Port}", Wide.Port, 1)
	if "" != confChannel {
		Wide.Channel = confChannel
	}
}

// FixedTimeCheckEnv checks Wide runtime enviorment periodically (7 minutes).
//
// Exits process if found fatal issues (such as not found $GOPATH),
// Notifies user by notification queue if found warning issues (such as not found gocode).
func FixedTimeCheckEnv() {
	checkEnv() // check immediately

	go func() {
		for _ = range time.Tick(time.Minute * 7) {
			checkEnv()
		}
	}()
}

func checkEnv() {
	defer util.Recover()

	cmd := exec.Command("go", "version")
	buf, err := cmd.CombinedOutput()
	if nil != err {
		logger.Error("Not found 'go' command, please make sure Go has been installed correctly")

		os.Exit(-1)
	}
	logger.Trace(string(buf))

	if "" == os.Getenv("GOPATH") {
		logger.Error("Not found $GOPATH, please configure it before running Wide")

		os.Exit(-1)
	}

	gocode := util.Go.GetExecutableInGOBIN("gocode")
	cmd = exec.Command(gocode)
	_, err = cmd.Output()
	if nil != err {
		event.EventQueue <- &event.Event{Code: event.EvtCodeGocodeNotFound}

		logger.Warnf("Not found gocode [%s], please install it with this command: go get github.com/nsf/gocode", gocode)
	}

	ideStub := util.Go.GetExecutableInGOBIN("gotools")
	cmd = exec.Command(ideStub, "version")
	_, err = cmd.Output()
	if nil != err {
		event.EventQueue <- &event.Event{Code: event.EvtCodeIDEStubNotFound}

		logger.Warnf("Not found gotools [%s], please install it with this command: go get github.com/visualfc/gotools", ideStub)
	}
}

// GetUserWorkspace gets workspace path with the specified username, returns "" if not found.
func GetUserWorkspace(username string) string {
	for _, user := range Users {
		if user.Name == username {
			return user.GetWorkspace()
		}
	}

	return ""
}

// GetGoFmt gets the path of Go format tool, returns "gofmt" if not found "goimports".
func GetGoFmt(username string) string {
	for _, user := range Users {
		if user.Name == username {
			switch user.GoFormat {
			case "gofmt":
				return "gofmt"
			case "goimports":
				return util.Go.GetExecutableInGOBIN("goimports")
			default:
				logger.Errorf("Unsupported Go Format tool [%s]", user.GoFormat)
				return "gofmt"
			}
		}
	}

	return "gofmt"
}

// GetUser gets configuration of the user specified by the given username, returns nil if not found.
func GetUser(username string) *User {
	if "playground" == username { // reserved user for Playground
		// mock it
		return NewUser("playground", "", "", "")
	}

	for _, user := range Users {
		if user.Name == username {
			return user
		}
	}

	return nil
}

// initCustomizedConfs initializes the user customized configurations.
func initCustomizedConfs() {
	for _, user := range Users {
		UpdateCustomizedConf(user.Name)
	}
}

// UpdateCustomizedConf creates (if not exists) or updates user customized configuration files.
//
//  1. /static/user/{username}/style.css
func UpdateCustomizedConf(username string) {
	var u *User
	for _, user := range Users { // maybe it is a beauty of the trade-off of the another world between design and implementation
		if user.Name == username {
			u = user
		}
	}

	if nil == u {
		return
	}

	model := map[string]interface{}{"user": u}

	t, err := template.ParseFiles("static/user/style.css.tmpl")
	if nil != err {
		logger.Error(err)

		os.Exit(-1)
	}

	wd := util.OS.Pwd()
	dir := filepath.Clean(wd + "/static/user/" + u.Name)
	if err := os.MkdirAll(dir, 0755); nil != err {
		logger.Error(err)

		os.Exit(-1)
	}

	fout, err := os.Create(dir + PathSeparator + "style.css")
	if nil != err {
		logger.Error(err)

		os.Exit(-1)
	}

	defer fout.Close()

	if err := t.Execute(fout, model); nil != err {
		logger.Error(err)

		os.Exit(-1)
	}
}

// initWorkspaceDirs initializes the directories of users' workspaces.
//
// Creates directories if not found on path of workspace.
func initWorkspaceDirs() {
	paths := []string{}

	for _, user := range Users {
		paths = append(paths, filepath.SplitList(user.GetWorkspace())...)
	}

	for _, path := range paths {
		CreateWorkspaceDir(path)
	}
}

// CreateWorkspaceDir creates (if not exists) directories on the path.
//
//  1. root directory:{path}
//  2. src directory: {path}/src
//  3. package directory: {path}/pkg
//  4. binary directory: {path}/bin
func CreateWorkspaceDir(path string) {
	createDir(path)
	createDir(path + PathSeparator + "src")
	createDir(path + PathSeparator + "pkg")
	createDir(path + PathSeparator + "bin")
}

// createDir creates a directory on the path if it not exists.
func createDir(path string) {
	if !util.File.IsExist(path) {
		if err := os.MkdirAll(path, 0775); nil != err {
			logger.Error(err)

			os.Exit(-1)
		}

		logger.Tracef("Created a dir [%s]", path)
	}
}

// GetEditorThemes gets the names of editor themes.
func GetEditorThemes() []string {
	ret := []string{}

	f, _ := os.Open("static/js/overwrite/codemirror" + "/theme")
	names, _ := f.Readdirnames(-1)
	f.Close()

	for _, name := range names {
		ret = append(ret, name[:strings.LastIndex(name, ".")])
	}

	sort.Strings(ret)

	return ret
}

// GetThemes gets the names of themes.
func GetThemes() []string {
	ret := []string{}

	f, _ := os.Open("static/css/themes")
	names, _ := f.Readdirnames(-1)
	f.Close()

	for _, name := range names {
		ret = append(ret, name[:strings.LastIndex(name, ".")])
	}

	sort.Strings(ret)

	return ret
}
