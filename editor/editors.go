package editor

import (
	"bytes"
	"encoding/json"
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/user"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

func FmtHandler(w http.ResponseWriter, r *http.Request) {
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

	data["code"] = string(output)
}

func AutocompleteHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	//session, _ := user.Session.Get(r, "wide-session")
	//username := session.Values["username"].(string)

	code := args["code"].(string)
	line := int(args["cursorLine"].(float64))
	ch := int(args["cursorCh"].(float64))

	offset := getCursorOffset(code, line, ch)

	// glog.Infof("offset: %d", offset)

	argv := []string{"-f=json", "autocomplete", strconv.Itoa(offset)}

	var output bytes.Buffer

	cmd := exec.Command("gocode", argv...)

	// 设置环境变量（设置当前用户的 GOPATH 等）
	// FIXME: setCmdEnv(cmd, username)

	cmd.Stdout = &output

	stdin, _ := cmd.StdinPipe()
	cmd.Start()
	stdin.Write([]byte(code))
	stdin.Close()
	cmd.Wait()

	w.Header().Set("Content-Type", "application/json")
	w.Write(output.Bytes())
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
	userRepos := strings.Replace(conf.Wide.UserRepos, "{user}", username, -1)
	userWorkspace := userRepos[:strings.LastIndex(userRepos, "/src")]

	glog.Infof("User [%s] workspace [%s]", username, userWorkspace)

	masterWorkspace := conf.Wide.Repos[:strings.LastIndex(conf.Wide.Repos, "/src")]
	glog.Infof("Master workspace [%s]", masterWorkspace)

	GOPATH := os.Getenv("GOPATH")
	glog.Infof("Env GOPATH [%s]", GOPATH)

	cmd.Env = append(cmd.Env,
		"GOPATH="+userWorkspace+string(os.PathListSeparator)+
			masterWorkspace+string(os.PathListSeparator)+
			GOPATH,
		"GOROOT="+os.Getenv("GOROOT"))
}
