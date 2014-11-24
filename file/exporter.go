package file

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

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
