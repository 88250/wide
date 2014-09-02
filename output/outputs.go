package output

import (
	"encoding/json"
	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/user"
	"github.com/b3log/wide/util"
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
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	session, _ := user.Session.Get(r, "wide-session")
	sid := session.Values["id"].(string)

	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	filePath := args["executable"].(string)
	curDir := filePath[:strings.LastIndex(filePath, string(os.PathSeparator))]

	cmd := exec.Command(filePath)
	cmd.Dir = curDir

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false

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
}

func BuildHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	session, _ := user.Session.Get(r, "wide-session")
	sid := session.Values["id"].(string)
	username := session.Values["username"].(string)

	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	filePath := args["file"].(string)
	curDir := filePath[:strings.LastIndex(filePath, string(os.PathSeparator))]

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

	suffix := ""
	if "windows" == runtime.GOOS {
		suffix = ".exe"
	}
	executable := "main" + suffix
	argv := []string{"build", "-o", executable, filePath}

	cmd := exec.Command("go", argv...)
	cmd.Dir = curDir

	// 设置环境变量（设置当前用户的 GOPATH 等）
	setCmdEnv(cmd, username)

	glog.Infof("go build -o %s %s", executable, filePath)

	executable = curDir + string(os.PathSeparator) + executable

	// 先把可执行文件删了
	err = os.RemoveAll(executable)
	if nil != err {
		glog.Info(err)
		data["succ"] = false

		return
	}

	stdout, err := cmd.StdoutPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	stderr, err := cmd.StderrPipe()
	if nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	if data["succ"].(bool) {
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
	}
}

func setCmdEnv(cmd *exec.Cmd, username string) {
	userRepos := strings.Replace(conf.Wide.UserRepos, "{user}", username, -1)
	userWorkspace := userRepos[:strings.LastIndex(userRepos, "/src")]

	// glog.Infof("User [%s] workspace [%s]", username, userWorkspace)

	masterWorkspace := conf.Wide.Repos[:strings.LastIndex(conf.Wide.Repos, "/src")]
	// glog.Infof("Master workspace [%s]", masterWorkspace)

	cmd.Env = append(cmd.Env,
		"GOPATH="+userWorkspace+string(os.PathListSeparator)+
			masterWorkspace+string(os.PathListSeparator)+
			os.Getenv("GOPATH"),
		"GOROOT="+os.Getenv("GOROOT"))
}
