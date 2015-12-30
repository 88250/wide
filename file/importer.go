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
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/b3log/wide/util"
)

type fileInfo struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

func handleUpload(p *multipart.Part, dir string) (fi *fileInfo) {
	fi = &fileInfo{
		Name: p.FileName(),
		Type: p.Header.Get("Content-Type"),
	}

	path := filepath.Clean(dir + "/" + fi.Name)
	f, _ := os.Create(path)

	io.Copy(f, p)

	f.Close()

	return
}

func handleUploads(r *http.Request, dir string) (fileInfos []*fileInfo) {
	fileInfos = make([]*fileInfo, 0)
	mr, err := r.MultipartReader()

	part, err := mr.NextPart()

	for err == nil {
		if name := part.FormName(); name != "" {
			if part.FileName() != "" {
				fileInfos = append(fileInfos, handleUpload(part, dir))
			}
		}

		part, err = mr.NextPart()
	}

	return
}

// UploadHandler handles request of file upload.
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	q := r.URL.Query()
	dir := q["path"][0]

	result.Data = handleUploads(r, dir)
}
