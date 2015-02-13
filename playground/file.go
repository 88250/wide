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

package playground

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"github.com/b3log/wide/conf"
)

// SaveHandler handles request of Playground code save.
func SaveHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	session, _ := session.HTTPSession.Get(r, "wide-session")
	if session.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		data["succ"] = false

		return
	}

	code := args["code"].(string)

	// generate file name
	hasher := md5.New()
	hasher.Write([]byte(code))
	fileName := hex.EncodeToString(hasher.Sum(nil))
	fileName += ".go"
	filePath := filepath.Clean(conf.Wide.Playground + "/" + fileName)

	fout, err := os.Create(filePath)
	if nil != err {
		logger.Error(err)
		data["succ"] = false

		return
	}

	fout.WriteString(code)
	if err := fout.Close(); nil != err {
		logger.Error(err)
		data["succ"] = false

		return
	}

	data["filePath"] = filePath
	data["url"] = filepath.ToSlash(filePath)

	argv := []string{filePath}
	cmd := exec.Command("gofmt", argv...)

	bytes, _ := cmd.Output()
	output := string(bytes)
	if "" == output {
		// format error, returns the original content
		data["succ"] = true
		data["code"] = code

		return
	}

	code = string(output)
	data["code"] = code

	// generate file name
	hasher = md5.New()
	hasher.Write([]byte(code))
	fileName = hex.EncodeToString(hasher.Sum(nil))
	fileName += ".go"
	filePath = filepath.Clean(conf.Wide.Playground + "/" +  fileName)
	data["filePath"] = filePath
	data["url"] = filepath.ToSlash(filePath)

	fout, err = os.Create(filePath)
	fout.WriteString(code)
	if err := fout.Close(); nil != err {
		logger.Error(err)
		data["succ"] = false

		return
	}

}
