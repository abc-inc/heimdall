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

//go:build !no_atlassian && !no_jira

package jira

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
)

type jiraVersionCfg struct {
	project string
	jiraCfg
}

func NewVersionCmd() *cobra.Command {
	cfg := jiraVersionCfg{jiraCfg: *newJiraCfg()}
	cmd := &cobra.Command{
		Use:   "version",
		Short: "List project versions",
		Example: heredoc.Doc(`
			# list all versions formatted as CSV
			heimdall jira version --project ABC --output csv

			# get details about a specific version
			heimdall jira version --project ABC --jq '.[] | select(.name == "ABC 1.2")'
		`),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient(cfg.apiURL, cfg.token)
			if err != nil {
				return err
			}

			vs, _, err := listVersions(client, cfg)
			if err == nil {
				cli.Fmtln(vs)
			}
			return err
		},
	}

	cmd.Flags().StringVarP(&cfg.project, "project", "p", cfg.project, "Project name")

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkFlagRequired("project"))
	return cmd
}

// listVersions returns all versions of a given project.
func listVersions(client *jira.Client, cfg jiraVersionCfg) ([]jira.Version, *jira.Response, error) {
	p, resp, err := client.Project.Get(cfg.project)
	if err != nil {
		return nil, resp, err
	}
	return p.Versions, resp, err
}
