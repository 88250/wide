package util

import (
	"runtime"
)

type myos struct{}

// 操作系统工具.
var OS = myos{}

// 判断是否是 Windows 操作系统.
func (*myos) IsWindows() bool {
	return "windows" == runtime.GOOS
}
