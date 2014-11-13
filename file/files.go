// Copyright (c) 2014, B3log
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// File tree manipulations.
package file

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/event"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

// File node, used to construct the file tree.
type FileNode struct {
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	IconSkin  string      `json:"iconSkin"`  // Value should be end with a space
	Type      string      `json:"type"`      // "f": file, "d": directory
	Creatable bool        `json:"creatable"` // whether can create file in this file node
	Removable bool        `json:"removable"` // whether can remove this file node
	Mode      string      `json:"mode"`
	FileNodes []*FileNode `json:"children"`
}

// Source code snippet, used to as the result of "Find Usages", "Search".
type Snippet struct {
	Path     string   `json:"path"`     // file path
	Line     int      `json:"line"`     // line number
	Ch       int      `json:"ch"`       // column number
	Contents []string `json:"contents"` // lines nearby
}

// GetFiles handles request of constructing user workspace file tree.
//
// The Go API source code package ($GOROOT/src/pkg) also as a child node,
// so that users can easily view the Go API source code.
func GetFiles(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	session, _ := session.HTTPSession.Get(r, "wide-session")

	username := session.Values["username"].(string)
	userWorkspace := conf.Wide.GetUserWorkspace(username)
	workspaces := filepath.SplitList(userWorkspace)

	root := FileNode{Name: "root", Path: "", IconSkin: "ico-ztree-dir ", Type: "d", FileNodes: []*FileNode{}}

	// workspace node process
	for _, workspace := range workspaces {
		workspacePath := workspace + conf.PathSeparator + "src"

		workspaceNode := FileNode{Name: workspace[strings.LastIndex(workspace, conf.PathSeparator)+1:],
			Path: workspacePath, IconSkin: "ico-ztree-dir-workspace ", Type: "d",
			Creatable: true, Removable: false, FileNodes: []*FileNode{}}

		walk(workspacePath, &workspaceNode, true, true)

		// add workspace node
		root.FileNodes = append(root.FileNodes, &workspaceNode)
	}

	// construct Go API node
	apiPath := runtime.GOROOT() + conf.PathSeparator + "src" + conf.PathSeparator + "pkg"
	apiNode := FileNode{Name: "Go API", Path: apiPath, IconSkin: "ico-ztree-dir-api ", Type: "d",
		Creatable: false, Removable: false, FileNodes: []*FileNode{}}

	goapiBuildOKSignal := make(chan bool)
	go func() {
		walk(apiPath, &apiNode, false, false)

		// go-ahead
		close(goapiBuildOKSignal)
	}()

	// waiting
	<-goapiBuildOKSignal

	// add Go API node
	root.FileNodes = append(root.FileNodes, &apiNode)

	data["root"] = root
}

// GetFile handles request of opening file by editor.
func GetFile(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	path := args["path"].(string)
	buf, _ := ioutil.ReadFile(path)

	extension := filepath.Ext(path)

	if isImg(extension) {
		// image file will be open in a browser tab

		data["mode"] = "img"

		user := GetUsre(path)
		if nil == user {
			glog.Warningf("The path [%s] has no owner")
			data["path"] = ""

			return
		}

		data["path"] = "/workspace/" + user.Name + "/" + strings.Replace(path, user.GetWorkspace(), "", 1)

		return
	}

	content := string(buf)

	if isBinary(content) {
		data["succ"] = false
		data["msg"] = "Can't open a binary file :("
	} else {
		data["content"] = content
		data["mode"] = getEditorMode(extension)
		data["path"] = path
	}
}

// SaveFile handles request of saving file.
func SaveFile(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	filePath := args["file"].(string)
	sid := args["sid"].(string)

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

		wSession := session.WideSessions.Get(sid)
		wSession.EventQueue.Queue <- &event.Event{Code: event.EvtCodeServerInternalError, Sid: sid,
			Data: "can't save file " + filePath}

		return
	}
}

// NewFile handles request of creating file or directory.
func NewFile(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	path := args["path"].(string)
	fileType := args["fileType"].(string)
	sid := args["sid"].(string)

	wSession := session.WideSessions.Get(sid)

	if !createFile(path, fileType) {
		data["succ"] = false

		wSession.EventQueue.Queue <- &event.Event{Code: event.EvtCodeServerInternalError, Sid: sid,
			Data: "can't create file " + path}

		return
	}

	if "f" == fileType {
		extension := filepath.Ext(path)
		data["mode"] = getEditorMode(extension)
	}
}

// RemoveFile handles request of removing file or directory.
func RemoveFile(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	path := args["path"].(string)
	sid := args["sid"].(string)

	wSession := session.WideSessions.Get(sid)

	if !removeFile(path) {
		data["succ"] = false

		wSession.EventQueue.Queue <- &event.Event{Code: event.EvtCodeServerInternalError, Sid: sid,
			Data: "can't remove file " + path}
	}
}

// RenameFile handles request of renaming file or directory.
func RenameFile(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	oldPath := args["oldPath"].(string)
	newPath := args["newPath"].(string)
	sid := args["sid"].(string)

	wSession := session.WideSessions.Get(sid)

	if !renameFile(oldPath, newPath) {
		data["succ"] = false

		wSession.EventQueue.Queue <- &event.Event{Code: event.EvtCodeServerInternalError, Sid: sid,
			Data: "can't rename file " + oldPath}
	}
}

// Use to find results sorting.
type foundPath struct {
	Path  string `json:"path"`
	score int
}

type foundPaths []*foundPath

func (f foundPaths) Len() int           { return len(f) }
func (f foundPaths) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f foundPaths) Less(i, j int) bool { return f[i].score > f[j].score }

// Find handles request of find files under the specified directory with the specified filename pattern.
func Find(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	path := args["path"].(string) // path of selected file in file tree
	name := args["name"].(string)

	session, _ := session.HTTPSession.Get(r, "wide-session")
	username := session.Values["username"].(string)
	userWorkspace := conf.Wide.GetUserWorkspace(username)
	workspaces := filepath.SplitList(userWorkspace)

	if "" != path && !isDir(path) {
		path = filepath.Dir(path)
	}

	founds := foundPaths{}

	for _, workspace := range workspaces {
		rs := find(workspace, name, []*string{})

		for _, r := range rs {
			substr := util.Str.LCS(path, *r)

			founds = append(founds, &foundPath{Path: *r, score: len(substr)})
		}
	}

	sort.Sort(founds)

	data["founds"] = founds
}

// SearchText handles request of searching files under the specified directory with the specified keyword.
func SearchText(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	dir := args["dir"].(string)
	extension := args["extension"].(string)
	text := args["text"].(string)

	founds := search(dir, extension, text, []*Snippet{})

	data["founds"] = founds
}

// walk traverses the specified path to build a file tree.
func walk(path string, node *FileNode, creatable, removable bool) {
	files := listFiles(path)

	for _, filename := range files {
		fpath := filepath.Join(path, filename)

		fio, _ := os.Lstat(fpath)

		child := FileNode{Name: filename, Path: fpath, Removable: removable, FileNodes: []*FileNode{}}
		node.FileNodes = append(node.FileNodes, &child)

		if nil == fio {
			glog.Warningf("Path [%s] is nil", fpath)

			continue
		}

		if fio.IsDir() {
			child.Type = "d"
			child.Creatable = creatable
			child.IconSkin = "ico-ztree-dir "

			walk(fpath, &child, creatable, removable)
		} else {
			child.Type = "f"
			child.Creatable = creatable
			ext := filepath.Ext(fpath)

			child.IconSkin = getIconSkin(ext)
			child.Mode = getEditorMode(ext)
		}
	}

	return
}

// listFiles lists names of files under the specified dirname.
func listFiles(dirname string) []string {
	f, _ := os.Open(dirname)

	names, _ := f.Readdirnames(-1)
	f.Close()

	sort.Strings(names)

	dirs := []string{}
	files := []string{}

	// sort: directories in front of files
	for _, name := range names {
		path := filepath.Join(dirname, name)
		fio, err := os.Lstat(path)

		if nil != err {
			glog.Warningf("Can't read file info [%s]", path)

			continue
		}

		if fio.IsDir() {
			// exclude the .git direcitory
			if ".git" == fio.Name() {
				continue
			}

			dirs = append(dirs, name)
		} else {
			files = append(files, name)
		}
	}

	return append(dirs, files...)
}

// getIconSkin gets CSS class name of icon with the specified filename extension.
//
// Refers to the zTree document for CSS class names.
func getIconSkin(filenameExtension string) string {
	if isImg(filenameExtension) {
		return "ico-ztree-img "
	}

	switch filenameExtension {
	case ".html", ".htm":
		return "ico-ztree-html "
	case ".go":
		return "ico-ztree-go "
	case ".css":
		return "ico-ztree-css "
	case ".txt":
		return "ico-ztree-text "
	case ".sql":
		return "ico-ztree-sql "
	case ".properties":
		return "ico-ztree-pro "
	case ".md":
		return "ico-ztree-md "
	case ".js", ".json":
		return "ico-ztree-js "
	case ".xml":
		return "ico-ztree-xml "
	default:
		return "ico-ztree-other "
	}
}

// getEditorMode gets editor mode with the specified filename extension.
//
// Refers to the CodeMirror document for modes.
func getEditorMode(filenameExtension string) string {
	switch filenameExtension {
	case ".go":
		return "text/x-go"
	case ".html":
		return "text/html"
	case ".md":
		return "text/x-markdown"
	case ".js":
		return "text/javascript"
	case ".json":
		return "application/json"
	case ".css":
		return "text/css"
	case ".xml":
		return "application/xml"
	case ".sh":
		return "text/x-sh"
	case ".sql":
		return "text/x-sql"
	default:
		return "text/plain"
	}
}

// createFile creates file on the specified path.
//
// fileType:
//
//  "f": file
//  "d": directory
func createFile(path, fileType string) bool {
	switch fileType {
	case "f":
		file, err := os.OpenFile(path, os.O_CREATE, 0775)
		if nil != err {
			glog.Error(err)

			return false
		}

		defer file.Close()

		glog.V(5).Infof("Created file [%s]", path)

		return true
	case "d":
		err := os.Mkdir(path, 0775)

		if nil != err {
			glog.Error(err)

			return false
		}

		glog.V(5).Infof("Created directory [%s]", path)

		return true
	default:
		glog.Errorf("Unsupported file type [%s]", fileType)

		return false
	}
}

// removeFile removes file on the specified path.
func removeFile(path string) bool {
	if err := os.RemoveAll(path); nil != err {
		glog.Errorf("Removes [%s] failed: [%s]", path, err.Error())

		return false
	}

	glog.Infof("Removed [%s]", path)

	return true
}

// renameFile renames (moves) a file from the specified old path to the specified new path.
func renameFile(oldPath, newPath string) bool {
	if err := os.Rename(oldPath, newPath); nil != err {
		glog.Errorf("Renames [%s] failed: [%s]", oldPath, err.Error())

		return false
	}

	glog.Infof("Renamed [%s] to [%s]", oldPath, newPath)

	return true
}

// find finds files under the specified dir and its sub-directoryies with the specified name,
// likes the command 'find dir -name name'.
func find(dir, name string, results []*string) []*string {
	if !strings.HasSuffix(dir, conf.PathSeparator) {
		dir += conf.PathSeparator
	}

	f, _ := os.Open(dir)
	fileInfos, err := f.Readdir(-1)
	f.Close()

	if nil != err {
		glog.Errorf("Read dir [%s] failed: [%s]", dir, err.Error())

		return results
	}

	for _, fileInfo := range fileInfos {
		path := dir + fileInfo.Name()

		if fileInfo.IsDir() {
			// enter the directory recursively
			results = find(path, name, results)
		} else {
			// match filename
			pattern := filepath.Dir(path) + conf.PathSeparator + name
			match, err := filepath.Match(pattern, path)

			if nil != err {
				glog.Errorf("Find match filename failed: [%s]", err.Error)

				continue
			}

			if match {
				results = append(results, &path)
			}
		}
	}

	return results
}

// search finds file under the specified dir and its sub-directories with the specified text, likes the command 'grep'
// or 'findstr'.
func search(dir, extension, text string, snippets []*Snippet) []*Snippet {
	if !strings.HasSuffix(dir, conf.PathSeparator) {
		dir += conf.PathSeparator
	}

	f, _ := os.Open(dir)
	fileInfos, err := f.Readdir(-1)
	f.Close()

	if nil != err {
		glog.Errorf("Read dir [%s] failed: [%s]", dir, err.Error())

		return snippets
	}

	for _, fileInfo := range fileInfos {
		path := dir + fileInfo.Name()

		if fileInfo.IsDir() {
			// enter the directory recursively
			snippets = search(path, extension, text, snippets)
		} else if strings.HasSuffix(path, extension) {
			// grep in file
			ss := searchInFile(path, text)

			snippets = append(snippets, ss...)
		}
	}

	return snippets
}

// searchInFile finds file with the specified path and text.
func searchInFile(path string, text string) []*Snippet {
	ret := []*Snippet{}

	bytes, err := ioutil.ReadFile(path)
	if nil != err {
		glog.Errorf("Read file [%s] failed: [%s]", path, err.Error())

		return ret
	}

	content := string(bytes)
	if isBinary(content) {
		return ret
	}

	lines := strings.Split(content, "\n")

	for idx, line := range lines {
		ch := strings.Index(line, text)

		if -1 != ch {
			snippet := &Snippet{Path: path, Line: idx + 1, Ch: ch + 1, Contents: []string{line}}

			ret = append(ret, snippet)
		}
	}

	return ret
}

// isBinary determines whether the specified content is a binary file content.
func isBinary(content string) bool {
	for _, b := range content {
		if 0 == b {
			return true
		}
	}

	return false
}

// isImg determines whether the specified extension is a image.
func isImg(extension string) bool {
	ext := strings.ToLower(extension)

	switch ext {
	case ".jpg", ".jpeg", ".bmp", ".gif", ".png", ".svg", ".ico":
		return true
	default:
		return false
	}
}

// isDir determines whether the specified path is a directory.
func isDir(path string) bool {
	fio, err := os.Lstat(path)
	if nil != err {
		glog.Warningf("Determines whether [%s] is a directory failed: [%v]", path, err)

		return false
	}

	return fio.IsDir()
}

// GetUsre gets the user the specified path belongs to. Returns nil if not found.
func GetUsre(path string) *conf.User {
	for _, user := range conf.Wide.Users {
		if strings.HasPrefix(path, user.GetWorkspace()) {
			return user
		}
	}

	return nil
}
