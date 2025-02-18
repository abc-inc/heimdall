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

//go:build !no_golang

package golang

import (
	"encoding/json"
	"io"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

type modCfg struct {
	cli.OutCfg
	file string
}

func NewModCmd() *cobra.Command {
	cfg := modCfg{file: "go.mod"}
	cmd := &cobra.Command{
		Use:   "mod",
		Short: "Parse a Go module and print its content",
		Example: heredoc.Doc(`
			heimdall go mod --output text --query "Go.Version"
		`),
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			cli.Fmtln(readMod(cfg))
		},
	}

	cli.AddFileFlag(cmd, &cfg.file, "Path to the the go.mod file", cli.Optional)
	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func readMod(cfg modCfg) (m map[string]any) {
	r := internal.Must(res.Open(cfg.file))
	defer func() { _ = r.Close() }()

	f := parseMod(cfg.file, r)
	bs := internal.Must(json.Marshal(f))
	internal.MustNoErr(json.Unmarshal(bs, &m))
	removeAllKeys(m, "Syntax")
	return
}

func parseMod(file string, r io.Reader) *modfile.File {
	data := internal.Must(io.ReadAll(r))
	return internal.Must(modfile.Parse(file, data, nil))
}

func removeAllKeys(m map[string]any, del string) {
	delete(m, del)
	for k := range m {
		switch v := m[k].(type) {
		case map[string]any:
			removeAllKeys(v, del)
		case []any:
			for _, it := range v {
				if nested, ok := it.(map[string]any); ok {
					removeAllKeys(nested, del)
				}
			}
		default:
			// ignore primitive types
		}
	}
}
