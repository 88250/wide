package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type myos struct{}

// 操作系统工具.
var OS = myos{}

// 判断是否是 Windows 操作系统.
func (*myos) IsWindows() bool {
	return "windows" == runtime.GOOS
}

// 获取当前执行程序的工作目录的绝对路径.
func (*myos) Pwd() string {
	file, _ := exec.LookPath(os.Args[0])
	pwd, _ := filepath.Abs(file)

	return filepath.Dir(pwd)
}
