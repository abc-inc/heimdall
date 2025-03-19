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

//go:build !no_atlassian && !no_confluence

package confluence

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	goconfluence "github.com/virtomize/confluence-go-api"
)

type confluenceSearchCfg struct {
	confluenceCfg
	limit  int
	offset int
	cql    string
	expand string
	export string
	file   string
}

func NewSearchCmd() *cobra.Command {
	cfg := confluenceSearchCfg{confluenceCfg: confluenceCfg{
		baseURL: os.Getenv("CONFLUENCE_API_URL"),
		timeout: 30 * time.Second},
		expand: "content.body.storage",
		limit:  10,
	}

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for Confluence pages",
		Args:  cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			if cfg.token == "" {
				cfg.token = os.Getenv("CONFLUENCE_TOKEN")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if zerolog.GlobalLevel() == zerolog.TraceLevel || zerolog.GlobalLevel() == zerolog.DebugLevel {
				goconfluence.SetDebug(true)
			}
			if cfg.export == "html" {
				cfg.expand = "content.body.view," + cfg.expand
			}
			s := search(cfg)
			if cfg.export != "" {
				export(s, cfg)
			} else {
				cli.Fmtln(s)
			}
		},
	}

	cmd.Flags().StringVar(&cfg.expand, "expand", cfg.expand, "Expand specific entities in the returned list")
	cmd.Flags().StringVarP(&cfg.file, "file", "O", cfg.file, "File to save the page to (use '-' for standard output)")
	cmd.Flags().StringVar(&cfg.cql, "filter", cfg.cql, "CQL query for searching")
	cmd.Flags().IntVar(&cfg.limit, "limit", cfg.limit, "Maximum number of items to return")
	cmd.Flags().IntVar(&cfg.offset, "offset", cfg.offset, "Starting index of the returned list")
	cmd.Flags().StringVar(&cfg.export, "export", cfg.export, "Export page (supported modes: pdf, html)")
	addCommonFlags(cmd, &cfg.confluenceCfg)

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkFlagRequired("filter"))
	cmd.MarkFlagsMutuallyExclusive("export", "limit")
	cmd.MarkFlagsMutuallyExclusive("export", "output")
	return cmd
}

func search(cfg confluenceSearchCfg) *goconfluence.Search {
	api := internal.Must(newClient(cfg.baseURL, cfg.token))
	s := internal.Must(api.Search(goconfluence.SearchQuery{
		CQL:    cfg.cql,
		Limit:  cfg.limit,
		Start:  cfg.offset,
		Expand: []string{cfg.expand},
	}))

	return s
}

func export(s *goconfluence.Search, cfg confluenceSearchCfg) {
	internal.MustOkMsgf(1, len(s.Results) == 1, "expected 1 page, but found %d", len(s.Results))

	switch cfg.export {
	case "html":
		exportHTML(cfg, s.Results[0])
	case "pdf":
		exportPDF(cfg, s.Results[0])
	default:
		internal.MustOkMsgf(cfg.export, false, "invalid export format: %s", cfg.export)
	}
}

func exportPDF(cfg confluenceSearchCfg, page goconfluence.Results) {
	u := createPDFExportURL(cfg.baseURL, page.Content.ID)
	req := internal.Must(http.NewRequest("GET", u, nil))
	api := internal.Must(newClient(cfg.baseURL, cfg.token))
	data := internal.Must(api.Request(req))

	if cfg.file == "-" {
		internal.Must(os.Stdout.Write(data))
		return
	}

	if cfg.file == "" {
		cfg.file = page.Content.Title + ".pdf"
	}

	internal.MustNoErr(os.WriteFile(cfg.file, data, 0640))
}

func exportHTML(cfg confluenceSearchCfg, page goconfluence.Results) {
	data := page.Content.Body.View.Value

	if cfg.file == "-" {
		internal.Must(os.Stdout.WriteString(data))
		return
	}

	if cfg.file == "" {
		cfg.file = page.Content.Title + ".html"
	}

	f := internal.Must(os.OpenFile(cfg.file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640))
	defer func() { _ = f.Close() }()
	_ = internal.Must(io.WriteString(f, data))
}

func createPDFExportURL(baseURL string, pageID string) string {
	u := internal.Must(url.Parse(baseURL))
	u = u.JoinPath("../../spaces/flyingpdf/pdfpageexport.action")
	u.RawQuery = "pageId=" + pageID
	return u.String()
}
