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
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"

	"github.com/abc-inc/gutenfmt/gfmt"
	"github.com/abc-inc/heimdall/internal"
	"github.com/cli/go-gh/v2/pkg/jq"
)

type JQWriter struct {
	w      gfmt.Writer
	Expr   string
	Indent string
	Color  bool
}

func NewJQPrinter(w gfmt.Writer, expr string) gfmt.Writer {
	return &JQWriter{
		w:      w,
		Expr:   expr,
		Indent: "",
		Color:  false,
	}
}

func (w JQWriter) Write(i interface{}) (int, error) {
	var r io.Reader
	switch x := i.(type) {
	case io.Reader:
		r = x
	case string:
		r = strings.NewReader(x)
	case []byte:
		r = bytes.NewReader(x)
	case []string, map[string]string, map[string]any:
		b := internal.Must(json.Marshal(i))
		r = bytes.NewReader(b)
	default:
		typ := reflect.TypeOf(i)
		if canMarshal(typ) {
			b := internal.Must(json.Marshal(i))
			r = bytes.NewReader(b)
		} else if typ.Kind() == reflect.Slice && canMarshal(typ.Elem()) {
			b := internal.Must(json.Marshal(i))
			r = bytes.NewReader(b)
		} else if typ.Kind() == reflect.Map && canMarshal(typ.Key()) && canMarshal(typ.Elem()) {
			b := internal.Must(json.Marshal(i))
			r = bytes.NewReader(b)
		} else {
			internal.MustOkMsgf(x, false, "Cannot query json of type %s", typ.Kind())
		}
	}

	b := bytes.Buffer{}
	internal.MustNoErr(jq.EvaluateFormatted(r, &b, w.Expr, w.Indent, w.Color))
	str := b.String()
	if strings.HasPrefix(str, "[") || strings.HasPrefix(str, "{") {
		if err := json.Unmarshal(b.Bytes(), &i); err == nil {
			return w.w.Write(i)
		}
	}
	return Msg(strings.TrimSuffix(str, "\n"))
}

var marshalType = make(map[string]bool)

func canMarshal(t reflect.Type) bool {
	if s, ok := marshalType[t.String()]; ok {
		return s
	}
	if t.Kind() == reflect.Ptr {
		return canMarshal(t.Elem())
	}
	if t.Kind() != reflect.Struct {
		return false
	}

	marshalType[t.String()] = false
	for i := 0; i < t.NumField(); i++ {
		if _, ok := t.Field(i).Tag.Lookup("json"); ok {
			EnableMarshalling(t)
			break
		}
	}
	return marshalType[t.String()]
}

func EnableMarshalling(t reflect.Type) {
	marshalType[t.String()] = true
}
