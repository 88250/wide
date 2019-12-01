// Copyright (c) 2014-present, b3log.org
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

package editor

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"

	"github.com/88250/gulu"
	"github.com/88250/wide/conf"
	"github.com/88250/wide/session"
)

// GoFmtHandler handles request of formatting Go source code.
//
// This function will select a format tooll based on user's configuration:
//  1. gofmt
//  2. goimports
func GoFmtHandler(w http.ResponseWriter, r *http.Request) {
	result := gulu.Ret.NewResult()
	defer gulu.Ret.RetResult(w, r, result)

	session, _ := session.HTTPSession.Get(r, session.CookieName)
	if session.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	uid := session.Values["uid"].(string)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Code = -1

		return
	}

	filePath := args["file"].(string)

	if gulu.Go.IsAPI(filePath) {
		result.Code = -1

		return
	}

	fout, err := os.Create(filePath)

	if nil != err {
		logger.Error(err)
		result.Code = -1

		return
	}

	code := args["code"].(string)

	fout.WriteString(code)
	if err := fout.Close(); nil != err {
		logger.Error(err)
		result.Code = -1

		return
	}

	data := map[string]interface{}{}
	result.Data = &data

	data["code"] = code

	result.Data = data

	fmt := conf.GetGoFmt(uid)

	argv := []string{filePath}
	cmd := exec.Command(fmt, argv...)

	bytes, _ := cmd.Output()
	output := string(bytes)
	if "" == output {
		// format error, returns the original content
		result.Code = 0

		return
	}

	code = string(output)
	data["code"] = code

	fout, err = os.Create(filePath)
	fout.WriteString(code)
	if err := fout.Close(); nil != err {
		logger.Error(err)
		result.Code = -1

		return
	}
}
