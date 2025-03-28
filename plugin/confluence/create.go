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
	"log"
	"os"
	"time"

	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	goconfluence "github.com/virtomize/confluence-go-api"
)

func NewCreateCmd() *cobra.Command {
	cfg := confluenceUpdateCfg{
		confluenceCfg: confluenceCfg{baseURL: os.Getenv("CONFLUENCE_API_URL"), timeout: 30 * time.Second},
		expand:        "content.ancestors,content.body.storage,content.space",
	}

	cmd := &cobra.Command{
		Use:   "create [flags] <file>",
		Short: "Create a Confluence page",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			if cfg.token == "" {
				cfg.token = os.Getenv("CONFLUENCE_TOKEN")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if zerolog.GlobalLevel() == zerolog.TraceLevel || zerolog.GlobalLevel() == zerolog.DebugLevel {
				goconfluence.SetDebug(true)
			}
			cfg.file = args[0]
			cli.Fmtln(create(cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.expand, "expand", cfg.expand, "Expand specific entities in the returned list")
	cmd.Flags().StringVar(&cfg.cql, "filter", cfg.cql, "CQL query for searching")
	cmd.Flags().StringVar(&cfg.title, "title", cfg.title, "Title of the page")
	addCommonFlags(cmd, &cfg.confluenceCfg)

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkFlagRequired("filter"))
	internal.MustNoErr(cmd.MarkFlagRequired("title"))
	return cmd
}

func create(cfg confluenceUpdateCfg) *goconfluence.Content {
	s := search(confluenceSearchCfg{
		confluenceCfg: cfg.confluenceCfg,
		limit:         2,
		offset:        0,
		cql:           cfg.cql,
		expand:        cfg.expand,
	})
	if s.Size != 1 {
		log.Fatal("Exactly one page must be found")
	}

	p := s.Results[0]

	api := internal.Must(newClient(cfg.baseURL, cfg.token))

	data := &goconfluence.Content{
		Type:      "page",
		Title:     cfg.title,
		Ancestors: []goconfluence.Ancestor{{p.Content.ID}},
		Body: goconfluence.Body{
			Storage: goconfluence.Storage{
				Value:          readContent(cfg.file),
				Representation: "storage",
			},
		},
		Version: &goconfluence.Version{Number: 1},
		Space:   p.Content.Space,
	}

	return internal.Must(api.CreateContent(data))
}
