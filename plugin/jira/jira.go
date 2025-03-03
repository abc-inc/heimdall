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
	"fmt"
	"strings"
	"time"

	"github.com/abc-inc/heimdall/cli"
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
)

type jiraCfg struct {
	cli.OutCfg
	apiURL  string
	token   string
	timeout time.Duration
	opts    *jira.SearchOptions
}

const envHelp = `
JIRA_API_URL  https://jira.company.corp/rest/api
JIRA_TOKEN    <PERSONAL_ACCESS_TOKEN>
`

func NewJiraCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "jira <subcommand>",
		Short:       "Query Jira",
		GroupID:     cli.ServiceGroup,
		Args:        cobra.ExactArgs(0),
		Annotations: map[string]string{"help:environment": envHelp},
	}

	cmd.AddCommand(
		NewDevStatusCmd(),
		NewIssueCmd(),
		NewVersionCmd(),
	)

	return cmd
}

// newClient creates a new Jira client.
func newClient(apiURL, token string) (*jira.Client, error) {
	if apiURL == "" || token == "" {
		return nil, fmt.Errorf("JIRA_API_URL and JIRA_TOKEN must be defined")
	}
	tp := jira.PATAuthTransport{Token: token}
	return jira.NewClient(tp.Client(), baseURL(apiURL))
}

func addCommonFlags(cmd *cobra.Command, cfg *jiraCfg) {
	cmd.Flags().DurationVarP(&cfg.timeout, "timeout", "T", cfg.timeout, "Set the network timeout in seconds")
}

func addSearchOpts(cmd *cobra.Command) *jira.SearchOptions {
	opts := &jira.SearchOptions{StartAt: 0, MaxResults: 50, Expand: "", Fields: nil, ValidateQuery: ""}
	cmd.Flags().IntVar(&opts.StartAt, "start-at", opts.StartAt, "Starting index of the returned list")
	cmd.Flags().IntVar(&opts.MaxResults, "max-results", opts.MaxResults, "Maximum number of items to return per page")
	cmd.Flags().StringVar(&opts.Expand, "expand", opts.Expand, "Expand specific sections in the returned list")
	cmd.Flags().StringSliceVar(&opts.Fields, "fields", opts.Fields, "List of fields to return. By default, all fields are returned.")
	cmd.Flags().StringVar(&opts.ValidateQuery, "validation", opts.ValidateQuery, "Whether to validate and how strictly to treat the validation (strict/warn) (default strict)")
	return opts
}

func baseURL(url string) string {
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, "/api")
	url = strings.TrimSuffix(url, "/rest")
	return url
}
