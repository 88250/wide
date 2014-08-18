package output

import (
	"encoding/json"
	"github.com/88250/wide/session"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
)

var outputWS = map[string]*websocket.Conn{}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "wide-session")
	sid := session.Values["id"].(string)

	outputWS[sid], _ = websocket.Upgrade(w, r, nil, 1024, 1024)

	ret := map[string]interface{}{"output": "Ouput initialized", "cmd": "init-output"}
	outputWS[sid].WriteJSON(&ret)

	glog.Infof("Open a new [Output] with session [%s], %d", sid, len(outputWS))
}

func RunHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "wide-session")
	sid := session.Values["id"].(string)

	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	filePath := args["file"].(string)

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

	argv := []string{"run", filePath}

	cmd := exec.Command("go", argv...)

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

	rec := map[string]interface{}{}

	go func(runningId int) {
		glog.Infof("Session [%s] is running [id=%d, file=%s]", sid, runningId, filePath)

		for {
			buf := make([]byte, 1024)
			count, err := reader.Read(buf)

			if nil != err || 0 == count {
				glog.Infof("Session [%s] 's running [id=%d, file=%s] has done", sid, runningId, filePath)

				break
			} else {
				rec["output"] = string(buf[:count])

				if nil != outputWS[sid] {
					err := outputWS[sid].WriteJSON(&rec)
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
