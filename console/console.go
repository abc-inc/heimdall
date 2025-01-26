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

//go:build !dasel

package console

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/abc-inc/gutenfmt/gfmt"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/charmbracelet/gum/style"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var Version string

const (
	FileGroup     = "File Commands"
	MiscGroup     = "Misc Commands"
	ServiceGroup  = "Service Commands"
	SoftwareGroup = "Software Commands"
	HeimdallGroup = "Heimdall Commands"
)

type OutCfg struct {
	Output   string
	Query    string
	JQFilter string
	Pretty   bool
}

type Printer func(any) (int, error)

var Output io.Writer = os.Stdout
var printer Printer = func(a any) (int, error) {
	return fmt.Fprint(Output, a)
}

var Optional = func(cmd *cobra.Command) {}
var Required = func(name string) func(cmd *cobra.Command) {
	return func(cmd *cobra.Command) { internal.MustNoErr(cmd.MarkFlagRequired(name)) }
}

func AddFileFlag(cmd *cobra.Command, v *string, usage string, opts ...func(cmd *cobra.Command)) {
	cmd.Flags().StringVarP(v, "file", "f", *v, usage)
	if len(opts) == 0 {
		opts = append(opts, Required("file"))
	}
	for _, opt := range opts {
		opt(cmd)
	}
}

func AddOutputFlags(cmd *cobra.Command, cfg *OutCfg) {
	AddOutputFlag(cmd, &cfg.Output)
	addPrettyFlag(cmd, &cfg.Pretty)
	addQueryFlag(cmd, &cfg.Query)
	addJQFlag(cmd, &cfg.JQFilter)
}

func addJQFlag(cmd *cobra.Command, v *string) {
	cmd.PersistentFlags().StringVar(v, "jq", *v, "Specify a jq filter for modifying the output")
}

func AddOutputFlag(cmd *cobra.Command, v *string) {
	cmd.PersistentFlags().StringVarP(v, "output", "o", *v, "Output format (csv, json, table, template, template-file, text, tsv, yaml)")
}

func addPrettyFlag(cmd *cobra.Command, v *bool) {
	*v = isatty.IsTerminal(os.Stdout.Fd())
	cmd.PersistentFlags().BoolVar(v, "pretty", *v, "Pretty-print the output")
}

func addQueryFlag(cmd *cobra.Command, v *string) {
	cmd.PersistentFlags().StringVarP(v, "query", "q", *v, "Specify a JMESPath query to use in filtering the output")
}

func Fmt(a any) {
	if _, err := printer(a); err != nil {
		log.Fatal().Err(err).Msg("Cannot write output")
	}
}

func Fmtln(a any) {
	Fmt(a)
	_ = internal.Must(Msg("\n"))
}

func Msg(str string) (int, error) {
	return io.WriteString(Output, str)
}

func getWriter(f string, pretty bool) (w gfmt.Writer) {
	t, q, _ := strings.Cut(f, ":")
	switch t {
	case "csv":
		w = gfmt.NewText(Output)
		w.(*gfmt.Text).Sep = ","
	case "", "json":
		if pretty {
			w = gfmt.NewJSON(Output, gfmt.WithPretty())
		} else {
			w = gfmt.NewJSON(Output)
		}
	case "table":
		w = gfmt.NewTab(Output)
	case "template":
		tmpl := internal.Must(template.New("output").Parse(q)).Funcs(sprig.HermeticTxtFuncMap())
		w = gfmt.NewTemplate(Output, tmpl)
	case "template-file":
		r := internal.Must(res.Open(q))
		defer func() { _ = r.Close() }()
		pat := string(internal.Must(io.ReadAll(r)))
		tmpl := internal.Must(template.New("output").Parse(pat)).Funcs(sprig.HermeticTxtFuncMap())
		w = gfmt.NewTemplate(Output, tmpl)
	case "text":
		w = gfmt.NewText(Output)
		w.(*gfmt.Text).Sep = "="
	case "tsv":
		w = gfmt.NewText(Output)
		w.(*gfmt.Text).Sep = "\t"
	case "yaml":
		if pretty {
			w = gfmt.NewYAML(Output, gfmt.WithPretty())
		} else {
			w = gfmt.NewYAML(Output)
		}
	default:
		log.Fatal().Str("output", f).Msg("Invalid output format")
		os.Exit(1)
	}

	return
}

// Parse attempts to detect the input format e.g., JSON and returns the value,
// which could be a key-value pairs (map) or a slice thereof.
func Parse(r io.Reader) any {
	m := map[string]any{}
	in := bufio.NewScanner(r)
	for in.Scan() {
		s := in.Text()
		b := []byte(s)
		if json.Valid(b) {
			if b[0] == '[' {
				m2 := []any{}
				if err := json.Unmarshal(b, &m2); err != nil {
					log.Fatal().Err(err).Msg("Cannot output JSON")
				}
				m[""] = m2
			} else if err := json.Unmarshal(b, &m); err != nil {
				log.Fatal().Err(err).Msg("Cannot output JSON")
			}
		} else if idx := strings.IndexAny(s, "=:\t"); idx > 0 {
			m[s[:idx]] = s[idx+1:]
		}
	}
	if _, ok := m[""]; ok {
		return m[""]
	}
	return m
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

	pretty := strFunc(opts, "pretty")
	var w = getWriter(strFunc(opts, "output"), pretty == "true")
	if q := strFunc(opts, "query"); q != "" {
		printer = gfmt.NewJMESPath(w, q).Write
	} else if q = strFunc(opts, "jq"); q != "" {
		printer = gfmt.NewJQ(w, q).Write
	} else {
		printer = func(a any) (int, error) { return w.Write(a) }
	}
}

func Reset() {
	Output = os.Stdout
}

var StyleProps []string

func init() {
	t := reflect.TypeOf(style.Styles{})
	// Note that the env variables are inserted in reverse order.
	// This is important because viper performs lookup in lexicographic order.
	// Otherwise, viper.GetString("border-foreground") would return the value of BORDER instead of BORDER_FOREGROUND.
	for i := t.NumField() - 1; i >= 0; i-- {
		e := t.Field(i).Tag.Get("env")
		StyleProps = append(StyleProps, e)
	}
}
