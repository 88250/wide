package util

import (
	"runtime/debug"

	"github.com/golang/glog"
)

// Recover recovers a panic.
func Recover() {
	if re := recover(); re != nil {
		glog.Errorf("PANIC RECOVERED:\n %v, %s", re, debug.Stack())
	}
}
