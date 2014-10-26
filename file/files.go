// 文件树操作.
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
	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

// 文件节点，用于构造文件树.
type FileNode struct {
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	IconSkin  string      `json:"iconSkin"` // 值的末尾应该有一个空格
	Type      string      `json:"type"`     // "f"：文件，"d"：文件夹
	Mode      string      `json:"mode"`
	FileNodes []*FileNode `json:"children"`
}

// 代码片段. 这个结构可用于“查找使用”、“文件搜索”等的返回值.
type Snippet struct {
	Path     string   `json:"path"`     // 文件路径
	Line     int      `json:"line"`     // 行号
	Ch       int      `json:"ch"`       // 列号
	Contents []string `json:"contents"` // 附近几行
}

// 构造用户工作空间文件树.
//
// 将 Go API 源码包（$GOROOT/src/pkg）也作为子节点，这样能方便用户查看 Go API 源码.
func GetFiles(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	session, _ := session.HTTPSession.Get(r, "wide-session")

	username := session.Values["username"].(string)
	userWorkspace := conf.Wide.GetUserWorkspace(username)
	workspaces := strings.Split(userWorkspace, conf.PathListSeparator)

	root := FileNode{Name: "root", Path: "", IconSkin: "ico-ztree-dir ", Type: "d", FileNodes: []*FileNode{}}

	// 工作空间节点处理
	for _, workspace := range workspaces {
		workspacePath := workspace + conf.PathSeparator + "src"

		workspaceNode := FileNode{Name: workspace[strings.LastIndex(workspace, conf.PathSeparator)+1:] + " (" +
			workspace + ")",
			Path: workspacePath, IconSkin: "ico-ztree-dir ", Type: "d", FileNodes: []*FileNode{}}

		walk(workspacePath, &workspaceNode)

		// 添加工作空间节点
		root.FileNodes = append(root.FileNodes, &workspaceNode)
	}

	// 构造 Go API 节点
	apiPath := runtime.GOROOT() + conf.PathSeparator + "src" + conf.PathSeparator + "pkg"
	apiNode := FileNode{Name: "Go API", Path: apiPath, FileNodes: []*FileNode{}}

	goapiBuildOKSignal := make(chan bool)
	go func() {
		apiNode.Type = "d"
		// TOOD: Go API 用另外的样式
		apiNode.IconSkin = "ico-ztree-dir "

		walk(apiPath, &apiNode)

		// 放行信号
		close(goapiBuildOKSignal)
	}()

	// 等待放行
	<-goapiBuildOKSignal

	// 添加 Go API 节点
	root.FileNodes = append(root.FileNodes, &apiNode)

	data["root"] = root
}

// 编辑器打开一个文件.
func GetFile(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	path := args["path"].(string)
	buf, _ := ioutil.ReadFile(path)

	extension := filepath.Ext(path)

	// 通过文件扩展名判断是否是图片文件（图片在浏览器里新建 tab 打开）
	if isImg(extension) {
		data["mode"] = "img"

		path2 := strings.Replace(path, "\\", "/", -1)
		idx := strings.Index(path2, "/data/user_workspaces")
		data["path"] = path2[idx:]

		return
	}

	isBinary := false
	// 判断是否是其他二进制文件
	for _, b := range buf {
		if 0 == b { // 包含 0 字节就认为是二进制文件
			isBinary = true
		}
	}

	if isBinary {
		// 是二进制文件的话前端编辑器不打开
		data["succ"] = false
		data["msg"] = "Can't open a binary file :("
	} else {
		data["content"] = string(buf)
		data["mode"] = getEditorMode(extension)
		data["path"] = path
	}
}

// 保存文件.
func SaveFile(w http.ResponseWriter, r *http.Request) {
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
}

// 新建文件/目录.
func NewFile(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	path := args["path"].(string)
	fileType := args["fileType"].(string)

	if !createFile(path, fileType) {
		data["succ"] = false

		return
	}

	if "f" == fileType {
		extension := filepath.Ext(path)
		data["mode"] = getEditorMode(extension)
	}
}

// 删除文件/目录.
func RemoveFile(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	decoder := json.NewDecoder(r.Body)

	var args map[string]interface{}

	if err := decoder.Decode(&args); err != nil {
		glog.Error(err)
		data["succ"] = false

		return
	}

	path := args["path"].(string)

	if !removeFile(path) {
		data["succ"] = false
	}
}

// 在目录中搜索包含指定字符串的文件.
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

// 遍历指定的路径，构造文件树.
func walk(path string, node *FileNode) {
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
			child.IconSkin = "ico-ztree-dir "

			walk(fpath, &child)
		} else {
			child.Type = "f"
			ext := filepath.Ext(fpath)

			child.IconSkin = getIconSkin(ext)
			child.Mode = getEditorMode(ext)
		}
	}

	return
}

// 列出 dirname 指定目录下的文件/目录名.
func listFiles(dirname string) []string {
	f, _ := os.Open(dirname)

	names, _ := f.Readdirnames(-1)
	f.Close()

	sort.Strings(names)

	dirs := []string{}
	files := []string{}

	// 排序：目录靠前，文件靠后
	for _, name := range names {
		fio, _ := os.Lstat(filepath.Join(dirname, name))

		if fio.IsDir() {
			// 排除 .git 目录
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

// 根据文件后缀获取文件树图标 CSS 类名.
//
// CSS 类名可参考 zTree 文档.
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

// 根据文件后缀获取编辑器 mode.
//
// 编辑器 mode 可参考 CodeMirror 文档.
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

// 在 path 指定的路径上创建文件.
//
// fileType:
//
//  "f": 文件
//  "d": 目录
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

// 删除 path 指定路径的文件或目录.
func removeFile(path string) bool {
	if err := os.RemoveAll(path); nil != err {
		glog.Errorf("Removes [%s] failed: [%s]", path, err.Error())

		return false
	}

	glog.Infof("Removed [%s]", path)

	return true
}

// 在 dir 指定的目录（包含子目录）中的 extension 指定后缀的文件中搜索包含 text 文本的文件，类似 grep/findstr 命令.
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
			// 进入目录递归
			snippets = search(path, extension, text, snippets)
		} else if strings.HasSuffix(path, extension) {
			// 在文件中进行搜索
			ss := searchInFile(path, text)

			snippets = append(snippets, ss...)
		}
	}

	return snippets
}

// 在 path 指定的文件内容中搜索 text 指定的文本.
func searchInFile(path string, text string) []*Snippet {
	ret := []*Snippet{}

	bytes, err := ioutil.ReadFile(path)
	if nil != err {
		glog.Errorf("Read file [%s] failed: [%s]", path, err.Error())

		return ret
	}

	content := string(bytes)
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

// 根据文件名后缀判断是否是图片文件.
func isImg(extension string) bool {
	ext := strings.ToLower(extension)

	switch ext {
	case ".jpg", ".jpeg", ".bmp", ".gif", ".png", ".svg", ".ico":
		return true
	default:
		return false
	}
}
