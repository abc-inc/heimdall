// Copyright 2025 The Heimdall authors
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

package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/abc-inc/gutenfmt/gfmt"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

var Version string

var IO *IOStreams
var writer gfmt.Writer
var options = map[string]string{
	"output": "json",
	"pretty": "false",
	"query":  "",
	"jq":     "",
}

func init() {
	IO = System()
}

func Fmt(a any) {
	if _, err := getWriter().Write(a); err != nil {
		log.Fatal().Err(err).Msg("Cannot write output")
	}
}

func Fmtln(a any) {
	Fmt(a)
	_ = internal.Must(Msg("\n"))
}

func Msg(str string) (int, error) {
	return io.WriteString(IO.Out, str)
}

func SetFormat(opts map[string]any) {
	var strFunc = func(opts map[string]any, key string) string {
		a, ok := opts[key]
		if !ok || a == nil {
			return ""
		}

		switch s := a.(type) {
		case string:
			return s
		case fmt.Stringer:
			return s.String()
		case *pflag.Flag:
			if s == nil {
				return ""
			}
			return s.Value.String()
		default:
			panic(s)
		}
	}
	var setOpt = func(k, v string) {
		if curr, ok := options[k]; ok && curr != v {
			options[k] = v
			writer = nil
		}
	}

	setOpt("output", strFunc(opts, "output"))
	setOpt("pretty", strFunc(opts, "pretty"))
	setOpt("query", strFunc(opts, "query"))
	setOpt("jq", strFunc(opts, "jq"))
}

var StyleProps []string

func getWriter() gfmt.Writer {
	if writer != nil {
		return writer
	}

	t, q, _ := strings.Cut(options["output"], ":")
	switch t {
	case "csv":
		writer = gfmt.NewText(IO.Out)
		writer.(*gfmt.Text).Sep = ","
	case "", "json":
		if options["pretty"] == "true" {
			writer = gfmt.NewJSON(IO.Out, gfmt.WithPretty())
		} else {
			writer = gfmt.NewJSON(IO.Out)
		}
	case "table":
		writer = gfmt.NewTab(IO.Out)
	case "template":
		tmpl := internal.Must(template.New("output").Parse(q)).Funcs(sprig.HermeticTxtFuncMap())
		writer = gfmt.NewTemplate(IO.Out, tmpl)
	case "template-file":
		r := internal.Must(res.Open(q))
		defer func() { _ = r.Close() }()
		pat := string(internal.Must(io.ReadAll(r)))
		tmpl := internal.Must(template.New("output").Parse(pat)).Funcs(sprig.HermeticTxtFuncMap())
		writer = gfmt.NewTemplate(IO.Out, tmpl)
	case "text":
		writer = gfmt.NewText(IO.Out)
		writer.(*gfmt.Text).Sep = "="
	case "tsv":
		writer = gfmt.NewText(IO.Out)
		writer.(*gfmt.Text).Sep = "\t"
	case "yaml":
		if options["pretty"] == "true" {
			writer = gfmt.NewYAML(IO.Out, gfmt.WithPretty())
		} else {
			writer = gfmt.NewYAML(IO.Out)
		}
	default:
		log.Fatal().Str("output", options["output"]).Msg("Invalid output format")
		os.Exit(1)
	}

	if q := options["query"]; q != "" {
		writer = gfmt.NewJMESPath(writer, q)
	} else if q = options["jq"]; q != "" {
		writer = gfmt.NewJQ(writer, q)
	}

	return writer
}
