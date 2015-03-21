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

package file

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/b3log/wide/util"
)

// DecompressHandler handles request of decompressing zip/tar.gz.
func DecompressHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		data["succ"] = false

		return
	}

	path := args["path"].(string)
	//	base := filepath.Base(path)
	dir := filepath.Dir(path)

	if !util.File.IsExist(path) {
		data["succ"] = false
		data["msg"] = "Can't find file [" + path + "] to descompress"

		return
	}

	err := util.Zip.Unzip(path, dir)
	if nil != err {
		logger.Error(err)
		data["succ"] = false

		return
	}

}
