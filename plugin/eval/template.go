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

//go:build !no_eval_template

package eval

import (
	"reflect"
	"strings"
	"text/template"

	"github.com/abc-inc/heimdall/internal"
)

type tmplEngine struct {
	tmpl *template.Template
}

func newTmplEngine() engine {
	return &tmplEngine{template.New("base")}
}

func (e *tmplEngine) eval(cfg evalCfg, envMap map[string]any) (res []string, err error) {
	internal.MustNoErr(e.addFunc(envMap))
	for _, str := range cfg.expr {
		e.tmpl = template.Must(e.tmpl.Parse(str))
		w := strings.Builder{}
		internal.MustNoErr(e.tmpl.Execute(&w, envMap))
		res = append(res, w.String())
	}
	return
}

func (e *tmplEngine) addFunc(funcMap map[string]any) error {
	fns := make(map[string]any)
	for n, f := range funcMap {
		if reflect.TypeOf(f).Kind() == reflect.Func {
			fns[n] = f
		}
	}
	e.tmpl.Funcs(fns)
	return nil
}

func init() {
	engines["template"] = newTmplEngine
}
