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

//go:build !no_github

//go:generate go run github.com/abc-inc/heimdall/cmd/cmddoc github.com/google/go-github/v69@v69.2.0/github/markdown.go ../../docs

package github

import (
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/google/go-github/v69/github"
	"github.com/spf13/cobra"
)

func NewMarkdownCmd() *cobra.Command {
	cfg := &ghCfg{}
	cmd := &cobra.Command{
		Use:   "markdown",
		Short: "Renders an arbitrary markdown document.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			md := internal.Must(execMarkdown(cfg, args[0]))
			_ = internal.Must(cli.Msg(md))
		},
	}

	return cmd
}

func execMarkdown(cfg *ghCfg, md string) (string, error) {
	setHostOwnerRepo(cfg, cfg.host, cfg.owner, cfg.repo)
	cfg.client = newClient()

	opts := &github.MarkdownOptions{Mode: "markdown", Context: ""}
	text, _, err := cfg.client.Markdown.Render(getCtx(cfg), md, opts)
	return text, err
}
