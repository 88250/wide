// Copyright (c) 2014, B3log
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
	"path/filepath"
	"path"
	"runtime"
	"strings"
)

type mygo struct{}

// Go utilities.
var Go = mygo{}

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

	return strings.HasPrefix(path, apiPath)
}
