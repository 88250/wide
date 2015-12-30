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
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
)

// SaveHandler handles request of Playground code save.
func SaveHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	session, _ := session.HTTPSession.Get(r, "wide-session")
	if session.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	code := args["code"].(string)

	// Step1. format code
	cmd := exec.Command("gofmt")

	stdin, err := cmd.StdinPipe()
	if nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	io.WriteString(stdin, code)
	stdin.Close()

	bytes, _ := cmd.Output()
	output := string(bytes)
	if "" != output {
		code = string(output)
	}

	data := map[string]interface{}{}
	result.Data = &data

	data["code"] = code

	// Step2. generate file name
	hasher := md5.New()
	hasher.Write([]byte(code))
	fileName := hex.EncodeToString(hasher.Sum(nil))
	fileName += ".go"
	data["fileName"] = fileName

	// Step3. write file
	filePath := filepath.Clean(conf.Wide.Playground + "/" + fileName)
	fout, err := os.Create(filePath)
	fout.WriteString(code)
	if err := fout.Close(); nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}
}

// ShortURLHandler handles request of short URL.
func ShortURLHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	session, _ := session.HTTPSession.Get(r, "wide-session")
	if session.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	url := args["url"].(string)

	resp, _ := http.Post("http://dwz.cn/create.php", "application/x-www-form-urlencoded",
		strings.NewReader("url="+url))

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	shortURL := url
	if 0 == response["status"].(float64) {
		shortURL = response["tinyurl"].(string)
	}

	result.Data = shortURL
}
