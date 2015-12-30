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
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestGetCrossPlatforms(t *testing.T) {
	crossPlatforms := Go.GetCrossPlatforms()

	if len(crossPlatforms) < 1 {
		t.Error("should have one platform at least")
	}
}

func TestGetAPIPath(t *testing.T) {
	apiPath := Go.GetAPIPath()

	v := runtime.Version()[2:]
	v = v[:3]

	verNum, err := strconv.ParseFloat(v, 64)

	if nil != err {
		t.Error(err)

		return
	}

	if verNum >= 1.4 {
		if !strings.HasSuffix(apiPath, "src") {
			t.Error("api path should end with \"src\"")

			return
		}
	} else {
		if !strings.HasSuffix(apiPath, "pkg") {
			t.Error("api path should end with \"pkg\"")
		}
	}
}

func TestIsAPI(t *testing.T) {
	apiPath := Go.GetAPIPath()

	if !Go.IsAPI(apiPath) {
		t.Error("api path root should belong to api path")

		return
	}

	root := "/root"

	if Go.IsAPI(root) {
		t.Error("root should not belong to api path")

		return
	}
}

func TestGetGoFormats(t *testing.T) {
	formats := Go.GetGoFormats()

	if len(formats) < 1 {
		t.Error("should have one go format tool [gofmt] at least")
	}
}

func TestGetExecutableInGOBIN(t *testing.T) {
	bin := Go.GetExecutableInGOBIN("test")

	if OS.IsWindows() {
		if !strings.HasSuffix(bin, ".exe") {
			t.Error("Executable binary should end with .exe")

			return
		}
	}
}
