package output

import (
	"encoding/json"
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/user"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var outputWS = map[string]*websocket.Conn{}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := user.Session.Get(r, "wide-session")
	sid := session.Values["id"].(string)

	outputWS[sid], _ = websocket.Upgrade(w, r, nil, 1024, 1024)

	ret := map[string]interface{}{"output": "Ouput initialized", "cmd": "init-output"}
	outputWS[sid].WriteJSON(&ret)

	glog.Infof("Open a new [Output] with session [%s], %d", sid, len(outputWS))
}

func RunHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := user.Session.Get(r, "wide-session")
	sid := session.Values["id"].(string)

	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	filePath := args["executable"].(string)
	curDir := filePath[:strings.LastIndex(filePath, string(os.PathSeparator))]

	cmd := exec.Command(filePath)
	cmd.Dir = curDir

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	reader := io.MultiReader(stdout, stderr)

	cmd.Start()

	channelRet := map[string]interface{}{}

	go func(runningId int) {
		glog.Infof("Session [%s] is running [id=%d, file=%s]", sid, runningId, filePath)

		for {
			buf := make([]byte, 1024)
			count, err := reader.Read(buf)

			if nil != err || 0 == count {
				glog.Infof("Session [%s] 's running [id=%d, file=%s] has done", sid, runningId, filePath)

				break
			} else {
				channelRet["output"] = string(buf[:count])
				channelRet["cmd"] = "run"

				if nil != outputWS[sid] {
					err := outputWS[sid].WriteJSON(&channelRet)
					if nil != err {
						glog.Error(err)
						break
					}
				}
			}
		}
	}(rand.Int())

	ret, _ := json.Marshal(map[string]interface{}{"succ": true})

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func BuildHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := user.Session.Get(r, "wide-session")
	sid := session.Values["id"].(string)

	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	filePath := args["file"].(string)
	curDir := filePath[:strings.LastIndex(filePath, string(os.PathSeparator))]

	fout, err := os.Create(filePath)

	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	code := args["code"].(string)

	fout.WriteString(code)

	if err := fout.Close(); nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	suffix := ""
	if "windows" == runtime.GOOS {
		suffix = ".exe"
	}
	executable := "main" + suffix
	argv := []string{"build", "-o", executable, filePath}

	cmd := exec.Command("go", argv...)
	cmd.Dir = curDir

	// 设置环境变量（设置当前用户的 GOPATH 等）
	setCmdEnv(cmd)

	glog.Infof("go build -o %s %s", executable, filePath)

	executable = curDir + string(os.PathSeparator) + executable

	// 先把可执行文件删了
	os.RemoveAll(executable)

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	reader := io.MultiReader(stdout, stderr)

	cmd.Start()

	go func(runningId int) {
		glog.Infof("Session [%s] is building [id=%d, file=%s]", sid, runningId, filePath)

		// 一次性读取
		buf := make([]byte, 1024*8)
		count, _ := reader.Read(buf)

		channelRet := map[string]interface{}{}

		channelRet["output"] = string(buf[:count])
		channelRet["cmd"] = "build"
		channelRet["nextCmd"] = "run"
		channelRet["executable"] = executable

		if nil != outputWS[sid] {
			glog.Infof("Session [%s] 's build [id=%d, file=%s] has done", sid, runningId, filePath)

			err := outputWS[sid].WriteJSON(&channelRet)
			if nil != err {
				glog.Error(err)
			}
		}

	}(rand.Int())

	ret, _ := json.Marshal(map[string]interface{}{"succ": true})

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func setCmdEnv(cmd *exec.Cmd) {
	// TODO: 使用用户自己的仓库路径设置 GOPATH
	cmd.Env = append(cmd.Env, "GOPATH="+conf.Wide.Repos, "GOROOT="+os.Getenv("GOROOT"))
}
