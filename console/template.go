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
	"io"
	"strings"
	"text/template"

	"github.com/abc-inc/gutenfmt/formatter"
	"github.com/abc-inc/gutenfmt/gfmt"
	sprig "github.com/go-task/slim-sprig/v3"
)

type Tmpl struct {
	out       io.Writer
	tmpl      *template.Template
	Formatter *formatter.CompFormatter
}

func NewTemplate(output io.Writer, p string) gfmt.Writer {
	p = strings.ReplaceAll(strings.ReplaceAll(p, "<<", "{{"), ">>", "}}")
	tmpl := template.Must(template.New("output").Parse(p))
	tmpl.Funcs(sprig.HermeticTxtFuncMap())
	return &Tmpl{
		out:       output,
		tmpl:      tmpl,
		Formatter: formatter.NewComp(),
	}
}

func (w Tmpl) Write(a any) (int, error) {
	return 0, w.tmpl.Execute(w.out, a)
}
