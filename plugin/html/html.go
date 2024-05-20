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

//go:build !no_html

package html

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/spf13/cobra"
)

type htmlCfg struct {
	file     string
	selector string
}

func NewHTMLCmd() *cobra.Command {
	cfg := htmlCfg{}

	cmd := &cobra.Command{
		Use:   "html [flags] <file>",
		Short: "Load HTML files and process them",
		Example: heredoc.Doc(`
			heimdall html --query 'h1' index.html
		`),
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.file = args[0]
			_ = internal.Must(console.Msg(processHTML(cfg)))
		},
	}

	cmd.DisableFlagsInUseLine = true
	cmd.Flags().StringVar(&cfg.selector, "query", cfg.selector, "CSS selector to filter the HTML")

	internal.MustNoErr(cmd.MarkFlagRequired("query"))
	return cmd
}

func processHTML(cfg htmlCfg) string {
	h := readHTML(cfg.file)
	if cfg.selector != "" {
		return internal.Must(h.Find(cfg.selector).Html())
	}
	return internal.Must(h.Html())
}

func readHTML(name string) *goquery.Document {
	r := internal.Must(res.Open(name))
	defer func() { _ = r.Close() }()
	return internal.Must(goquery.NewDocumentFromReader(r))
}
