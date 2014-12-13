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
	"os"
	"runtime/debug"

	"github.com/b3log/wide/log"
)

// Logger.
var logger = log.NewLogger(os.Stdout)

// Recover recovers a panic.
func Recover() {
	if re := recover(); nil != re {
		logger.Errorf("PANIC RECOVERED:\n %v, %s", re, debug.Stack())
	}
}
