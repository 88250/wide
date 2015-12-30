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

package editor

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
)

// GoFmtHandler handles request of formatting Go source code.
//
// This function will select a format tooll based on user's configuration:
//  1. gofmt
//  2. goimports
func GoFmtHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	session, _ := session.HTTPSession.Get(r, "wide-session")
	if session.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := session.Values["username"].(string)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	filePath := args["file"].(string)

	if util.Go.IsAPI(filePath) {
		result.Succ = false

		return
	}

	fout, err := os.Create(filePath)

	if nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	code := args["code"].(string)

	fout.WriteString(code)
	if err := fout.Close(); nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	data := map[string]interface{}{}
	result.Data = &data

	data["code"] = code

	result.Data = data

	fmt := conf.GetGoFmt(username)

	argv := []string{filePath}
	cmd := exec.Command(fmt, argv...)

	bytes, _ := cmd.Output()
	output := string(bytes)
	if "" == output {
		// format error, returns the original content
		result.Succ = true

		return
	}

	code = string(output)
	data["code"] = code

	fout, err = os.Create(filePath)
	fout.WriteString(code)
	if err := fout.Close(); nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}
}
