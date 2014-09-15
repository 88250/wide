package editor

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"

	"github.com/88250/gohtml"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

// gofmt 格式化 Go 源码文件.
func GoFmtHandler(w http.ResponseWriter, r *http.Request) {
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

	argv := []string{filePath}
	cmd := exec.Command("gofmt", argv...)

	bytes, _ := cmd.Output()
	output := string(bytes)
	if "" == output {
		data["succ"] = false

		return
	}

	code = string(output)
	data["code"] = code

	fout, err = os.Create(filePath)
	fout.WriteString(code)
	if err := fout.Close(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}
}

// 格式化 HTML 文件.
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

	code = string(output)
	data["code"] = code

	fout, err = os.Create(filePath)
	fout.WriteString(code)
	if err := fout.Close(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}
}

// 格式化 JSON 文件.
func JSONFmtHandler(w http.ResponseWriter, r *http.Request) {
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

	obj := new(interface{})
	if err := json.Unmarshal([]byte(code), &obj); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	glog.Info(obj)

	bytes, err := json.MarshalIndent(obj, "", "    ")
	if nil != err {
		data["succ"] = false

		return
	}

	code = string(bytes)
	data["code"] = code

	fout, err = os.Create(filePath)
	fout.WriteString(code)
	if err := fout.Close(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}
}
