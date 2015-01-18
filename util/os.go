// Copyright (c) 2015, b3log.org
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
	"os/exec"
	"path/filepath"
	"runtime"
)

type myos struct{}

// OS utilities.
var OS = myos{}

// IsWindows determines whether current OS is Windows.
func (*myos) IsWindows() bool {
	return "windows" == runtime.GOOS
}

// Pwd gets the path of current working directory.
func (*myos) Pwd() string {
	file, _ := exec.LookPath(os.Args[0])
	pwd, _ := filepath.Abs(file)

	return filepath.Dir(pwd)
}
