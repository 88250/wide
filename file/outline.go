// Copyright (c) 2014, B3log
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
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"

	"github.com/b3log/wide/util"
)

type element struct {
	Name string
	Pos  token.Pos
	End  token.Pos
}

// GetOutline gets outfile of a go file.
func GetOutline(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"succ": true}
	defer util.RetJSON(w, r, data)

	var args map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		logger.Error(err)
		data["succ"] = false

		return
	}

	code := args["code"].(string)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", code, 0)
	if err != nil {
		data["succ"] = false

		return
	}

	//ast.Print(fset, f)

	data["package"] = &element{Name: f.Name.Name, Pos: f.Name.Pos(), End: f.Name.End()}

	imports := []*element{}
	for _, astImport := range f.Imports {

		impt := &element{Name: astImport.Path.Value, Pos: astImport.Path.Pos(), End: astImport.Path.End()}

		imports = append(imports, impt)
	}
	data["imports"] = imports

	funcDecls := []*element{}
	varDecls := []*element{}
	constDecls := []*element{}
	structDecls := []*element{}
	interfaceDecls := []*element{}
	for _, decl := range f.Decls {
		switch decl.(type) {
		case *ast.FuncDecl:
			funcDecl := decl.(*ast.FuncDecl)

			decl := &element{Name: funcDecl.Name.Name, Pos: funcDecl.Name.Pos(), End: funcDecl.Name.End()}

			funcDecls = append(funcDecls, decl)
		case *ast.GenDecl:
			genDecl := decl.(*ast.GenDecl)

			for _, spec := range genDecl.Specs {

				switch genDecl.Tok {
				case token.VAR:
					variableSpec := spec.(*ast.ValueSpec)
					decl := &element{Name: variableSpec.Names[0].Name, Pos: variableSpec.Pos(), End: variableSpec.End()}

					varDecls = append(varDecls, decl)
				case token.TYPE:
					typeSpec := spec.(*ast.TypeSpec)
					decl := &element{Name: typeSpec.Name.Name, Pos: typeSpec.Name.Pos(), End: typeSpec.Name.End()}

					switch typeSpec.Type.(type) {
					case *ast.StructType:
						structDecls = append(structDecls, decl)
					case *ast.InterfaceType:
						interfaceDecls = append(interfaceDecls, decl)
					}
				case token.CONST:
					constSpec := spec.(*ast.ValueSpec)
					decl := &element{Name: constSpec.Names[0].Name, Pos: constSpec.Pos(), End: constSpec.End()}

					constDecls = append(constDecls, decl)
				}
			}

		}
	}

	data["funcDecls"] = funcDecls
	data["varDecls"] = varDecls
	data["constDecls"] = constDecls
	data["structDecls"] = structDecls
	data["interfaceDecls"] = interfaceDecls
}
