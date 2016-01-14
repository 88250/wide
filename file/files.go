// Copyright (c) 2014-2016, b3log.org & hacpai.com
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

// Package file includes file related manipulations.
package file

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/b3log/wide/conf"
	"github.com/b3log/wide/event"
	"github.com/b3log/wide/log"
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
)

// Logger.
var logger = log.NewLogger(os.Stdout)

// Node represents a file node in file tree.
type Node struct {
	Id        string  `json:"id"`
	Name      string  `json:"name"`
	Path      string  `json:"path"`
	IconSkin  string  `json:"iconSkin"` // Value should be end with a space
	IsParent  bool    `json:"isParent"`
	Type      string  `json:"type"`      // "f": file, "d": directory
	Creatable bool    `json:"creatable"` // whether can create file in this file node
	Removable bool    `json:"removable"` // whether can remove this file node
	IsGoAPI   bool    `json:"isGOAPI"`
	Mode      string  `json:"mode"`
	Children  []*Node `json:"children"`
}

// Snippet represents a source code snippet, used to as the result of "Find Usages", "Search".
type Snippet struct {
	Path     string   `json:"path"`     // file path
	Line     int      `json:"line"`     // line number
	Ch       int      `json:"ch"`       // column number
	Contents []string `json:"contents"` // lines nearby
}

var apiNode *Node

// initAPINode builds the Go API file node.
func initAPINode() {
	apiPath := util.Go.GetAPIPath()

	apiNode = &Node{Name: "Go API", Path: apiPath, IconSkin: "ico-ztree-dir-api ", Type: "d",
		Creatable: false, Removable: false, IsGoAPI: true, Children: []*Node{}}

	walk(apiPath, apiNode, false, false, true)
}

// GetFilesHandler handles request of constructing user workspace file tree.
//
// The Go API source code package also as a child node,
// so that users can easily view the Go API source code in file tree.
func GetFilesHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)

	result := util.NewResult()
	defer util.RetGzResult(w, r, result)

	userWorkspace := conf.GetUserWorkspace(username)
	workspaces := filepath.SplitList(userWorkspace)

	root := Node{Name: "root", Path: "", IconSkin: "ico-ztree-dir ", Type: "d", IsParent: true, Children: []*Node{}}

	if nil == apiNode { // lazy init
		initAPINode()
	}

	// workspace node process
	for _, workspace := range workspaces {
		workspacePath := workspace + conf.PathSeparator + "src"

		workspaceNode := Node{
			Id:        filepath.ToSlash(workspacePath), // jQuery API can't accept "\", so we convert it to "/"
			Name:      workspace[strings.LastIndex(workspace, conf.PathSeparator)+1:],
			Path:      filepath.ToSlash(workspacePath),
			IconSkin:  "ico-ztree-dir-workspace ",
			Type:      "d",
			Creatable: true,
			Removable: false,
			IsGoAPI:   false,
			Children:  []*Node{}}

		walk(workspacePath, &workspaceNode, true, true, false)

		// add workspace node
		root.Children = append(root.Children, &workspaceNode)
	}

	// add Go API node
	root.Children = append(root.Children, apiNode)

	result.Data = root
}

// RefreshDirectoryHandler handles request of refresh a directory of file tree.
func RefreshDirectoryHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)

	r.ParseForm()
	path := r.FormValue("path")

	if !util.Go.IsAPI(path) && !session.CanAccess(username, path) {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	node := Node{Name: "root", Path: path, IconSkin: "ico-ztree-dir ", Type: "d", Children: []*Node{}}

	walk(path, &node, true, true, false)

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(node.Children)
	if err != nil {
		logger.Error(err)
		return
	}

	w.Write(data)
}

// GetFileHandler handles request of opening file by editor.
func GetFileHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)

	result := util.NewResult()
	defer util.RetResult(w, r, result)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	path := args["path"].(string)

	if !util.Go.IsAPI(path) && !session.CanAccess(username, path) {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	size := util.File.GetFileSize(path)
	if size > 5242880 { // 5M
		result.Succ = false
		result.Msg = "This file is too large to open :("

		return
	}

	data := map[string]interface{}{}
	result.Data = &data

	buf, _ := ioutil.ReadFile(path)

	extension := filepath.Ext(path)

	if util.File.IsImg(extension) {
		// image file will be open in a browser tab

		data["mode"] = "img"

		username := conf.GetOwner(path)
		if "" == username {
			logger.Warnf("The path [%s] has no owner")
			data["path"] = ""

			return
		}

		user := conf.GetUser(username)

		data["path"] = "/workspace/" + user.Name + "/" + strings.Replace(path, user.GetWorkspace(), "", 1)

		return
	}

	content := string(buf)

	if util.File.IsBinary(content) {
		result.Succ = false
		result.Msg = "Can't open a binary file :("
	} else {
		data["content"] = content
		data["path"] = path
	}
}

// SaveFileHandler handles request of saving file.
func SaveFileHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)

	result := util.NewResult()
	defer util.RetResult(w, r, result)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	filePath := args["file"].(string)
	sid := args["sid"].(string)

	if util.Go.IsAPI(filePath) || !session.CanAccess(username, filePath) {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	fout, err := os.Create(filePath)

	if nil != err {
		logger.Error(err)
		result.Succ = false

		return
	}

	code := args["code"].(string)

	fout.WriteString(code)

	if err := fout.Close(); nil != err {
		logger.Error(err)
		result.Succ = false

		wSession := session.WideSessions.Get(sid)
		wSession.EventQueue.Queue <- &event.Event{Code: event.EvtCodeServerInternalError, Sid: sid,
			Data: "can't save file " + filePath}

		return
	}
}

// NewFileHandler handles request of creating file or directory.
func NewFileHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)

	result := util.NewResult()
	defer util.RetResult(w, r, result)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	path := args["path"].(string)

	if util.Go.IsAPI(path) || !session.CanAccess(username, path) {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	fileType := args["fileType"].(string)
	sid := args["sid"].(string)

	wSession := session.WideSessions.Get(sid)

	if !createFile(path, fileType) {
		result.Succ = false

		wSession.EventQueue.Queue <- &event.Event{Code: event.EvtCodeServerInternalError, Sid: sid,
			Data: "can't create file " + path}

		return
	}

	if "f" == fileType {
		logger.Debugf("Created a file [%s] by user [%s]", path, wSession.Username)
	} else {
		logger.Debugf("Created a dir [%s] by user [%s]", path, wSession.Username)
	}

}

// RemoveFileHandler handles request of removing file or directory.
func RemoveFileHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)

	result := util.NewResult()
	defer util.RetResult(w, r, result)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	path := args["path"].(string)

	if util.Go.IsAPI(path) || !session.CanAccess(username, path) {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	sid := args["sid"].(string)

	wSession := session.WideSessions.Get(sid)

	if !removeFile(path) {
		result.Succ = false

		wSession.EventQueue.Queue <- &event.Event{Code: event.EvtCodeServerInternalError, Sid: sid,
			Data: "can't remove file " + path}

		return
	}

	logger.Debugf("Removed a file [%s] by user [%s]", path, wSession.Username)
}

// RenameFileHandler handles request of renaming file or directory.
func RenameFileHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)

	result := util.NewResult()
	defer util.RetResult(w, r, result)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	oldPath := args["oldPath"].(string)
	if util.Go.IsAPI(oldPath) ||
		!session.CanAccess(username, oldPath) {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	newPath := args["newPath"].(string)
	if util.Go.IsAPI(newPath) || !session.CanAccess(username, newPath) {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	sid := args["sid"].(string)

	wSession := session.WideSessions.Get(sid)

	if !renameFile(oldPath, newPath) {
		result.Succ = false

		wSession.EventQueue.Queue <- &event.Event{Code: event.EvtCodeServerInternalError, Sid: sid,
			Data: "can't rename file " + oldPath}

		return
	}

	logger.Debugf("Renamed a file [%s] to [%s] by user [%s]", oldPath, newPath, wSession.Username)
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

// FindHandler handles request of find files under the specified directory with the specified filename pattern.
func FindHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}
	username := httpSession.Values["username"].(string)

	result := util.NewResult()
	defer util.RetResult(w, r, result)

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	path := args["path"].(string) // path of selected file in file tree
	if !util.Go.IsAPI(path) && !session.CanAccess(username, path) {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	name := args["name"].(string)

	userWorkspace := conf.GetUserWorkspace(username)
	workspaces := filepath.SplitList(userWorkspace)

	if "" != path && !util.File.IsDir(path) {
		path = filepath.Dir(path)
	}

	founds := foundPaths{}

	for _, workspace := range workspaces {
		rs := find(workspace+conf.PathSeparator+"src", name, []*string{})

		for _, r := range rs {
			substr := util.Str.LCS(path, *r)

			founds = append(founds, &foundPath{Path: filepath.ToSlash(*r), score: len(substr)})
		}
	}

	sort.Sort(founds)

	result.Data = founds
}

// SearchTextHandler handles request of searching files under the specified directory with the specified keyword.
func SearchTextHandler(w http.ResponseWriter, r *http.Request) {
	httpSession, _ := session.HTTPSession.Get(r, "wide-session")
	if httpSession.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	result := util.NewResult()
	defer util.RetResult(w, r, result)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	sid := args["sid"].(string)
	wSession := session.WideSessions.Get(sid)
	if nil == wSession {
		result.Succ = false

		return
	}

	// XXX: just one directory

	dir := args["dir"].(string)
	if "" == dir {
		userWorkspace := conf.GetUserWorkspace(wSession.Username)
		workspaces := filepath.SplitList(userWorkspace)
		dir = workspaces[0]
	}

	extension := args["extension"].(string)
	text := args["text"].(string)

	founds := []*Snippet{}
	if util.File.IsDir(dir) {
		founds = search(dir, extension, text, []*Snippet{})
	} else {
		founds = searchInFile(dir, text)
	}

	result.Data = founds
}

// walk traverses the specified path to build a file tree.
func walk(path string, node *Node, creatable, removable, isGOAPI bool) {
	files := listFiles(path)

	for _, filename := range files {
		fpath := filepath.Join(path, filename)

		fio, _ := os.Lstat(fpath)

		child := Node{
			Id:        filepath.ToSlash(fpath), // jQuery API can't accept "\", so we convert it to "/"
			Name:      filename,
			Path:      filepath.ToSlash(fpath),
			Removable: removable,
			IsGoAPI:   isGOAPI,
			Children:  []*Node{}}
		node.Children = append(node.Children, &child)

		if nil == fio {
			logger.Warnf("Path [%s] is nil", fpath)

			continue
		}

		if fio.IsDir() {
			child.Type = "d"
			child.Creatable = creatable
			child.IconSkin = "ico-ztree-dir "
			child.IsParent = true

			walk(fpath, &child, creatable, removable, isGOAPI)
		} else {
			child.Type = "f"
			child.Creatable = creatable
			ext := filepath.Ext(fpath)

			child.IconSkin = getIconSkin(ext)
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
			logger.Warnf("Can't read file info [%s]", path)

			continue
		}

		if fio.IsDir() {
			// exclude the .git, .svn, .hg direcitory
			if ".git" == fio.Name() || ".svn" == fio.Name() || ".hg" == fio.Name() {
				continue
			}

			dirs = append(dirs, name)
		} else {
			// exclude the .DS_Store directory on Mac OS X
			if ".DS_Store" == fio.Name() {
				continue
			}

			files = append(files, name)
		}
	}

	return append(dirs, files...)
}

// getIconSkin gets CSS class name of icon with the specified filename extension.
//
// Refers to the zTree document for CSS class names.
func getIconSkin(filenameExtension string) string {
	if util.File.IsImg(filenameExtension) {
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
			logger.Error(err)

			return false
		}

		defer file.Close()

		logger.Tracef("Created file [%s]", path)

		return true
	case "d":
		err := os.Mkdir(path, 0775)

		if nil != err {
			logger.Error(err)

			return false
		}

		logger.Tracef("Created directory [%s]", path)

		return true
	default:
		logger.Errorf("Unsupported file type [%s]", fileType)

		return false
	}
}

// removeFile removes file on the specified path.
func removeFile(path string) bool {
	if err := os.RemoveAll(path); nil != err {
		logger.Errorf("Removes [%s] failed: [%s]", path, err.Error())

		return false
	}

	logger.Tracef("Removed [%s]", path)

	return true
}

// renameFile renames (moves) a file from the specified old path to the specified new path.
func renameFile(oldPath, newPath string) bool {
	if err := os.Rename(oldPath, newPath); nil != err {
		logger.Errorf("Renames [%s] failed: [%s]", oldPath, err.Error())

		return false
	}

	logger.Tracef("Renamed [%s] to [%s]", oldPath, newPath)

	return true
}

// Default exclude file name patterns when find.
var defaultExcludesFind = []string{".git", ".svn", ".repository", "CVS", "RCS", "SCCS", ".bzr", ".metadata", ".hg"}

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
		logger.Errorf("Read dir [%s] failed: [%s]", dir, err.Error())

		return results
	}

	for _, fileInfo := range fileInfos {
		fname := fileInfo.Name()
		path := dir + fname

		if fileInfo.IsDir() {
			if util.Str.Contains(fname, defaultExcludesFind) {
				continue
			}

			// enter the directory recursively
			results = find(path, name, results)
		} else {
			// match filename
			pattern := filepath.Dir(path) + conf.PathSeparator + name

			match, err := filepath.Match(strings.ToLower(pattern), strings.ToLower(path))

			if nil != err {
				logger.Errorf("Find match filename failed: [%s]", err.Error())

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
		logger.Errorf("Read dir [%s] failed: [%s]", dir, err.Error())

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
		logger.Errorf("Read file [%s] failed: [%s]", path, err.Error())

		return ret
	}

	content := string(bytes)
	if util.File.IsBinary(content) {
		return ret
	}

	lines := strings.Split(content, "\n")

	for idx, line := range lines {
		ch := strings.Index(strings.ToLower(line), strings.ToLower(text))

		if -1 != ch {
			snippet := &Snippet{Path: filepath.ToSlash(path),
				Line: idx + 1, Ch: ch + 1, Contents: []string{line}}

			ret = append(ret, snippet)
		}
	}

	return ret
}
