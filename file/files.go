package file

import (
	"encoding/json"
	"github.com/88250/wide/conf"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func GetFiles(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}

	root := FileNode{"projects", conf.Wide.GOPATH, "d", []*FileNode{}}
	fileInfo, _ := os.Lstat(conf.Wide.GOPATH)

	walk(conf.Wide.GOPATH, fileInfo, &root)

	data["root"] = root

	ret, _ := json.Marshal(data)

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	path := args["path"].(string)

	idx := strings.LastIndex(path, ".")

	buf, _ := ioutil.ReadFile(path)

	content := string(buf)

	data := map[string]interface{}{"succ": true}
	data["content"] = content

	extension := ""
	if 0 <= idx {
		extension = path[idx:]
	}
	data["mode"] = getEditorMode(extension)

	ret, _ := json.Marshal(data)

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func SaveFile(w http.ResponseWriter, r *http.Request) {
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

	ret, _ := json.Marshal(map[string]interface{}{"succ": true})

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func NewFile(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	data := map[string]interface{}{"succ": true}

	path := args["path"].(string)
	fileType := args["fileType"].(string)

	if !createFile(path, fileType) {
		data["succ"] = false
	}

	ret, _ := json.Marshal(data)

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

func RemoveFile(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	data := map[string]interface{}{"succ": true}

	path := args["path"].(string)

	if !removeFile(path) {
		data["succ"] = false
	}

	ret, _ := json.Marshal(data)

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

type FileNode struct {
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	Type      string      `json:"type"`
	FileNodes []*FileNode `json:"children"`
}

func walk(path string, info os.FileInfo, node *FileNode) {
	files := listFiles(path)

	for _, filename := range files {
		fpath := filepath.Join(path, filename)

		fio, _ := os.Lstat(fpath)

		child := FileNode{Name: filename, Path: fpath, FileNodes: []*FileNode{}}
		node.FileNodes = append(node.FileNodes, &child)

		if nil == fio {
			glog.Warningf("Path [%s] is nil", fpath)

			continue
		}

		if fio.IsDir() {
			child.Type = "d"
			walk(fpath, fio, &child)
		} else {
			child.Type = "f"
		}
	}

	return
}

func listFiles(dirname string) []string {
	f, _ := os.Open(dirname)

	names, _ := f.Readdirnames(-1)
	f.Close()

	sort.Strings(names)

	return names
}

func getEditorMode(filenameExtension string) string {
	switch filenameExtension {
	case ".go":
		return "go"
	case ".html":
		return "htmlmixed"
	case ".md":
		return "markdown"
	case ".js", ".json":
		return "javascript"
	case ".css":
		return "css"
	case ".xml":
		return "xml"
	case ".sh":
		return "shell"
	case ".sql":
		return "sql"
	default:
		return "text"
	}
}

func createFile(path, fileType string) bool {
	switch fileType {
	case "f":
		file, err := os.OpenFile(path, os.O_CREATE, 0664)
		if nil != err {
			glog.Info(err)

			return false
		}

		defer file.Close()

		glog.Infof("Created file [%s]", path)

		return true
	case "d":
		err := os.Mkdir(path, 0775)

		if nil != err {
			glog.Info(err)
		}

		glog.Infof("Created directory [%s]", path)

		return true
	default:

		glog.Infof("Unsupported file type [%s]", fileType)

		return false
	}
}

func removeFile(path string) bool {
	if err := os.RemoveAll(path); nil != err {
		glog.Errorf("Removes [%s] failed: [%s]", path, err.Error())

		return false
	}

	glog.Infof("Removed [%s]", path)

	return true
}
