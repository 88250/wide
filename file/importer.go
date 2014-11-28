package file

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/b3log/wide/util"
)

type FileInfo struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Error string `json:"error,omitempty"`
}

func handleUpload(p *multipart.Part, dir string) (fi *FileInfo) {
	fi = &FileInfo{
		Name: p.FileName(),
		Type: p.Header.Get("Content-Type"),
	}

	path := filepath.Clean(dir + "/" + fi.Name)
	f, _ := os.Create(path)

	io.Copy(f, p)

	f.Close()

	return
}

func handleUploads(r *http.Request, dir string) (fileInfos []*FileInfo) {
	fileInfos = make([]*FileInfo, 0)
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

// Upload handles request of file upload.
func Upload(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	q := r.URL.Query()
	dir := q["path"][0]

	data["files"] = handleUploads(r, dir)
}
