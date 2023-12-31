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

package format

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/spf13/cobra"
)

type formatCfg struct {
	console.OutCfg
	file string
}

func NewFormatCmd() *cobra.Command {
	cfg := formatCfg{}
	cmd := &cobra.Command{
		Use:   "format",
		Short: "Convert the input to the given output format",
		Example: heredoc.Doc(`
			heimdall format --output table <./gradle/wrapper/gradle-wrapper.properties 

			env | heimdall format --output json
		`),
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			format(cfg)
		},
	}

	console.AddFileFlag(cmd, &cfg.file, "Path to the input file or URL", console.Optional)
	console.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkPersistentFlagRequired("output"))
	return cmd
}

func format(cfg formatCfg) {
	if cfg.file == "" {
		cfg.file = "-"
	}

	r := internal.Must(res.Open(cfg.file))
	defer func() { _ = r.Close() }()
	console.Fmtln(console.Parse(r))
}
