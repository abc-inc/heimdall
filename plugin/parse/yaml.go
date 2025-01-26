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

//go:build !no_parse && !no_yaml

package parse

import (
	"io"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

type yamlCfg struct {
	console.OutCfg
}

func NewYAMLCmd() *cobra.Command {
	cfg := yamlCfg{}

	cmd := &cobra.Command{
		Use:     "yaml [flags] <file>...",
		Short:   "Load YAML files and process them",
		GroupID: console.FileGroup,
		Example: heredoc.Doc(`
			heimdall yaml --query 'to_number("web-app"."-version")' config.yaml"
		`),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m := make(map[string]any)
			for _, f := range args {
				maps.Copy(m, ProcessYAML(f))
			}
			console.Fmtln(m)
		},
	}

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	cmd.DisableFlagsInUseLine = true
	return cmd
}

func ProcessYAML(name string) (m map[string]any) {
	r := internal.Must(res.Open(name))
	defer func() { _ = r.Close() }()
	internal.MustNoErr(yaml.NewDecoder(r).Decode(&m))
	return m
}

func init() {
	Decoders["yaml"] = func(r io.Reader) (m any, err error) {
		d := yaml.NewDecoder(r)
		err = d.Decode(&m)
		return
	}
	Decoders["yml"] = Decoders["yaml"]
}
