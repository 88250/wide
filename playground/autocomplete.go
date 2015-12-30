// Copyright (c) 2014-2016, b3log.org & hacpai.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package playground

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/b3log/wide/session"
	"github.com/b3log/wide/util"
)

// AutocompleteHandler handles request of code autocompletion.
func AutocompleteHandler(w http.ResponseWriter, r *http.Request) {
	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	session, _ := session.HTTPSession.Get(r, "wide-session")
	if session.IsNew {
		http.Error(w, "Forbidden", http.StatusForbidden)

		return
	}

	code := args["code"].(string)
	line := int(args["cursorLine"].(float64))
	ch := int(args["cursorCh"].(float64))

	offset := getCursorOffset(code, line, ch)

	argv := []string{"-f=json", "autocomplete", strconv.Itoa(offset)}
	gocode := util.Go.GetExecutableInGOBIN("gocode")
	cmd := exec.Command(gocode, argv...)

	stdin, _ := cmd.StdinPipe()
	stdin.Write([]byte(code))
	stdin.Close()

	output, err := cmd.CombinedOutput()
	if nil != err {
		logger.Error(err)
		http.Error(w, err.Error(), 500)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

// getCursorOffset calculates the cursor offset.
//
// line is the line number, starts with 0 that means the first line
// ch is the column number, starts with 0 that means the first column
func getCursorOffset(code string, line, ch int) (offset int) {
	lines := strings.Split(code, "\n")

	// calculate sum length of lines before
	for i := 0; i < line; i++ {
		offset += len(lines[i])
	}

	// calculate length of the current line and column
	curLine := lines[line]
	var buffer bytes.Buffer
	r := []rune(curLine)
	for i := 0; i < ch; i++ {
		buffer.WriteString(string(r[i]))
	}

	offset += len(buffer.String()) // append length of current line
	offset += line                 // append number of '\n'

	return offset
}
