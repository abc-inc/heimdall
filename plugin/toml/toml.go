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

package toml

import (
	"github.com/BurntSushi/toml"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

type tomlCfg struct {
	console.OutCfg
}

func NewTOMLCmd() *cobra.Command {
	cfg := tomlCfg{}

	cmd := &cobra.Command{
		Use:   "toml [flags] <file>...",
		Short: "Load TOML files and process them",
		Example: heredoc.Doc(`
			heimdall toml --query 'libraries.junit.version' gradle/libs.versions.toml"
		`),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m := make(map[string]any)
			for _, f := range args {
				maps.Copy(m, processTOML(f))
			}
			console.Fmtln(m)
		},
	}

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	cmd.DisableFlagsInUseLine = true
	return cmd
}

func processTOML(name string) (m map[string]any) {
	r := internal.Must(res.Open(name))
	defer func() { _ = r.Close() }()
	internal.Must(toml.NewDecoder(r).Decode(&m))
	return m
}
