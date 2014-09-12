// 编辑器操作.
package editor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/user"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

var editorWS = map[string]*websocket.Conn{}

// 代码片段. 这个结构可用于“查找使用”、“文件搜索”的返回值.
type snippet struct {
	Path     string   `json:"path"`     // 文件路径
	Line     int      `json:"lline"`    // 行号
	Ch       int      `json:"ch"`       // 列号
	Contents []string `json:"contents"` // 代码行
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := user.Session.Get(r, "wide-session")
	sid := session.Values["id"].(string)

	editorWS[sid], _ = websocket.Upgrade(w, r, nil, 1024, 1024)

	ret := map[string]interface{}{"output": "Editor initialized", "cmd": "init-editor"}
	editorWS[sid].WriteJSON(&ret)

	glog.Infof("Open a new [Editor] with session [%s], %d", sid, len(editorWS))

	args := map[string]interface{}{}
	for {
		if err := editorWS[sid].ReadJSON(&args); err != nil {
			if err.Error() == "EOF" {
				return
			}

			if err.Error() == "unexpected EOF" {
				return
			}

			glog.Error("Editor WS ERROR: " + err.Error())
			return
		}

		code := args["code"].(string)
		line := int(args["cursorLine"].(float64))
		ch := int(args["cursorCh"].(float64))

		offset := getCursorOffset(code, line, ch)

		// glog.Infof("offset: %d", offset)

		argv := []string{"-f=json", "autocomplete", strconv.Itoa(offset)}

		var output bytes.Buffer

		cmd := exec.Command("gocode", argv...)
		cmd.Stdout = &output

		stdin, _ := cmd.StdinPipe()
		cmd.Start()
		stdin.Write([]byte(code))
		stdin.Close()
		cmd.Wait()

		ret = map[string]interface{}{"output": string(output.Bytes()), "cmd": "autocomplete"}

		if err := editorWS[sid].WriteJSON(&ret); err != nil {
			glog.Error("Editor WS ERROR: " + err.Error())
			return
		}
	}
}

func AutocompleteHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	session, _ := user.Session.Get(r, "wide-session")
	username := session.Values["username"].(string)

	code := args["code"].(string)
	line := int(args["cursorLine"].(float64))
	ch := int(args["cursorCh"].(float64))

	offset := getCursorOffset(code, line, ch)

	// glog.Infof("offset: %d", offset)

	userWorkspace := conf.Wide.GetUserWorkspace(username)

	//glog.Infof("User [%s] workspace [%s]", username, userWorkspace)
	userLib := userWorkspace + string(os.PathSeparator) + "pkg" + string(os.PathSeparator) +
		runtime.GOOS + "_" + runtime.GOARCH

	masterWorkspace := conf.Wide.Workspace
	//glog.Infof("Master workspace [%s]", masterWorkspace)
	masterLib := masterWorkspace + string(os.PathSeparator) + "pkg" + string(os.PathSeparator) +
		runtime.GOOS + "_" + runtime.GOARCH

	libPath := userLib + string(os.PathListSeparator) + masterLib
	//glog.Infof("gocode set lib-path %s", libPath)

	// FIXME: 使用 gocode set lib-path 在多工作空间环境下肯定是有问题的，需要考虑其他实现方式
	argv := []string{"set", "lib-path", libPath}
	cmd := exec.Command("gocode", argv...)
	cmd.Start()

	//gocode 试验性质特性：自动构建
	//argv = []string{"set", "autobuild", "true"}
	//cmd := exec.Command("gocode", argv...)
	//cmd.Start()

	argv = []string{"-f=json", "autocomplete", strconv.Itoa(offset)}
	cmd = exec.Command("gocode", argv...)

	stdin, _ := cmd.StdinPipe()
	stdin.Write([]byte(code))
	stdin.Close()

	output, err := cmd.CombinedOutput()
	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

// 查找声明.
func FindDeclarationHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	session, _ := user.Session.Get(r, "wide-session")
	username := session.Values["username"].(string)

	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	filePath := args["file"].(string)
	curDir := filePath[:strings.LastIndex(filePath, string(os.PathSeparator))]
	filename := filePath[strings.LastIndex(filePath, string(os.PathSeparator))+1:]

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

	line := int(args["cursorLine"].(float64))
	ch := int(args["cursorCh"].(float64))

	offset := getCursorOffset(code, line, ch)

	// TODO: 目前是调用 liteide_stub 工具来查找声明，后续需要重新实现
	argv := []string{"type", "-cursor", filename + ":" + strconv.Itoa(offset), "-def", "."}
	cmd := exec.Command("liteide_stub", argv...)
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

	output, err := cmd.CombinedOutput()
	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	found := strings.TrimSpace(string(output))
	if "" == found {
		data["succ"] = false

		return
	}

	part := found[:strings.LastIndex(found, ":")]
	cursorSep := strings.LastIndex(part, ":")
	path := found[:cursorSep]
	cursorLine := found[cursorSep+1 : strings.LastIndex(found, ":")]
	cursorCh := found[strings.LastIndex(found, ":")+1:]

	// glog.Infof("%s\n%s\n%s\n%s", found, path, cursorLine, cursorCh)

	data["path"] = path
	data["cursorLine"] = cursorLine
	data["cursorCh"] = cursorCh
}

// 查找使用.
func FindUsagesHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	session, _ := user.Session.Get(r, "wide-session")
	username := session.Values["username"].(string)

	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	filePath := args["file"].(string)
	curDir := filePath[:strings.LastIndex(filePath, string(os.PathSeparator))]
	filename := filePath[strings.LastIndex(filePath, string(os.PathSeparator))+1:]

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

	line := int(args["cursorLine"].(float64))
	ch := int(args["cursorCh"].(float64))

	offset := getCursorOffset(code, line, ch)

	// TODO: 目前是调用 liteide_stub 工具来查找使用，后续需要重新实现
	argv := []string{"type", "-cursor", filename + ":" + strconv.Itoa(offset), "-use", "."}
	cmd := exec.Command("liteide_stub", argv...)
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

	output, err := cmd.CombinedOutput()
	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	result := strings.TrimSpace(string(output))
	if "" == result {
		data["succ"] = false

		return
	}

	founds := strings.Split(result, "\n")
	usages := []snippet{}
	for _, found := range founds {
		found = strings.TrimSpace(found)

		part := found[:strings.LastIndex(found, ":")]
		cursorSep := strings.LastIndex(part, ":")
		path := found[:cursorSep]
		cursorLine, _ := strconv.Atoi(found[cursorSep+1 : strings.LastIndex(found, ":")])
		cursorCh, _ := strconv.Atoi(found[strings.LastIndex(found, ":")+1:])

		usage := snippet{Path: path, Line: cursorLine, Ch: cursorCh /* TODO: 获取附近的代码片段 */}
		usages = append(usages, usage)
	}

	// glog.Infof("%s\n%s\n%s\n%s", found, path, cursorLine, cursorCh)

	data["usages"] = usages
}

func getCursorOffset(code string, line, ch int) (offset int) {
	lines := strings.Split(code, "\n")

	for i := 0; i < line; i++ {
		offset += len(lines[i])
	}

	offset += line + ch

	return
}

func setCmdEnv(cmd *exec.Cmd, username string) {
	userWorkspace := conf.Wide.GetUserWorkspace(username)

	cmd.Env = append(cmd.Env,
		"GOPATH="+userWorkspace,
		"GOOS="+runtime.GOOS,
		"GOARCH="+runtime.GOARCH,
		"GOROOT="+runtime.GOROOT(),
		"PATH="+os.Getenv("PATH"))
}
