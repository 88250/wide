// Copyright (c) 2014-2019, b3log.org & hacpai.com
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

package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Panel represents a UI panel.
type Panel struct {
	State string `json:"state"` // panel state, "min"/"max"/"normal"
	Size  uint16 `json:"size"`  // panel size
}

// Layout represents the UI layout.
type Layout struct {
	Side      *Panel `json:"side"`      // Side panel
	SideRight *Panel `json:"sideRight"` // Right-Side panel
	Bottom    *Panel `json:"bottom"`    // Bottom panel
}

// LatestSessionContent represents the latest session content.
type LatestSessionContent struct {
	FileTree    []string `json:"fileTree"`    // paths of expanding nodes of file tree
	Files       []string `json:"files"`       // paths of files of opening editor tabs
	CurrentFile string   `json:"currentFile"` // path of file of the current focused editor tab
	Layout      *Layout  `json:"layout"`      // UI Layout
}

// User configuration.
type User struct {
	Id                    string
	Name                  string
	Avatar                string
	Workspace             string // the GOPATH of this user (maybe contain several paths splitted by os.PathListSeparator)
	Locale                string
	GoFormat              string
	GoBuildArgsForLinux   string
	GoBuildArgsForWindows string
	GoBuildArgsForDarwin  string
	FontFamily            string
	FontSize              string
	Theme                 string
	Keymap                string // wide/vim
	Created               int64  // user create time in unix nano
	Updated               int64  // preference update time in unix nano
	Lived                 int64  // the latest session activity in unix nano
	Editor                *editor
	LatestSessionContent  *LatestSessionContent
}

// Editor configuration of a user.
type editor struct {
	FontFamily string
	FontSize   string
	LineHeight string
	Theme      string
	TabSize    string
}

// Save saves the user's configurations in conf/users/{userId}.json.
func (u *User) Save() bool {
	bytes, err := json.MarshalIndent(u, "", "    ")

	if nil != err {
		logger.Error(err)

		return false
	}

	if "" == string(bytes) {
		logger.Error("Truncated user [" + u.Id + "]")

		return false
	}

	if err = ioutil.WriteFile(filepath.Join(Wide.Data, "users", u.Id+".json"), bytes, 0644); nil != err {
		logger.Error(err)

		return false
	}

	return true
}

// NewUser creates a user with the specified username and workspace.
func NewUser(id, name, avatar, workspace string) *User {
	now := time.Now().UnixNano()

	return &User{Id: id, Name: name, Avatar: avatar, Workspace: workspace,
		Locale: Wide.Locale, GoFormat: "gofmt",
		GoBuildArgsForLinux: "-i", GoBuildArgsForWindows: "-i", GoBuildArgsForDarwin: "-i",
		FontFamily: "Helvetica", FontSize: "13px", Theme: "default",
		Keymap:  "wide",
		Created: now, Updated: now, Lived: now,
		Editor: &editor{FontFamily: "Consolas, 'Courier New', monospace", FontSize: "inherit", LineHeight: "17px",
			Theme: "wide", TabSize: "4"}}
}

// WorkspacePath gets workspace path of the user.
//
// Compared to the use of Wide.Workspace, this function will be processed as follows:
//  1. Replace {WD} variable with the actual directory path
//  2. Replace ${GOPATH} with enviorment variable GOPATH
//  3. Replace "/" with "\\" (Windows)
func (u *User) WorkspacePath() string {
	w := u.Workspace
	w = strings.Replace(w, "${GOPATH}", os.Getenv("GOPATH"), 1)

	return filepath.FromSlash(w)
}

// BuildArgs get build args with the specified os.
func (u *User) BuildArgs(os string) []string {
	var tmp string
	if os == "windows" {
		tmp = u.GoBuildArgsForWindows
	}
	if os == "linux" {
		tmp = u.GoBuildArgsForLinux
	}
	if os == "darwin" {
		tmp = u.GoBuildArgsForDarwin
	}

	exp := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)'`)
	ret := exp.FindAllString(tmp, -1)
	for idx := range ret {
		ret[idx] = strings.Replace(ret[idx], "\"", "", -1)
	}

	return ret
}

// GetOwner gets the user the specified path belongs to. Returns "" if not found.
func GetOwner(path string) string {
	for _, user := range Users {
		if strings.HasPrefix(path, user.WorkspacePath()) {
			return user.Id
		}
	}

	return ""
}
