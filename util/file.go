package util

import (
	"os"
	"strings"

	"github.com/golang/glog"
)

type myfile struct{}

// File utilities.
var File = myfile{}

// GetFileSize get the length in bytes of file of the specified path.
func (*myfile) GetFileSize(path string) int64 {
	f, err := os.Open(path)
	if nil != err {
		return -1
	}

	fi, err := f.Stat()
	if nil != err {
		return -1
	}

	return fi.Size()
}

// IsExist determines whether the file spcified by the given path is exists.
func (*myfile) IsExist(path string) bool {
	_, err := os.Stat(path)

	return err == nil || os.IsExist(err)
}

// IsBinary determines whether the specified content is a binary file content.
func (*myfile) IsBinary(content string) bool {
	for _, b := range content {
		if 0 == b {
			return true
		}
	}

	return false
}

// IsImg determines whether the specified extension is a image.
func (*myfile) IsImg(extension string) bool {
	ext := strings.ToLower(extension)

	switch ext {
	case ".jpg", ".jpeg", ".bmp", ".gif", ".png", ".svg", ".ico":
		return true
	default:
		return false
	}
}

// IsDir determines whether the specified path is a directory.
func (*myfile) IsDir(path string) bool {
	fio, err := os.Lstat(path)
	if nil != err {
		glog.Warningf("Determines whether [%s] is a directory failed: [%v]", path, err)

		return false
	}

	return fio.IsDir()
}
