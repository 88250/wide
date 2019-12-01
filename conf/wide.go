// Copyright (c) 2014-present, b3gulu.Log.org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
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
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/88250/gulu"
	"github.com/88250/wide/event"
)

const (
	// PathSeparator holds the OS-specific path separator.
	PathSeparator = string(os.PathSeparator)
	// PathListSeparator holds the OS-specific path list separator.
	PathListSeparator = string(os.PathListSeparator)

	// WideVersion holds the current Wide's version.
	WideVersion = "1.6.0"
	// CodeMirrorVer holds the current editor version.
	CodeMirrorVer = "5.1"
	// UserAgent represents HTTP client user agent.
	UserAgent = "Wide/" + WideVersion + "; +https://github.com/b3log/wide"

	HelloWorld = `package main

import "fmt"

func main() {
	fmt.Println("欢迎通过《边看边练 Go 系列》来学习 Go 语言 https://hacpai.com/article/1437497122181")
}
`
)

// Configuration.
type conf struct {
	Server                string        // server
	LogLevel              string        // logging level: trace/debug/info/warn/error
	Data                  string        // data directory
	RuntimeMode           string        // runtime mode (dev/prod)
	HTTPSessionMaxAge     int           // HTTP session max age (in seciond)
	StaticResourceVersion string        // version of static resources
	Locale                string        // default locale
	Autocomplete          bool          // default autocomplete
	SiteStatCode          template.HTML // site statistic code
}

// Logger.
var logger = gulu.Log.NewLogger(os.Stdout)

// Wide configurations.
var Wide *conf

// configurations of users.
var Users []*User

// Indicates whether Docker is available.
var Docker bool

// Docker image to run user's program
const DockerImageGo = "golang"

// Load loads the Wide configurations from wide.json and users' configurations from users/{userId}.json.
func Load(confPath, confData, confServer, confLogLevel string, confSiteStatCode template.HTML) {
	initWide(confPath, confData, confServer, confLogLevel, confSiteStatCode)
	initUsers()

	cmd := exec.Command("docker", "version")
	_, err := cmd.CombinedOutput()
	if nil != err {
		if !gulu.OS.IsWindows() {
			logger.Errorf("Not found 'docker' installed, running user's code will cause security problem")

			os.Exit(-1)
		}
	} else {
		Docker = true
	}
}

func initUsers() {
	f, err := os.Open(Wide.Data + PathSeparator + "users")
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

		bytes, _ := ioutil.ReadFile(filepath.Join(Wide.Data, "users", name))
		err := json.Unmarshal(bytes, user)
		if err != nil {
			logger.Errorf("Parses [%s] error: %v, skip loading this user", name, err)

			continue
		}

		// Compatibility upgrade (1.3.0): https://github.com/b3log/wide/issues/83
		if "" == user.Keymap {
			user.Keymap = "wide"
		}

		// Compatibility upgrade (1.5.3): https://github.com/b3log/wide/issues/308
		if "" == user.GoBuildArgsForLinux {
			user.GoBuildArgsForLinux = "-i"
		}
		if "" == user.GoBuildArgsForWindows {
			user.GoBuildArgsForWindows = "-i"
		}
		if "" == user.GoBuildArgsForDarwin {
			user.GoBuildArgsForDarwin = "-i"
		}

		Users = append(Users, user)
	}

	initWorkspaceDirs()
	initCustomizedConfs()
}

func initWide(confPath, confData, confServer, confLogLevel string, confSiteStatCode template.HTML) {
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

	Wide.Autocomplete = true // default to true

	// Logging Level
	gulu.Log.SetLevel(Wide.LogLevel)
	if "" != confLogLevel {
		Wide.LogLevel = confLogLevel
		gulu.Log.SetLevel(confLogLevel)
	}

	logger.Debug("Conf: \n" + string(bytes))

	// User Home
	home, err := gulu.OS.Home()
	if nil != err {
		logger.Error("Can't get user's home, please report this issue to developer", err)

		os.Exit(-1)
	}
	logger.Debugf("${user.home} [%s]", home)

	// Data directory
	if "" != confData {
		Wide.Data = confData
	}
	Wide.Data = strings.Replace(Wide.Data, "${home}", home, -1)
	Wide.Data = filepath.Clean(Wide.Data)
	if err := os.MkdirAll(Wide.Data+"/playground/", 0755); nil != err {
		logger.Errorf("Create data directory [%s] error", err)

		os.Exit(-1)
	}
	if err := os.MkdirAll(Wide.Data+"/users/", 0755); nil != err {
		logger.Errorf("Create data directory [%s] error", err)

		os.Exit(-1)
	}
	if err := os.MkdirAll(Wide.Data+"/workspaces/", 0755); nil != err {
		logger.Errorf("Create data directory [%s] error", err)

		os.Exit(-1)
	}

	// Server
	if "" != confServer {
		Wide.Server = confServer
	}

	// SiteStatCode
	if "" != confSiteStatCode {
		Wide.SiteStatCode = confSiteStatCode
	}

	time := strconv.FormatInt(time.Now().UnixNano(), 10)
	logger.Debugf("${time} [%s]", time)
	Wide.StaticResourceVersion = strings.Replace(Wide.StaticResourceVersion, "${time}", time, 1)
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
	defer gulu.Panic.Recover(nil)

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

	gocode := gulu.Go.GetExecutableInGOBIN("gocode")
	cmd = exec.Command(gocode)
	_, err = cmd.Output()
	if nil != err {
		event.EventQueue <- &event.Event{Code: event.EvtCodeGocodeNotFound}

		logger.Warnf("Not found gocode [%s], please install it with this command: go get github.com/nsf/gocode", gocode)
	}

	ideStub := gulu.Go.GetExecutableInGOBIN("gotools")
	cmd = exec.Command(ideStub, "version")
	_, err = cmd.Output()
	if nil != err {
		event.EventQueue <- &event.Event{Code: event.EvtCodeIDEStubNotFound}

		logger.Warnf("Not found gotools [%s], please install it with this command: go get github.com/visualfc/gotools", ideStub)
	}
}

// GetUserWorkspace gets workspace path with the specified user id, returns "" if not found.
func GetUserWorkspace(userId string) string {
	for _, user := range Users {
		if user.Id == userId {
			return user.WorkspacePath()
		}
	}

	return ""
}

// GetGoFmt gets the path of Go format tool, returns "gofmt" if not found "goimports".
func GetGoFmt(userId string) string {
	for _, user := range Users {
		if user.Id == userId {
			switch user.GoFormat {
			case "gofmt":
				return "gofmt"
			case "goimports":
				return gulu.Go.GetExecutableInGOBIN("goimports")
			default:
				logger.Errorf("Unsupported Go Format tool [%s]", user.GoFormat)
				return "gofmt"
			}
		}
	}

	return "gofmt"
}

// GetUser gets configuration of the user specified by the given user id, returns nil if not found.
func GetUser(id string) *User {
	if "playground" == id { // reserved user for Playground
		return NewUser("playground", "playground", "", "")
	}

	for _, user := range Users {
		if user.Id == id {
			return user
		}
	}

	return nil
}

// initCustomizedConfs initializes the user customized configurations.
func initCustomizedConfs() {
	for _, user := range Users {
		UpdateCustomizedConf(user.Id)
	}
}

// UpdateCustomizedConf creates (if not exists) or updates user customized configuration files.
//
//  1. /static/users/{userId}/style.css
func UpdateCustomizedConf(userId string) {
	var u *User
	for _, user := range Users { // maybe it is a beauty of the trade-off of the another world between design and implementation
		if user.Id == userId {
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

	dir := filepath.Clean(Wide.Data + "/static/users/" + userId)
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
		paths = append(paths, filepath.SplitList(user.WorkspacePath())...)
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
	if !gulu.File.IsExist(path) {
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
