package util

import (
	"runtime/debug"

	"github.com/golang/glog"
)

// panic 恢复.
func Recover() {
	if re := recover(); re != nil {
		glog.Errorf("PANIC RECOVERED:\n %v, %s", re, debug.Stack())
	}
}
