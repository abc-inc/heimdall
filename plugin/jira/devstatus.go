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
	"net/http"
	"os"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
)

type jiraDevStatusCfg struct {
	issueID string
	jiraCfg
}

func NewDevStatusCmd() *cobra.Command {
	cfg := jiraDevStatusCfg{jiraCfg: jiraCfg{baseURL: os.Getenv("JIRA_API_URL"), timeout: 30 * time.Second}}
	cmd := &cobra.Command{
		Use:   "dev-status",
		Short: "Get details about the development status.",
		Example: heredoc.Doc(`
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

			vs, _, err := getDetails(client, cfg)
			if err == nil {
				console.Fmtln(vs)
			}
			return err
		},
	}

	cmd.Flags().StringVarP(&cfg.issueID, "issue-id", "i", cfg.issueID, "ID of the Jira issue")
	addCommonFlags(cmd, &cfg.jiraCfg)

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkFlagRequired("issue-id"))
	return cmd
}

func getDetails(client *jira.Client, cfg jiraDevStatusCfg) (body map[string]any, resp *jira.Response, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()
	req := internal.Must(client.NewRequestWithContext(ctx, http.MethodGet,
		"/rest/dev-status/1.0/issue/detail?issueId="+cfg.issueID+"&applicationType=githube&dataType=repository", nil))
	resp, err = client.Do(req, &body)
	return
}
