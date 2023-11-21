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

package docker

import (
	"io"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/moby/buildkit/frontend/dockerfile/dockerignore"
	"github.com/spf13/cobra"
)

func NewIgnoreCmd() *cobra.Command {
	cfg := dockerCfg{file: ".dockerignore"}
	cmd := &cobra.Command{
		Use:   "ignore",
		Short: "Parse a .dockerignore file and print its patterns",
		Example: heredoc.Doc(`
			heimdall docker ignore --output text
		`),
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			console.Fmtln(readIgnore(cfg))
		},
	}

	console.AddFileFlag(cmd, &cfg.file, "Path to the .dockerignore file", console.Optional)
	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func readIgnore(cfg dockerCfg) []string {
	r := internal.Must(res.Open(cfg.file))
	defer func() { _ = r.Close() }()
	return parseIgnore(r)
}

func parseIgnore(r io.Reader) []string {
	return internal.Must(dockerignore.ReadAll(r))
}
