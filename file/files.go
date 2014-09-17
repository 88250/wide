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
	"github.com/b3log/wide/user"
	"github.com/b3log/wide/util"
	"github.com/golang/glog"
)

// 构造用户工作空间文件树.
// 将 Go API 源码包（$GOROOT/src/pkg）也作为子节点，这样能方便用户查看 Go API 源码.
func GetFiles(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	session, _ := user.HTTPSession.Get(r, "wide-session")

	username := session.Values["username"].(string)
	userSrc := conf.Wide.GetUserWorkspace(username) + string(os.PathSeparator) + "src"

	root := FileNode{Name: "projects", Path: userSrc, IconSkin: "ico-ztree-dir ", Type: "d", FileNodes: []*FileNode{}}

	// 构造 Go API 节点
	apiPath := runtime.GOROOT() + string(os.PathSeparator) + "src" + string(os.PathSeparator) + "pkg"
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

	// 构造用户工作空间文件树
	walk(userSrc, &root)

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

// 文件节点，用于构造文件树.
type FileNode struct {
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	IconSkin  string      `json:"iconSkin"` // 值的末尾应该有一个空格
	Type      string      `json:"type"`     // "f"：文件，"d"：文件夹
	Mode      string      `json:"mode"`
	FileNodes []*FileNode `json:"children"`
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
	case ".js", "json":
		return "ico-ztree-js "
	case ".xml":
		return "ico-ztree-xml "
	default:
		return "ico-ztree-other "
	}
}

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

func isImg(extension string) bool {
	ext := strings.ToLower(extension)

	switch ext {
	case ".jpg", ".jpeg", ".bmp", ".gif", ".png", ".svg", ".ico":
		return true
	default:
		return false
	}
}
