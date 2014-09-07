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
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

var editorWS = map[string]*websocket.Conn{}

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

	userWorkspace := conf.Wide.UserWorkspaces + string(os.PathSeparator) + username

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

func getCursorOffset(code string, line, ch int) (offset int) {
	lines := strings.Split(code, "\n")

	for i := 0; i < line; i++ {
		offset += len(lines[i])
	}

	offset += line + ch

	return
}
