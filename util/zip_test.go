// Copyright (c) 2014-2015, b3log.org
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
	"testing"
)

var packageName = "test_zip"

func TestCreate(t *testing.T) {
	zipFile, err := Zip.Create(packageName + ".zip")
	if nil != err {
		t.Error(err)

		return
	}

	zipFile.AddDirectoryN(".", ".")
	if nil != err {
		t.Error(err)

		return
	}

	err = zipFile.Close()
	if nil != err {
		t.Error(err)

		return
	}
}

func TestUnzip(t *testing.T) {
	err := Zip.Unzip(packageName+".zip", packageName)
	if nil != err {
		t.Error(err)

		return
	}
}

func TestMain(m *testing.M) {
	retCode := m.Run()

	// clean test data
	os.RemoveAll(packageName + ".zip")
	os.RemoveAll(packageName)

	os.Exit(retCode)
}
