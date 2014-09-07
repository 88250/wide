package editor

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/88250/gohtml"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

func HTMLFmtHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	filePath := args["file"].(string)

	fout, err := os.Create(filePath)

	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	code := args["code"].(string)

	fout.WriteString(code)

	if err := fout.Close(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	output := gohtml.Format(code)

	if "" == output {
		data["succ"] = false

		return
	}

	data["code"] = string(output)
}
