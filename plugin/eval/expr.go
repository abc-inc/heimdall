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

//go:build !no_eval_expr

package eval

import (
	"fmt"
	"strings"

	"github.com/abc-inc/heimdall/internal"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/conf"
)

type exprEngine struct {
	funcMap map[string]any
	opts    []expr.Option
}

func newExprEngine() engine {
	return &exprEngine{make(map[string]any), nil}
}

func (e *exprEngine) eval(cfg evalCfg, envMap map[string]any) (res []string, err error) {
	internal.MustNoErr(e.addFunc(envMap))
	e.AddOpt(expr.Env(e.funcMap))
	for _, str := range cfg.expr {
		e.AddOpt(AllowUndefinedVariables(!strings.Contains(str, "??")))

		prg := internal.Must(expr.Compile(str, e.opts...))
		out := internal.Must(expr.Run(prg, e.funcMap))
		res = append(res, fmt.Sprint(out))
	}
	return
}

func (e *exprEngine) addFunc(funcMap map[string]any) error {
	for n, f := range funcMap {
		e.funcMap[n] = f
	}
	return nil
}

func (e *exprEngine) AddOpt(opt expr.Option) {
	e.opts = append(e.opts, opt)
}

func init() {
	engines["expr"] = newExprEngine
}

func AllowUndefinedVariables(strict bool) expr.Option {
	return func(c *conf.Config) {
		c.Strict = strict
	}
}
