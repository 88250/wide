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

package file

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"strings"

	"github.com/b3log/wide/util"
)

type element struct {
	Name string
	Line int
	Ch   int
}

// GetOutlineHandler gets outfile of a go file.
func GetOutlineHandler(w http.ResponseWriter, r *http.Request) {
	result := util.NewResult()
	defer util.RetResult(w, r, result)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		result.Succ = false

		return
	}

	code := args["code"].(string)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", code, 0)
	if err != nil {
		result.Succ = false

		return
	}

	data := map[string]interface{}{}
	result.Data = &data

	// ast.Print(fset, f)

	line, ch := getCursor(code, int(f.Name.Pos()))
	data["package"] = &element{Name: f.Name.Name, Line: line, Ch: ch}

	imports := []*element{}
	for _, astImport := range f.Imports {
		line, ch := getCursor(code, int(astImport.Path.Pos()))

		imports = append(imports, &element{Name: astImport.Path.Value, Line: line, Ch: ch})
	}
	data["imports"] = imports

	funcDecls := []*element{}
	varDecls := []*element{}
	constDecls := []*element{}
	structDecls := []*element{}
	interfaceDecls := []*element{}
	typeDecls := []*element{}
	for _, decl := range f.Decls {
		switch decl.(type) {
		case *ast.FuncDecl:
			funcDecl := decl.(*ast.FuncDecl)

			line, ch := getCursor(code, int(funcDecl.Name.Pos()))

			funcDecls = append(funcDecls, &element{Name: funcDecl.Name.Name, Line: line, Ch: ch})
		case *ast.GenDecl:
			genDecl := decl.(*ast.GenDecl)

			for _, spec := range genDecl.Specs {

				switch genDecl.Tok {
				case token.VAR:
					variableSpec := spec.(*ast.ValueSpec)

					for _, varName := range variableSpec.Names {
						line, ch := getCursor(code, int(varName.Pos()))

						varDecls = append(varDecls, &element{Name: varName.Name, Line: line, Ch: ch})
					}
				case token.TYPE:
					typeSpec := spec.(*ast.TypeSpec)
					line, ch := getCursor(code, int(typeSpec.Pos()))

					switch typeSpec.Type.(type) {
					case *ast.StructType:
						structDecls = append(structDecls, &element{Name: typeSpec.Name.Name, Line: line, Ch: ch})
					case *ast.InterfaceType:
						interfaceDecls = append(interfaceDecls, &element{Name: typeSpec.Name.Name, Line: line, Ch: ch})
					case *ast.Ident:
						typeDecls = append(typeDecls, &element{Name: typeSpec.Name.Name, Line: line, Ch: ch})
					}
				case token.CONST:
					constSpec := spec.(*ast.ValueSpec)

					for _, constName := range constSpec.Names {
						line, ch := getCursor(code, int(constName.Pos()))

						constDecls = append(constDecls, &element{Name: constName.Name, Line: line, Ch: ch})
					}
				}
			}
		}
	}

	data["funcDecls"] = funcDecls
	data["varDecls"] = varDecls
	data["constDecls"] = constDecls
	data["structDecls"] = structDecls
	data["interfaceDecls"] = interfaceDecls
	data["typeDecls"] = typeDecls
}

// getCursor calculates the cursor position (line, ch) by the specified offset.
//
// line is the line number, starts with 0 that means the first line
// ch is the column number, starts with 0 that means the first column
func getCursor(code string, offset int) (line, ch int) {
	code = code[:offset]

	lines := strings.Split(code, "\n")

	line = 0
	for range lines {
		line++
	}

	var buffer bytes.Buffer
	runes := []rune(lines[line-1])
	for _, r := range runes {
		buffer.WriteString(string(r))
	}

	ch = len(buffer.String())

	return line - 1, ch - 1
}
