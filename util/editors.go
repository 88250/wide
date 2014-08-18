package util

import (
	"strings"
)

type editor struct{}

var Editor = editor{}

func (editor) GetCursorOffset(code string, line, ch int) (offset int) {
	lines := strings.Split(code, "\n")

	for i := 0; i < line; i++ {
		offset += len(lines[i])
	}

	offset += line + ch

	return
}
