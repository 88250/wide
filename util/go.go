package util

import (
	"path"
	"runtime"
	"strings"
)

type mygo struct{}

// Go utilities.
var Go = mygo{}

// GetAPIPath gets the Go source code path.
//
//  1. before Go 1.4: $GOROOT/src/pkg
//  2. Go 1.4 and after: $GOROOT/src
func (*mygo) GetAPIPath() string {
	ret := runtime.GOROOT() + "/src/pkg" // before Go 1.4
	if !File.IsExist(ret) {
		ret = runtime.GOROOT() + "/src" // Go 1.4 and after
	}

	return path.Clean(ret)
}

// IsAPI determines whether the specified path belongs to Go API.
func (*mygo) IsAPI(path string) bool {
	apiPath := Go.GetAPIPath()

	return strings.HasPrefix(path, apiPath)
}
