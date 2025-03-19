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
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
)

type jiraIssueCfg struct {
	jql string
	jiraCfg
}

func NewIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Work with Jira issues",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(
		NewViewIssueCmd(),
		NewListIssuesCmd(),
	)

	return cmd
}

func NewViewIssueCmd() *cobra.Command {
	cfg := jiraIssueCfg{jiraCfg: *newJiraCfg()}
	cmd := &cobra.Command{
		Use:   "view",
		Short: "View a Jira issue by its ID or key",
		Example: heredoc.Doc(`
			heimdall jira issue view ABC-123
		`),
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.Must(newClient(cfg.apiURL, cfg.token))
			cli.Fmtln(handle(client.Issue.Get(args[0], &jira.GetQueryOptions{
				Fields: strings.Join(cfg.jiraCfg.opts.Fields, ","),
				Expand: cfg.jiraCfg.opts.Expand,
			})))
		},
	}

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	addCommonFlags(cmd, &cfg.jiraCfg)

	return cmd
}

func NewListIssuesCmd() *cobra.Command {
	cfg := jiraIssueCfg{jiraCfg: *newJiraCfg()}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Search Jira issues",
		Example: heredoc.Doc(`
			# list the top 10 open bugs and return all details (including nested fields) about assignee, priority and summary.
			heimdall jira issue list --filter "project = ABC AND type = Bug AND resolution = Unresolved ORDER BY priority DESC, updated DESC" \
			    --fields 'assignee,priority,summary' --max-results 10

			# list all stories and output a custom formatted JSON for summarizing the release notes
			heimdall jira issue list --filter "project = ABC AND type = Story AND fixVersion='ABC 1.2' ORDER BY id" \
			    --jq ".[] | {id: .key, summary: .fields.summary, status: .fields.status.name}"
		`),
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.Must(newClient(cfg.apiURL, cfg.token))
			cli.Fmtln(handle(client.Issue.Search(cfg.jql, cfg.opts)))
		},
	}

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	addCommonFlags(cmd, &cfg.jiraCfg)
	addSearchFlags(cmd, cfg.opts)

	cmd.Flags().StringVar(&cfg.jql, "filter", cfg.jql, "JQL query for searching")
	internal.MustNoErr(cmd.MarkFlagRequired("filter"))
	return cmd
}
