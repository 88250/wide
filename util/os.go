package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type myos struct{}

// OS utilities.
var OS = myos{}

// IsWindows determines whether current OS is Windows.
func (*myos) IsWindows() bool {
	return "windows" == runtime.GOOS
}

// Pwd gets the path of current working directory.
func (*myos) Pwd() string {
	file, _ := exec.LookPath(os.Args[0])
	pwd, _ := filepath.Abs(file)

	return filepath.Dir(pwd)
}
