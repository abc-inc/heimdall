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

//go:build !no_parse && !no_json

package parse

import (
	"encoding/json"
	"io"
	"maps"
	"path/filepath"
	"reflect"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/spf13/cobra"
)

type jsonCfg struct {
	cli.OutCfg
}

func NewJSONCmd() *cobra.Command {
	cfg := jsonCfg{}

	cmd := &cobra.Command{
		Use:     "json [flags] <file>...",
		Short:   "Load JSON files and process them",
		GroupID: cli.FileGroup,
		Example: heredoc.Doc(`
		`),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m := make(map[string]any)
			for _, f := range args {
				v, err := ReadJSON(internal.Must(res.Open(f)))
				internal.MustNoErr(err)
				maps.Copy(m, toMap(v, filepath.Base(f)))
			}
			cli.Fmtln(m)
		},
	}

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	cmd.DisableFlagsInUseLine = true
	return cmd
}

func ReadJSON(r io.Reader) (m any, err error) {
	err = json.NewDecoder(r).Decode(&m)
	return
}

func ReadJSONMap(name string) (m map[string]any) {
	r := internal.Must(res.Open(name))
	defer func() { _ = r.Close() }()
	internal.MustNoErr(json.NewDecoder(r).Decode(&m))
	return
}

func toMap(a any, defKey string) map[string]any {
	if m, ok := a.(map[string]any); ok {
		return m
	}
	return map[string]any{defKey: a}
}

func toAnyMap(a any, defKey string) any {
	if reflect.TypeOf(a).Kind() == reflect.Map {
		return a
	}
	return map[string]any{defKey: a}
}

func init() {
	Decoders["json"] = ReadJSON
}
