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
	start  int
	cql    string
	expand string
}

func NewSearchCmd() *cobra.Command {
	cfg := confluenceSearchCfg{confluenceCfg: confluenceCfg{
		baseURL: os.Getenv("CONFLUENCE_API_URL"),
		timeout: 30 * time.Second},
		expand: "content.body.storage",
		limit:  1,
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
			s := search(cfg)
			if cfg.limit == 1 && len(s.Results) == 1 {
				cli.Fmtln(s.Results[0])
			} else {
				cli.Fmtln(s.Results)
			}
		},
	}

	cmd.Flags().StringVar(&cfg.expand, "expand", cfg.expand, "Expand specific entities in the returned list")
	cmd.Flags().StringVar(&cfg.cql, "filter", cfg.cql, "CQL query for searching")
	cmd.Flags().IntVar(&cfg.limit, "limit", cfg.limit, "Maximum items to return")
	cmd.Flags().IntVar(&cfg.start, "start", cfg.start, "Starting index of the returned list")
	addCommonFlags(cmd, &cfg.confluenceCfg)

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkFlagRequired("filter"))
	return cmd
}

func search(cfg confluenceSearchCfg) *goconfluence.Search {
	api := internal.Must(newClient(cfg.baseURL, cfg.token))
	s := internal.Must(api.Search(goconfluence.SearchQuery{
		CQL:    cfg.cql,
		Limit:  cfg.limit,
		Start:  cfg.start,
		Expand: []string{cfg.expand},
	}))

	return s
}
