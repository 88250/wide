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

// GoFmtHandler handles request of formatting Go source code.
//
// This function will select a format tooll based on user's configuration:
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
	if strings.HasPrefix(filePath, apiPath) { // if it is Go API source code
		// ignore it
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

// HTMLFmtHandler handles request of formatting HTML source code.
// FIXME: gohtml has some issues...
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
