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

package playground

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
)

// BuildHandler handles request of Playground building.
func BuildHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	fileName := args["fileName"].(string)
	filePath := filepath.Clean(conf.Wide.Playground + "/" + fileName)

	suffix := ""
	if util.OS.IsWindows() {
		suffix = ".exe"
	}

	data := map[string]interface{}{}
	result.Data = &data

	executable := filepath.Clean(conf.Wide.Playground + "/" + strings.Replace(fileName, ".go", suffix, -1))

	cmd := exec.Command("go", "build", "-o", executable, filePath)
	out, err := cmd.CombinedOutput()

	data["output"] = template.HTML(string(out))

	if nil != err {
		result.Succ = false

		return
	}

	data["executable"] = executable
}
