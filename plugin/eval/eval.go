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

//go:build !no_eval

package eval

import (
	"encoding/json"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/plugin/parse"
	"github.com/abc-inc/heimdall/res"
	sprig "github.com/go-task/slim-sprig/v3"
	"github.com/gobwas/glob"
	"github.com/mattn/go-zglob"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type engine interface {
	eval(cfg evalCfg, envMap map[string]any) ([]string, error)
	addFunc(funcMap map[string]any) error
}

type evalCfg struct {
	output   string
	engine   string
	expr     []string
	files    []string
	template string
	ignMiss  bool
	quiet    bool
	verbose  bool
}

var engines = make(map[string]func() engine)

const envHelp = `
CSV_DELIMITER           ,
HEIMDALL_TEMPLATE       
HEIMDALL_TEMPLATE_FILE  
`

func NewEvalCmd() *cobra.Command {
	names := maps.Keys(engines)
	slices.Sort(names)

	cfg := evalCfg{engine: "expr"}
	cmd := &cobra.Command{
		Use:   "eval [flags] [<file>...]",
		Short: "Evaluate the given expression on all input files",
		Long: heredoc.Doc(`
			Evaluate the given expression on all input files.
			The following file formats are supported: csv, json, properties, xml, yaml
		`),
		Example: heredoc.Doc(`
			# check whether the filename of the URL matches the given regular expression
			heimdall eval -e 'base(distributionUrl) matches "gradle-[6-9][.]"' gradle/wrapper/gradle-wrapper.properties

			# print all variable names from multiple files in lexical order
			# (if a variable is defined multiple times, the last definition takes precedence)
			heimdall eval --ignore-missing -e 'sortAlpha(keys(_))' ${GRADLE_USER_HOME:-~/.gradle}/gradle.properties gradle.properties

			# first, get the summary of the JaCoCo code coverage report in JSON format
			# then, feed it into the JavaScript interpreter to calculate coverage ratio
			# (note the "-::json", which means: take standard input ("-"), use no variable prefix (""), and treat it as json)
			heimdall java jacoco --summary jacoco.csv |
			    heimdall eval -E javascript -e 'line_covered / (line_covered + line_missed)' -- -::json
		`),
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			eval(cfg, args)
		},
	}

	cmd.DisableFlagsInUseLine = true
	cmd.Flags().StringVarP(&cfg.engine, "engine", "E", cfg.engine, `Engine to use ("`+strings.Join(names, `", "`)+`")`)
	cmd.Flags().StringArrayVarP(&cfg.expr, "expression", "e", cfg.expr, "Expression to evaluate against the input files. May be provided multiple times.")
	cmd.Flags().BoolVar(&cfg.ignMiss, "ignore-missing", cfg.ignMiss, "Don't fail or report status for missing files")
	cmd.Flags().BoolVar(&cfg.quiet, "quiet", false, "Enable quiet mode (suppress normal output)")
	cmd.Flags().BoolVarP(&cfg.verbose, "verbose", "v", false, "Enable verbose mode")

	cli.AddOutputFlag(cmd, &cfg.output)
	internal.MustNoErr(cmd.MarkFlagRequired("expression"))
	return cmd
}

func eval(cfg evalCfg, args []string) {
	if _, ok := engines[cfg.engine]; !ok {
		names := maps.Keys(engines)
		slices.Sort(names)
		log.Fatal().Msgf(`cannot find engine "%s", must be one of "%s"`, cfg.engine, strings.Join(names, `", "`))
	}
	if t := os.Getenv("HEIMDALL_TEMPLATE_FILE"); t != "" && cfg.template == "" {
		internal.MustNoErr(os.Setenv("HEIMDALL_TEMPLATE", string(internal.Must(os.ReadFile(t)))))
	}
	if t := os.Getenv("HEIMDALL_TEMPLATE"); t != "" && cfg.template == "" {
		cfg.template = t
	}

	cfg.files = args
	result, err := doEval(cfg)
	if err != nil {
		log.WithLevel(zerolog.FatalLevel).Err(err).Send()
		os.Exit(2)
	}
	if !cfg.quiet {
		tmpl := internal.Must(template.New("result").Parse(cfg.template))
		for _, r := range result {
			if cfg.template != "" {
				if i, convIntErr := strconv.Atoi(r); convIntErr == nil {
					internal.MustNoErr(tmpl.Execute(cli.IO.Out, i))
				} else if f, convFloatErr := strconv.ParseFloat(r, 64); convFloatErr == nil {
					internal.MustNoErr(tmpl.Execute(cli.IO.Out, f))
				} else if b, convBoolErr := strconv.ParseBool(r); convBoolErr == nil {
					internal.MustNoErr(tmpl.Execute(cli.IO.Out, b))
				} else {
					internal.MustNoErr(tmpl.Execute(cli.IO.Out, r))
				}
			} else {
				cli.Fmtln(r)
			}
		}
	}
	all := strings.Join(result, "\n")
	if all == "" || all == "0" || all == "false" {
		os.Exit(1)
	}
}

func doEval(cfg evalCfg) ([]string, error) {
	fs := resolveFiles(cfg.files)

	envMap := make(map[string]any)
	for _, f := range fs {
		if cfg.ignMiss && f.File != "-" {
			if fi, err := os.Stat(f.File); err != nil || !fi.Mode().IsRegular() {
				continue
			}
		}
		load(f, envMap)
	}
	if cfg.verbose {
		v := internal.Must(json.Marshal(envMap))
		log.Debug().RawJSON("vars", v).Msg("Initialized variables")
	}
	envMap["_"] = maps.Clone(envMap)

	e := engines[cfg.engine]()
	internal.MustNoErr(e.addFunc(map[string]any{"glob": func(p, s string) bool {
		return internal.Must(glob.Compile(p)).Match(s)
	}}))
	internal.MustNoErr(e.addFunc(map[string]any{"urlEncode": urlEncode, "urlDecode": urlDecode}))
	internal.MustNoErr(e.addFunc(sprig.GenericFuncMap()))

	return e.eval(cfg, envMap)
}

func urlDecode(str string) string { s, _ := url.QueryUnescape(str); return s }

func urlEncode(str string) string { return url.QueryEscape(str) }

func resolveFiles(fs []string) (list []parse.Input) {
	for _, f := range fs {
		n, post, _ := strings.Cut(f, ":")
		if n == "-" {
			list = append(list, parse.SplitNamePrefixType(f))
			continue
		}
		log.Debug().Str("glob", n).Msg("Resolving files")
		gs, err := zglob.Glob(n)
		internal.MustOkMsgf(gs, err == nil, "cannot resolve any files matching glob '%s'", n)
		for _, g := range internal.Must(gs, err) {
			list = append(list, parse.SplitNamePrefixType(g+":"+post))
		}
	}
	return list
}

func load(i parse.Input, envMap map[string]any) {
	log.Debug().Str("file", i.File).Msg("Loading")
	r := internal.Must(res.Open(i.File))
	defer func() { _ = r.Close() }()

	if d, ok := parse.Decoders[i.Typ]; ok {
		log.Debug().Str("type", i.Typ).Msg("Using decoder")
		v := internal.Must[any](d(r))
		if reflect.TypeOf(v).Kind() == reflect.Map {
			merge(envMap, i.Alias, v.(map[string]any))
		} else {
			envMap[i.Alias] = v
		}
	}
}

func merge(envMap map[string]any, alias string, varMap map[string]any) {
	if alias != "" {
		envMap[alias] = varMap
	} else {
		maps.Copy(envMap, varMap)
	}
}
