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

package util

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const (
	pathSeparator     = string(os.PathSeparator)     // OS-specific path separator
	pathListSeparator = string(os.PathListSeparator) // OS-specific path list separator
)

type mygo struct{}

// Go utilities.
var Go = mygo{}

func (*mygo) GetCrossPlatforms() []string {
	ret := []string{}

	toolDir := runtime.GOROOT() + "/pkg/tool"
	f, _ := os.Open(toolDir)
	names, _ := f.Readdirnames(-1)
	f.Close()

	for _, name := range names {
		subDir, _ := os.Open(toolDir + "/" + name)
		tools, _ := subDir.Readdirnames(10)
		subDir.Close()

		if len(tools) > 5 {
			ret = append(ret, name)
		}
	}

	return ret
}

// GetAPIPath gets the Go source code path.
//
//  1. before Go 1.4: $GOROOT/src/pkg
//  2. Go 1.4 and after: $GOROOT/src
func (*mygo) GetAPIPath() string {
	ret := runtime.GOROOT() + "/src/pkg" // before Go 1.4
	if !File.IsExist(ret) {
		ret = runtime.GOROOT() + "/src" // Go 1.4 and after
	}

	return filepath.FromSlash(path.Clean(ret))
}

// IsAPI determines whether the specified path belongs to Go API.
func (*mygo) IsAPI(path string) bool {
	apiPath := Go.GetAPIPath()

	return strings.HasPrefix(filepath.FromSlash(path), apiPath)
}

// GetGoFormats gets Go format tools. It may return ["gofmt", "goimports"].
func (*mygo) GetGoFormats() []string {
	ret := []string{"gofmt"}

	p := Go.GetExecutableInGOBIN("goimports")
	if File.IsExist(p) {
		ret = append(ret, "goimports")
	}

	sort.Strings(ret)

	return ret
}

// GetExecutableInGOBIN gets executable file under GOBIN path.
//
// The specified executable should not with extension, this function will append .exe if on Windows.
func (*mygo) GetExecutableInGOBIN(executable string) string {
	if OS.IsWindows() {
		executable += ".exe"
	}

	gopaths := filepath.SplitList(os.Getenv("GOPATH"))

	for _, gopath := range gopaths {
		// $GOPATH/bin/$GOOS_$GOARCH/executable
		ret := gopath + pathSeparator + "bin" + pathSeparator +
			os.Getenv("GOOS") + "_" + os.Getenv("GOARCH") + pathSeparator + executable
		if File.IsExist(ret) {
			return ret
		}

		// $GOPATH/bin/{runtime.GOOS}_{runtime.GOARCH}/executable
		ret = gopath + pathSeparator + "bin" + pathSeparator +
			runtime.GOOS + "_" + runtime.GOARCH + pathSeparator + executable
		if File.IsExist(ret) {
			return ret
		}

		// $GOPATH/bin/executable
		ret = gopath + pathSeparator + "bin" + pathSeparator + executable
		if File.IsExist(ret) {
			return ret
		}
	}

	// $GOBIN/executable
	gobin := os.Getenv("GOBIN")
	if "" != gobin {
		ret := gobin + pathSeparator + executable
		if File.IsExist(ret) {
			return ret
		}
	}

	return "./" + executable
}
