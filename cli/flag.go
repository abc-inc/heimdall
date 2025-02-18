// Copyright 2025 The Heimdall authors
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

package cli

import (
	"os"

	"github.com/abc-inc/heimdall/internal"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

const (
	FileGroup     = "File Commands"
	MiscGroup     = "Misc Commands"
	ServiceGroup  = "Service Commands"
	SoftwareGroup = "Software Commands"
	HeimdallGroup = "Heimdall Commands"
)

type OutCfg struct {
	Output   string
	Query    string
	JQFilter string
	Pretty   bool
}

var Optional = func(cmd *cobra.Command) {}
var Required = func(name string) func(cmd *cobra.Command) {
	return func(cmd *cobra.Command) { internal.MustNoErr(cmd.MarkFlagRequired(name)) }
}

func AddFileFlag(cmd *cobra.Command, v *string, usage string, opts ...func(cmd *cobra.Command)) {
	cmd.Flags().StringVarP(v, "file", "f", *v, usage)
	if len(opts) == 0 {
		opts = append(opts, Required("file"))
	}
	for _, opt := range opts {
		opt(cmd)
	}
}

func AddOutputFlags(cmd *cobra.Command, cfg *OutCfg) {
	AddOutputFlag(cmd, &cfg.Output)
	addPrettyFlag(cmd, &cfg.Pretty)
	addQueryFlag(cmd, &cfg.Query)
	addJQFlag(cmd, &cfg.JQFilter)
}

func addJQFlag(cmd *cobra.Command, v *string) {
	cmd.PersistentFlags().StringVar(v, "jq", *v, "Specify a jq filter for modifying the output")
}

func AddOutputFlag(cmd *cobra.Command, v *string) {
	cmd.PersistentFlags().StringVarP(v, "output", "o", *v, "Output format (csv, json, table, template, template-file, text, tsv, yaml)")
}

func addPrettyFlag(cmd *cobra.Command, v *bool) {
	*v = isatty.IsTerminal(os.Stdout.Fd())
	cmd.PersistentFlags().BoolVar(v, "pretty", *v, "Pretty-print the output")
}

func addQueryFlag(cmd *cobra.Command, v *string) {
	cmd.PersistentFlags().StringVarP(v, "query", "q", *v, "Specify a JMESPath query to use in filtering the output")
}
