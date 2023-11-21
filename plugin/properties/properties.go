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

package properties

import (
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/magiconair/properties"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type propsCfg struct {
	console.OutCfg
	file    string
	get     []string
	set     []string
	unset   []string
	sep     string
	comment string
	enc     Enc
}

func NewPropertiesCmd() *cobra.Command {
	cfg := propsCfg{sep: "=", enc: UTF_8}

	cmd := &cobra.Command{
		Use:   "properties [flag] <file>",
		Short: "Load a properties file and process it",
		Example: heredoc.Doc(`
			heimdall properties --get name,email .git/config"
		`),
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.file = args[0]
			console.Fmtln(processProps(cfg))
		},
	}

	cmd.Flags().StringVarP(&cfg.comment, "comment", "c", "", `Write comments with prefix (example: "# ", default: (no comments))`)
	cmd.Flags().StringSliceVarP(&cfg.get, "get", "g", nil, "Print the value matching the keys. May be provided multiple times.")
	cmd.Flags().StringSliceVarP(&cfg.set, "set", "s", nil, "Set or replace key-value pairs. May be provided multiple times.")
	cmd.Flags().StringVarP(&cfg.sep, "separator", "t", cfg.sep, "Key-value separator")
	cmd.Flags().StringSliceVarP(&cfg.unset, "unset", "u", nil, "Unset properties matching the key. May be provided multiple times.")
	cmd.Flags().Var(&cfg.enc, "encoding", "File encoding")

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func processProps(cfg propsCfg) any {
	if cfg.enc == 0 {
		cfg.enc = UTF_8
	}
	if cfg.sep == "" {
		cfg.sep = "="
	}

	props := properties.MustLoadFile(cfg.file, properties.Encoding(cfg.enc))
	props.WriteSeparator = cfg.sep

	for _, k := range cfg.unset {
		props.Delete(k)
	}

	for _, kv := range cfg.set {
		k, v, ok := strings.Cut(kv, cfg.sep)
		if !ok {
			log.Fatal().Str("property", kv).Str("separator", cfg.sep).Msg("Invalid key-value pair")
		}
		_, _ = props.MustSet(k, v)
	}

	if cfg.Output == "text" {
		var ls []string
		if len(cfg.get) == 0 {
			for _, k := range props.Keys() {
				if cfg.comment != "" {
					for _, c := range props.GetComments(k) {
						ls = append(ls, cfg.comment+c)
					}
				}
				ls = append(ls, k+cfg.sep+props.MustGet(k))
			}
		} else {
			for _, k := range cfg.get {
				ls = append(ls, props.MustGet(k))
			}
		}
		return ls
	}

	if len(cfg.get) == 0 {
		cfg.get = props.Keys()
	}
	if cfg.comment != "" {
		var ls []pair
		for _, k := range cfg.get {
			ls = append(ls, pair{
				Key:     k,
				Value:   props.MustGet(k),
				Comment: strings.Join(props.GetComments(k), "\n"),
			})
		}
		return ls
	}

	ls := make(map[string]string)
	for _, k := range cfg.get {
		ls[k] = props.MustGet(k)
	}
	return ls
}

type pair struct {
	Key     string `json:"key" yaml:"key"`
	Value   string `json:"value" yaml:"value"`
	Comment string `json:"comment,omitempty" yaml:"comment,omitempty"`
}
