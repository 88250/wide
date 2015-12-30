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

package file

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/b3log/wide/util"
)

// GetZipHandler handles request of retrieving zip file.
func GetZipHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	path := q["path"][0]

	if ".zip" != filepath.Ext(path) {
		http.Error(w, "Bad Request", 400)

		return
	}

	if !util.File.IsExist(path) {
		http.Error(w, "Not Found", 404)

		return
	}

	filename := filepath.Base(path)

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/zip")
	http.ServeFile(w, r, path)

	os.Remove(path)
}

// CreateZipHandler handles request of creating zip.
func CreateZipHandler(w http.ResponseWriter, r *http.Request) {
	data := util.NewResult()
	defer util.RetResult(w, r, data)

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		data.Succ = false

		return
	}

	path := args["path"].(string)
	var name string

	base := filepath.Base(path)

	if nil != args["name"] {
		name = args["name"].(string)
	} else {
		name = base
	}

	dir := filepath.Dir(path)

	if !util.File.IsExist(path) {
		data.Succ = false
		data.Msg = "Can't find file [" + path + "]"

		return
	}

	zipPath := filepath.Join(dir, name)
	zipFile, err := util.Zip.Create(zipPath + ".zip")
	if nil != err {
		logger.Error(err)
		data.Succ = false

		return
	}
	defer zipFile.Close()

	if util.File.IsDir(path) {
		zipFile.AddDirectory(base, path)
	} else {
		zipFile.AddEntry(base, path)
	}

	data.Data = zipPath
}
