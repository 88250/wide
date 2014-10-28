package editor

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/88250/gohtml"
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

// 格式化 Go 源码文件.
// 根据用户的 GoFormat 配置选择格式化工具：
//  1. gofmt
//  2. goimports
func GoFmtHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	session, _ := session.HTTPSession.Get(r, "wide-session")
	username := session.Values["username"].(string)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	filePath := args["file"].(string)

	apiPath := runtime.GOROOT() + conf.PathSeparator + "src" + conf.PathSeparator + "pkg"
	if strings.HasPrefix(filePath, apiPath) { // 如果是 Go API 源码文件
		// 忽略修改
		return
	}

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

	fmt := conf.Wide.GetGoFmt(username)

	argv := []string{filePath}
	cmd := exec.Command(fmt, argv...)

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
// FIXME：依赖的工具 gohtml 格式化 HTML 时有问题
func HTMLFmtHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
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

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
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
