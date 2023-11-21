// Copyright 2023 The Heimdall authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/abc-inc/heimdall/internal"
	"github.com/mattn/go-zglob"
)

func main() {
	if len(os.Args) != 3 {
		internal.Must(fmt.Fprintf(os.Stderr, "Usage: %s <SRC_FILE> <OUT_DIR>\n", filepath.Base(os.Args[0])))
		os.Exit(1)
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}

	srcFile, outDir := os.Args[1], os.Args[2]
	srcFile = strings.ReplaceAll(srcFile, `\*`, `*`)
	cnt := strings.Count(srcFile, string(filepath.Separator)) + 1
	join := filepath.Join(gopath, "pkg", "mod", srcFile)
	for _, f := range internal.Must(zglob.Glob(join)) {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		parts := strings.Split(f, string(filepath.Separator))
		parseFile(f, filepath.Join(outDir, filepath.Join(parts[(len(parts)-cnt):len(parts)-1]...)))
	}
}

func parseFile(srcFile, destFile string) {
	fset := token.NewFileSet()
	src := internal.Must(os.ReadFile(srcFile))
	f := internal.Must(parser.ParseFile(fset, filepath.Base(srcFile), src, parser.ParseComments))

	for _, s := range f.Decls {
		if d, ok := s.(*ast.FuncDecl); ok && d.Recv != nil {
			var t string
			if typ, ok := d.Recv.List[0].Type.(*ast.StarExpr); ok {
				t = fmt.Sprint(typ.X)
			} else if typ, ok := d.Recv.List[0].Type.(*ast.Ident); ok {
				t = typ.String()
			}

			n := filepath.Join(destFile, t, d.Name.String()+".txt")
			internal.MustNoErr(os.MkdirAll(filepath.Dir(n), 0755))
			internal.MustNoErr(os.WriteFile(n, []byte(d.Doc.Text()), 0600))
		}
	}
}
