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

//go:build !no_eval && !no_eval_javascript

package eval

import (
	"fmt"

	"github.com/abc-inc/heimdall/internal"
	"github.com/dop251/goja"
)

type gojaEngine struct {
	vm *goja.Runtime
}

func newGojaEngine() engine {
	return &gojaEngine{goja.New()}
}

func (e *gojaEngine) eval(cfg evalCfg, envMap map[string]any) (res []string, err error) {
	internal.MustNoErr(e.addFunc(envMap))
	for _, str := range cfg.expr {
		v, err := e.vm.RunString(str)
		if err != nil {
			return res, err
		}
		res = append(res, fmt.Sprint(v.Export()))
	}
	return res, nil
}

func (e *gojaEngine) addFunc(funcMap map[string]any) error {
	for n, f := range funcMap {
		if err := e.vm.Set(n, f); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	engines["javascript"] = newGojaEngine
}
