// Copyright 2024 The Heimdall authors
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

//go:build !no_parse && !no_toml

package parse

import (
	"io"

	"github.com/BurntSushi/toml"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

type tomlCfg struct {
	cli.OutCfg
}

func NewTOMLCmd() *cobra.Command {
	cfg := tomlCfg{}

	cmd := &cobra.Command{
		Use:     "toml [flags] <file>...",
		Short:   "Load TOML files and process them",
		GroupID: cli.FileGroup,
		Example: heredoc.Doc(`
			heimdall toml --query 'libraries.junit.version' gradle/libs.versions.toml"
		`),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m := make(map[string]any)
			for _, f := range args {
				maps.Copy(m, ProcessTOML(f))
			}
			cli.Fmtln(m)
		},
	}

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	cmd.DisableFlagsInUseLine = true
	return cmd
}

func ProcessTOML(name string) (m map[string]any) {
	r := internal.Must(res.Open(name))
	defer func() { _ = r.Close() }()
	internal.Must(toml.NewDecoder(r).Decode(&m))
	return m
}

func init() {
	Decoders["toml"] = func(r io.Reader) (any, error) {
		var m map[string]any
		_, err := toml.NewDecoder(r).Decode(&m)
		return m, err
	}
}
