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

package conf

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/b3log/wide/util"
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
	Name                 string
	Password             string
	Salt                 string
	Email                string
	Gravatar             string // see http://gravatar.com
	Workspace            string // the GOPATH of this user (maybe contain several paths splitted by os.PathListSeparator)
	Locale               string
	GoFormat             string
	FontFamily           string
	FontSize             string
	Theme                string
	Keymap               string // wide/vim
	Created              int64  // user create time in unix nano
	Updated              int64  // preference update time in unix nano
	Lived                int64  // the latest session activity in unix nano
	Editor               *editor
	LatestSessionContent *LatestSessionContent
}

// Editor configuration of a user.
type editor struct {
	FontFamily string
	FontSize   string
	LineHeight string
	Theme      string
	TabSize    string
}

// NewUser creates a user with the specified username, password, email and workspace.
func NewUser(username, password, email, workspace string) *User {
	md5hash := md5.New()
	md5hash.Write([]byte(email))
	gravatar := hex.EncodeToString(md5hash.Sum(nil))

	salt := util.Rand.String(16)
	password = Salt(password, salt)

	now := time.Now().UnixNano()

	return &User{Name: username, Password: password, Salt: salt, Email: email, Gravatar: gravatar, Workspace: workspace,
		Locale: Wide.Locale, GoFormat: "gofmt", FontFamily: "Helvetica", FontSize: "13px", Theme: "default",
		Keymap:  "wide",
		Created: now, Updated: now, Lived: now,
		Editor: &editor{FontFamily: "Consolas, 'Courier New', monospace", FontSize: "inherit", LineHeight: "17px",
			Theme: "wide", TabSize: "4"}}
}

// Save saves the user's configurations in conf/users/{username}.json.
func (u *User) Save() bool {
	bytes, err := json.MarshalIndent(u, "", "    ")

	if nil != err {
		logger.Error(err)

		return false
	}

	if "" == string(bytes) {
		logger.Error("Truncated user [" + u.Name + "]")

		return false
	}

	if err = ioutil.WriteFile("conf/users/"+u.Name+".json", bytes, 0644); nil != err {
		logger.Error(err)

		return false
	}

	return true
}

// GetWorkspace gets workspace path of the user.
//
// Compared to the use of Wide.Workspace, this function will be processed as follows:
//  1. Replace {WD} variable with the actual directory path
//  2. Replace ${GOPATH} with enviorment variable GOPATH
//  3. Replace "/" with "\\" (Windows)
func (u *User) GetWorkspace() string {
	w := strings.Replace(u.Workspace, "{WD}", Wide.WD, 1)
	w = strings.Replace(w, "${GOPATH}", os.Getenv("GOPATH"), 1)

	return filepath.FromSlash(w)
}

// GetOwner gets the user the specified path belongs to. Returns "" if not found.
func GetOwner(path string) string {
	for _, user := range Users {
		if strings.HasPrefix(path, user.GetWorkspace()) {
			return user.Name
		}
	}

	return ""
}

// Salt salts the specified password with the specified salt.
func Salt(password, salt string) string {
	sha1hash := sha1.New()
	sha1hash.Write([]byte(password + salt))

	return hex.EncodeToString(sha1hash.Sum(nil))
}
