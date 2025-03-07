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
	"net/http"

	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
)

type jiraDevStatusCfg struct {
	issueID string
	jiraCfg
}

func NewDevStatusCmd() *cobra.Command {
	cfg := jiraDevStatusCfg{jiraCfg: *newJiraCfg()}
	cmd := &cobra.Command{
		Use:   "dev-status",
		Short: "Get details about the development status.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient(cfg.apiURL, cfg.token)
			if err != nil {
				return err
			}

			vs, _, err := getDetails(client, cfg)
			if err == nil {
				cli.Fmtln(vs)
			}
			return err
		},
	}

	cmd.Flags().StringVarP(&cfg.issueID, "issue-id", "i", cfg.issueID, "ID of the Jira issue")

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkFlagRequired("issue-id"))
	return cmd
}

func getDetails(client *jira.Client, cfg jiraDevStatusCfg) (body map[string]any, resp *jira.Response, err error) {
	req := internal.Must(client.NewRequest(http.MethodGet,
		"/rest/dev-status/1.0/issue/detail?issueId="+cfg.issueID+"&applicationType=githube&dataType=repository", nil))
	resp, err = client.Do(req, &body)
	return
}
