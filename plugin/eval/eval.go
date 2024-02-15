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
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	sprig "github.com/go-task/slim-sprig/v3"
	"github.com/gobwas/glob"
	"github.com/mattn/go-zglob"
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

type input struct {
	file  string
	alias string
	typ   string
}

func (i input) String() string {
	return i.file + ":" + i.alias + ":" + i.typ
}

var engines = make(map[string]func() engine)

func NewEvalCmd() *cobra.Command {
	names := maps.Keys(engines)
	slices.Sort(names)

	cfg := evalCfg{engine: "expr"}
	cmd := &cobra.Command{
		Use:   "eval [flags] [<file>...]",
		Short: "Evaluate the given expression on all input files",
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
			if _, ok := engines[cfg.engine]; !ok {
				log.Fatal().Msgf(`cannot find engine "%s", must be one of "%s"`, cfg.engine, strings.Join(names, `", "`))
			}
			if t := os.Getenv("HEIMDALL_TEMPLATE"); t != "" && cfg.template != "" {
				cfg.template = t
			}

			cfg.files = args
			result, err := eval(cfg)
			if err != nil {
				log.Fatal().Err(err).Send()
			}
			if !cfg.quiet {
				tmpl := template.Must(template.New("result").Parse(cfg.template))
				for _, r := range result {
					if cfg.template != "" {
						if i, convIntErr := strconv.Atoi(r); convIntErr == nil {
							internal.MustNoErr(tmpl.Execute(console.Output, i))
						} else if f, convFloatErr := strconv.ParseFloat(r, 64); convFloatErr == nil {
							internal.MustNoErr(tmpl.Execute(console.Output, f))
						} else if b, convBoolErr := strconv.ParseBool(r); convBoolErr == nil {
							internal.MustNoErr(tmpl.Execute(console.Output, b))
						} else {
							internal.MustNoErr(tmpl.Execute(console.Output, r))
						}
					} else {
						console.Fmtln(r)
					}
				}
			}
			if len(result) > 0 && result[len(result)-1] == "false" {
				os.Exit(1)
			}
		},
	}

	cmd.DisableFlagsInUseLine = true
	cmd.Flags().StringVarP(&cfg.engine, "engine", "E", cfg.engine, `Engine to use ("`+strings.Join(names, `", "`)+`")`)
	cmd.Flags().StringArrayVarP(&cfg.expr, "expression", "e", cfg.expr, "Expression to evaluate against the input files. May be provided multiple times.")
	cmd.Flags().BoolVar(&cfg.ignMiss, "ignore-missing", cfg.ignMiss, "Don't fail or report status for missing files")
	cmd.Flags().BoolVar(&cfg.quiet, "quiet", false, "Enable quiet mode (suppress normal output)")
	cmd.Flags().StringVar(&cfg.template, "template", cfg.template, "Output template to use (Go template syntax).")
	cmd.Flags().BoolVarP(&cfg.verbose, "verbose", "v", false, "Enable verbose mode")

	console.AddOutputFlag(cmd, &cfg.output)
	internal.MustNoErr(cmd.MarkFlagRequired("expression"))
	return cmd
}

func eval(cfg evalCfg) ([]string, error) {
	fs := resolveFiles(cfg.files)

	envMap := make(map[string]any)
	for _, f := range fs {
		if cfg.ignMiss && f.file != "-" {
			if fi, err := os.Stat(f.file); err != nil || !fi.Mode().IsRegular() {
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
	internal.MustNoErr(e.addFunc(sprig.GenericFuncMap()))

	return e.eval(cfg, envMap)
}

func resolveFiles(fs []string) (list []input) {
	for _, f := range fs {
		n, post, _ := strings.Cut(f, ":")
		if n == "-" {
			list = append(list, splitNamePrefixType(f))
			continue
		}
		log.Debug().Str("glob", n).Msg("Resolving files")
		gs, err := zglob.Glob(n)
		internal.MustOkMsgf(gs, err == nil, "cannot resolve any files matching glob '%s'", n)
		for _, g := range internal.Must(gs, err) {
			list = append(list, splitNamePrefixType(g+":"+post))
		}
	}
	return list
}

func load(i input, envMap map[string]any) {
	log.Debug().Str("file", i.file).Msg("Loading")
	r := internal.Must(res.Open(i.file))
	defer func() { _ = r.Close() }()

	if d, ok := res.Decoders[i.typ]; ok {
		log.Debug().Str("type", i.typ).Msg("Using decoder")
		merge(i.alias, envMap, internal.Must(d(r)))
	}
}

func splitNamePrefixType(name string) input {
	parts := strings.SplitN(name, ":", 3)
	switch len(parts) {
	case 3:
		return input{file: parts[0], alias: parts[1], typ: parts[2]}
	case 2:
		return input{file: parts[0], alias: parts[1], typ: strings.TrimPrefix(path.Ext(parts[0]), ".")}
	default:
		return input{file: parts[0], alias: "", typ: strings.TrimPrefix(path.Ext(parts[0]), ".")}
	}
}

func merge(alias string, envMap, varMap map[string]any) {
	if alias != "" {
		envMap[alias] = varMap
	} else {
		maps.Copy(envMap, varMap)
	}
}
