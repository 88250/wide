// 构建、运行、go tool 操作.
package output

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/i18n"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

const (
	lintSeverityError = "error"   // Lint 严重级别：错误
	lintSeverityWarn  = "warning" // Lint 严重级别：警告
)

// 代码 Lint 结构.
type Lint struct {
	File     string `json:"file"`
	LineNo   int    `json:"lineNo"`
	Severity string `json:"severity"`
	Msg      string `json:"msg"`
}

// 建立输出通道.
func WSHandler(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query()["sid"][0]

	conn, _ := websocket.Upgrade(w, r, nil, 1024, 1024)
	wsChan := util.WSChannel{Sid: sid, Conn: conn, Request: r, Time: time.Now()}

	session.OutputWS[sid] = &wsChan

	ret := map[string]interface{}{"output": "Ouput initialized", "cmd": "init-output"}
	wsChan.Conn.WriteJSON(&ret)

	glog.V(4).Infof("Open a new [Output] with session [%s], %d", sid, len(session.OutputWS))
}

// 运行一个可执行文件.
func RunHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	sid := args["sid"].(string)
	wSession := session.WideSessions.Get(sid)
	if nil == wSession {
		data["succ"] = false

		return
	}

	filePath := args["executable"].(string)
	curDir := filepath.Dir(filePath)

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

	reader := bufio.NewReader(io.MultiReader(stdout, stderr))

	if err := cmd.Start(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	// 添加到用户进程集中
	processes.add(wSession, cmd.Process)

	channelRet := map[string]interface{}{}
	channelRet["pid"] = cmd.Process.Pid

	go func(runningId int) {
		defer util.Recover()
		defer cmd.Wait()

		glog.V(3).Infof("Session [%s] is running [id=%d, file=%s]", sid, runningId, filePath)

		// 在读取程序输出前先返回一次，使前端获取到 run 状态与 pid
		if nil != session.OutputWS[sid] {
			wsChannel := session.OutputWS[sid]

			channelRet["cmd"] = "run"
			channelRet["output"] = ""
			err := wsChannel.Conn.WriteJSON(&channelRet)
			if nil != err {
				glog.Error(err)
				return
			}

			// 更新通道最近使用时间
			wsChannel.Time = time.Now()
		}

		for {
			buf, err := reader.ReadBytes('\n')

			if nil != err || 0 == len(buf) {
				// 从用户进程集中移除这个执行完毕（或是被主动停止）的进程
				processes.remove(wSession, cmd.Process)

				glog.V(3).Infof("Session [%s] 's running [id=%d, file=%s] has done", sid, runningId, filePath)

				if nil != session.OutputWS[sid] {
					wsChannel := session.OutputWS[sid]

					channelRet["cmd"] = "run-done"
					channelRet["output"] = "<pre>" + string(buf) + "</pre>"
					err := wsChannel.Conn.WriteJSON(&channelRet)
					if nil != err {
						glog.Error(err)
						break
					}

					// 更新通道最近使用时间
					wsChannel.Time = time.Now()
				}

				break
			} else {
				if nil != session.OutputWS[sid] {
					wsChannel := session.OutputWS[sid]

					channelRet["cmd"] = "run"
					channelRet["output"] = "<pre>" + string(buf) + "</pre>"
					err := wsChannel.Conn.WriteJSON(&channelRet)
					if nil != err {
						glog.Error(err)
						break
					}

					// 更新通道最近使用时间
					wsChannel.Time = time.Now()
				}
			}
		}
	}(rand.Int())
}

// 构建可执行文件.
func BuildHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	username := httpSession.Values["username"].(string)
	locale := conf.Wide.GetUser(username).Locale

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	sid := args["sid"].(string)

	filePath := args["file"].(string)
	curDir := filepath.Dir(filePath)

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
	if util.OS.IsWindows() {
		suffix = ".exe"
	}
	executable := "main" + suffix
	argv := []string{"build", "-o", executable}

	cmd := exec.Command("go", argv...)
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

	glog.V(5).Infof("go build -o %s", executable)

	executable = filepath.Join(curDir, executable)

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

	if !data["succ"].(bool) {
		return
	}

	channelRet := map[string]interface{}{}

	if nil != session.OutputWS[sid] {
		// 在前端 output 中显示“开始构建”

		channelRet["output"] = "<span class='start-build'>" + i18n.Get(locale, "start-build").(string) + "</span>\n"
		channelRet["cmd"] = "start-build"

		wsChannel := session.OutputWS[sid]

		err := wsChannel.Conn.WriteJSON(&channelRet)
		if nil != err {
			glog.Error(err)
			return
		}

		// 更新通道最近使用时间
		wsChannel.Time = time.Now()
	}

	reader := bufio.NewReader(io.MultiReader(stdout, stderr))

	if err := cmd.Start(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	go func(runningId int) {
		defer util.Recover()
		defer cmd.Wait()

		glog.V(3).Infof("Session [%s] is building [id=%d, dir=%s]", sid, runningId, curDir)

		// 一次性读取
		buf, _ := ioutil.ReadAll(reader)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "build"
		channelRet["executable"] = executable

		if 0 == len(buf) { // 说明构建成功，没有错误信息输出
			// 设置下一次执行命令（前端会根据该参数发送请求）
			channelRet["nextCmd"] = args["nextCmd"]
			channelRet["output"] = "<span class='build-succ'>" + i18n.Get(locale, "build-succ").(string) + "</span>\n"

			go func() { // 运行 go install，生成的库用于 gocode lib-path
				cmd := exec.Command("go", "install")
				cmd.Dir = curDir

				setCmdEnv(cmd, username)

				out, _ := cmd.CombinedOutput()
				if len(out) > 0 {
					glog.Warning(string(out))
				}
			}()
		} else { // 构建失败
			// 解析错误信息，返回给编辑器 gutter lint
			errOut := string(buf)
			channelRet["output"] = "<span class='build-error'>" + i18n.Get(locale, "build-error").(string) + "</span>\n" + errOut

			lines := strings.Split(errOut, "\n")

			if lines[0][0] == '#' {
				lines = lines[1:] // 跳过第一行
			}

			lints := []*Lint{}

			for _, line := range lines {
				if len(line) < 1 {
					continue
				}

				if line[0] == '\t' {
					// 添加到上一个 lint 中
					last := len(lints)
					msg := lints[last-1].Msg
					msg += line

					lints[last-1].Msg = msg

					continue
				}

				file := line[:strings.Index(line, ":")]
				left := line[strings.Index(line, ":")+1:]
				index := strings.Index(left, ":")
				lineNo := 0
				msg := left
				if index >= 0 {
					lineNo, _ = strconv.Atoi(left[:index])
					msg = left[index+2:]
				}

				lint := &Lint{
					File:     file,
					LineNo:   lineNo - 1,
					Severity: lintSeverityError,
					Msg:      msg,
				}

				lints = append(lints, lint)
			}

			channelRet["lints"] = lints
		}

		if nil != session.OutputWS[sid] {
			glog.V(3).Infof("Session [%s] 's build [id=%d, dir=%s] has done", sid, runningId, curDir)

			wsChannel := session.OutputWS[sid]
			err := wsChannel.Conn.WriteJSON(&channelRet)
			if nil != err {
				glog.Error(err)
			}

			// 更新通道最近使用时间
			wsChannel.Time = time.Now()
		}

	}(rand.Int())
}

// go test.
func GoTestHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	username := httpSession.Values["username"].(string)
	locale := conf.Wide.GetUser(username).Locale

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	sid := args["sid"].(string)

	filePath := args["file"].(string)
	curDir := filepath.Dir(filePath)

	cmd := exec.Command("go", "test", "-v")
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

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

	if !data["succ"].(bool) {
		return
	}

	channelRet := map[string]interface{}{}

	if nil != session.OutputWS[sid] {
		// 在前端 output 中显示“开始 go test

		channelRet["output"] = "<span class='start-test'>" + i18n.Get(locale, "start-test").(string) + "</span>\n"
		channelRet["cmd"] = "start-test"

		wsChannel := session.OutputWS[sid]

		err := wsChannel.Conn.WriteJSON(&channelRet)
		if nil != err {
			glog.Error(err)
			return
		}

		// 更新通道最近使用时间
		wsChannel.Time = time.Now()
	}

	reader := bufio.NewReader(io.MultiReader(stdout, stderr))

	if err := cmd.Start(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	go func(runningId int) {
		defer util.Recover()

		glog.V(3).Infof("Session [%s] is running [go test] [runningId=%d]", sid, runningId)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "go test"

		// 一次性读取
		buf, _ := ioutil.ReadAll(reader)

		// 同步点，等待 go test 执行完成
		cmd.Wait()

		if !cmd.ProcessState.Success() {
			glog.V(3).Infof("Session [%s] 's running [go test] [runningId=%d] has done (with error)", sid, runningId)

			channelRet["output"] = "<span class='test-error'>" + i18n.Get(locale, "test-error").(string) + "</span>\n" + string(buf)
		} else {
			glog.V(3).Infof("Session [%s] 's running [go test] [runningId=%d] has done", sid, runningId)

			channelRet["output"] = "<span class='test-succ'>" + i18n.Get(locale, "test-succ").(string) + "</span>\n" + string(buf)
		}

		if nil != session.OutputWS[sid] {
			wsChannel := session.OutputWS[sid]

			err := wsChannel.Conn.WriteJSON(&channelRet)
			if nil != err {
				glog.Error(err)
			}

			// 更新通道最近使用时间
			wsChannel.Time = time.Now()
		}
	}(rand.Int())
}

// go install.
func GoInstallHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	username := httpSession.Values["username"].(string)
	locale := conf.Wide.GetUser(username).Locale

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	sid := args["sid"].(string)

	filePath := args["file"].(string)
	curDir := filepath.Dir(filePath)

	cmd := exec.Command("go", "install")
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

	glog.V(5).Infof("go install %s", curDir)

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

	if !data["succ"].(bool) {
		return
	}

	channelRet := map[string]interface{}{}

	if nil != session.OutputWS[sid] {
		// 在前端 output 中显示“开始 go install”

		channelRet["output"] = "<span class='start-install'>" + i18n.Get(locale, "start-install").(string) + "</span>\n"
		channelRet["cmd"] = "start-install"

		wsChannel := session.OutputWS[sid]

		err := wsChannel.Conn.WriteJSON(&channelRet)
		if nil != err {
			glog.Error(err)
			return
		}

		// 更新通道最近使用时间
		wsChannel.Time = time.Now()
	}

	reader := bufio.NewReader(io.MultiReader(stdout, stderr))

	if err := cmd.Start(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	go func(runningId int) {
		defer util.Recover()
		defer cmd.Wait()

		glog.V(3).Infof("Session [%s] is running [go install] [id=%d, dir=%s]", sid, runningId, curDir)

		// 一次性读取
		buf, _ := ioutil.ReadAll(reader)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "go install"

		if 0 != len(buf) { // 构建失败
			// 解析错误信息，返回给编辑器 gutter lint
			errOut := string(buf)
			lines := strings.Split(errOut, "\n")

			if lines[0][0] == '#' {
				lines = lines[1:] // 跳过第一行
			}

			lints := []*Lint{}

			for _, line := range lines {
				if len(line) < 1 {
					continue
				}

				if line[0] == '\t' {
					// 添加到上一个 lint 中
					last := len(lints)
					msg := lints[last-1].Msg
					msg += line

					lints[last-1].Msg = msg

					continue
				}

				file := line[:strings.Index(line, ":")]
				left := line[strings.Index(line, ":")+1:]
				index := strings.Index(left, ":")
				lineNo := 0
				msg := left
				if index >= 0 {
					lineNo, _ = strconv.Atoi(left[:index])
					msg = left[index+2:]
				}

				lint := &Lint{
					File:     file,
					LineNo:   lineNo - 1,
					Severity: lintSeverityError,
					Msg:      msg,
				}

				lints = append(lints, lint)
			}

			channelRet["lints"] = lints

			channelRet["output"] = "<span class='install-error'>" + i18n.Get(locale, "install-error").(string) + "</span>\n" + errOut
		} else {
			channelRet["output"] = "<span class='install-succ'>" + i18n.Get(locale, "install-succ").(string) + "</span>\n"
		}

		if nil != session.OutputWS[sid] {
			glog.V(3).Infof("Session [%s] 's running [go install] [id=%d, dir=%s] has done", sid, runningId, curDir)

			wsChannel := session.OutputWS[sid]
			err := wsChannel.Conn.WriteJSON(&channelRet)
			if nil != err {
				glog.Error(err)
			}

			// 更新通道最近使用时间
			wsChannel.Time = time.Now()
		}

	}(rand.Int())
}

// go get.
func GoGetHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	username := httpSession.Values["username"].(string)
	locale := conf.Wide.GetUser(username).Locale

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	sid := args["sid"].(string)

	filePath := args["file"].(string)
	curDir := filepath.Dir(filePath)

	cmd := exec.Command("go", "get")
	cmd.Dir = curDir

	setCmdEnv(cmd, username)

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

	if !data["succ"].(bool) {
		return
	}

	channelRet := map[string]interface{}{}

	if nil != session.OutputWS[sid] {
		// 在前端 output 中显示“开始 go get

		channelRet["output"] = "<span class='start-get'>" + i18n.Get(locale, "start-get").(string) + "</span>\n"
		channelRet["cmd"] = "start-get"

		wsChannel := session.OutputWS[sid]

		err := wsChannel.Conn.WriteJSON(&channelRet)
		if nil != err {
			glog.Error(err)
			return
		}

		// 更新通道最近使用时间
		wsChannel.Time = time.Now()
	}

	reader := bufio.NewReader(io.MultiReader(stdout, stderr))

	if err := cmd.Start(); nil != err {
		glog.Error(err)
		data["succ"] = false

		return
	}

	go func(runningId int) {
		defer util.Recover()
		defer cmd.Wait()

		glog.V(3).Infof("Session [%s] is running [go get] [runningId=%d]", sid, runningId)

		channelRet := map[string]interface{}{}
		channelRet["cmd"] = "go get"

		// 一次性读取
		buf, _ := ioutil.ReadAll(reader)

		if 0 != len(buf) {
			glog.V(3).Infof("Session [%s] 's running [go get] [runningId=%d] has done (with error)", sid, runningId)

			channelRet["output"] = "<span class='get-error'>" + i18n.Get(locale, "get-error").(string) + "</span>\n" + string(buf)
		} else {
			glog.V(3).Infof("Session [%s] 's running [go get] [runningId=%d] has done", sid, runningId)

			channelRet["output"] = "<span class='get-succ'>" + i18n.Get(locale, "get-succ").(string) + "</span>\n"

		}

		if nil != session.OutputWS[sid] {
			wsChannel := session.OutputWS[sid]

			err := wsChannel.Conn.WriteJSON(&channelRet)
			if nil != err {
				glog.Error(err)
			}

			// 更新通道最近使用时间
			wsChannel.Time = time.Now()
		}
	}(rand.Int())
}

// 结束正在运行的进程.
func StopHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	sid := args["sid"].(string)
	pid := int(args["pid"].(float64))

	wSession := session.WideSessions.Get(sid)
	if nil == wSession {
		data["succ"] = false

		return
	}

	processes.kill(wSession, pid)
}

func setCmdEnv(cmd *exec.Cmd, username string) {
	userWorkspace := conf.Wide.GetUserWorkspace(username)
	masterWorkspace := conf.Wide.GetWorkspace()

	cmd.Env = append(cmd.Env,
		"GOPATH="+userWorkspace+conf.PathListSeparator+masterWorkspace,
		"GOOS="+runtime.GOOS,
		"GOARCH="+runtime.GOARCH,
		"GOROOT="+runtime.GOROOT(),
		"PATH="+os.Getenv("PATH"))
}
