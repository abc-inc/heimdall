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

package jira

import (
	"context"
	"os"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
)

type jiraIssueCfg struct {
	jql string
	jiraCfg
}

func NewIssueCmd() *cobra.Command {
	cfg := jiraIssueCfg{jiraCfg: jiraCfg{baseURL: os.Getenv("JIRA_API_URL"), timeout: 30 * time.Second}}
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Search Jira issues",
		Example: heredoc.Doc(`
			# list the top 10 open bugs and return all details (including nested fields) about assignee, priority and summary.
			heimdall jira issue --filter "project = ABC AND type = Bug AND resolution = Unresolved ORDER BY priority DESC, updated DESC" \
			    --fields 'assignee,priority,summary' --max-results 10
			
			# list all stories and output a custom formatted JSON for summarizing the release notes
			heimdall jira issue --filter "project = ABC AND type = Story AND fixVersion='ABC 1.2' ORDER BY id" \
			    --jq ".[] | {id: .key, summary: .fields.summary, status: .fields.status.name}"
		`),
		Args: cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			if cfg.token == "" {
				cfg.token = os.Getenv("JIRA_TOKEN")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient(cfg.baseURL, cfg.token)
			if err != nil {
				return err
			}

			is, _, err := listIssues(client, cfg)
			if err == nil {
				console.Fmtln(is)
			}
			return err
		},
	}

	cmd.Flags().StringVar(&cfg.jql, "filter", cfg.jql, "JQL query for searching")
	addCommonFlags(cmd, &cfg.jiraCfg)

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkFlagRequired("filter"))
	cfg.opts = addSearchOpts(cmd)
	return cmd
}

// listIssues searches for Jira issues matching the given JQL query and search options.
func listIssues(client *jira.Client, cfg jiraIssueCfg) ([]jira.Issue, *jira.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()
	return client.Issue.SearchWithContext(ctx, cfg.jql, cfg.opts)
}
