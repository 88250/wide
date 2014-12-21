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

package log

import (
	"os"
	"testing"
)

var logger = NewLogger(os.Stdout)

func TestSetLevel(t *testing.T) {
	SetLevel("trace")
}

func TestTrace(t *testing.T) {
	logger.Trace("trace")
}

func TestTracef(t *testing.T) {
	logger.Tracef("tracef")
}

func TestInfo(t *testing.T) {
	logger.Info("info")
}

func TestInfof(t *testing.T) {
	logger.Infof("infof")
}

func TestWarn(t *testing.T) {
	logger.Warn("warn")
}

func TestWarnf(t *testing.T) {
	logger.Warnf("warnf")
}

func TestError(t *testing.T) {
	logger.Error("error")
}

func TestErrorf(t *testing.T) {
	logger.Errorf("errorf")
}

func TestGetLevel(t *testing.T) {
	l := getLevel("info")
	logger.Info(l)
}

func TestLoggerSetLevel(t *testing.T) {
	if logger.level != Trace {
		t.FailNow()

		return
	}

	logger.SetLevel("info")

	if logger.level != Info {
		t.FailNow()

		return
	}
}

func TestIsTraceEnabled(t *testing.T) {
	logger.SetLevel("trace")

	if !logger.IsTraceEnabled() {
		t.FailNow()

		return
	}
}

func TestIsDebugEnabled(t *testing.T) {
	logger.SetLevel("debug")

	if !logger.IsDebugEnabled() {
		t.FailNow()

		return
	}
}
