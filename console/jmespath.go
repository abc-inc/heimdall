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

package console

import (
	"reflect"

	"github.com/abc-inc/gutenfmt/gfmt"
	"github.com/jmespath/go-jmespath"
)

type JMESPathWriter struct {
	w      gfmt.Writer
	Expr   *jmespath.JMESPath
	Indent string
	Color  bool
}

func NewJMESPathPrinter(w gfmt.Writer, expr string) gfmt.Writer {
	return &JMESPathWriter{
		w:      w,
		Expr:   jmespath.MustCompile(expr),
		Indent: "",
		Color:  false,
	}
}

func (w JMESPathWriter) Write(i any) (n int, err error) {
	k := reflect.TypeOf(i).Kind()
	switch k {
	case reflect.Ptr:
		return w.Write(reflect.ValueOf(i).Elem())
	case reflect.Map, reflect.Struct, reflect.Slice, reflect.Array:
		i, err = w.Expr.Search(i)
		if err != nil {
			return 0, err
		}
		return w.w.Write(i)
	case reflect.String:
		return w.w.Write(i)
	default:
		panic("unsupported type: " + reflect.TypeOf(i).String())
	}
}
