package file

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

// GetZip handles request of retrieving zip file.
func GetZip(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("Content-type", "application/zip")
	http.ServeFile(w, r, path)

	os.Remove(path)
}

// CreateZip handles request of creating zip.
func CreateZip(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	path := args["path"].(string)
	base := filepath.Base(path)

	if !util.File.IsExist(path) {
		data["succ"] = false
		data["msg"] = "Can't find file [" + path + "]"

		return
	}

	zipFile, err := util.Zip.Create(path + ".zip")
	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}
	defer zipFile.Close()

	if util.File.IsDir(path) {
		zipFile.AddDirectory(base, path)
	} else {
		zipFile.AddEntry(base, path)
	}
}
