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
	"os"
	"strings"
	"time"

	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
)

type jiraCfg struct {
	cli.OutCfg
	apiURL string
	token  string
	opts   *jira.SearchOptions
}

func newJiraCfg() *jiraCfg {
	return &jiraCfg{
		apiURL: os.Getenv("JIRA_API_URL"),
		token:  os.Getenv("JIRA_TOKEN"),
	}
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
	httpClient := tp.Client()
	httpClient.Timeout = 30 * time.Second
	return jira.NewClient(httpClient, baseURL(apiURL))
}

func addCommonFlags(cmd *cobra.Command, cfg *jiraCfg) {
	if cfg.opts == nil {
		cfg.opts = &jira.SearchOptions{MaxResults: 50}
	}
	cmd.Flags().StringVar(&cfg.opts.Expand, "expand", cfg.opts.Expand, "Expand specific sections in the returned list")
	cmd.Flags().StringSliceVar(&cfg.opts.Fields, "fields", cfg.opts.Fields, "List of fields to return. By default, all fields are returned.")
}

func addSearchFlags(cmd *cobra.Command, opts *jira.SearchOptions) *jira.SearchOptions {
	if opts == nil {
		opts = &jira.SearchOptions{MaxResults: 50}
	}
	cmd.Flags().IntVar(&opts.StartAt, "start-at", opts.StartAt, "Starting index of the returned list")
	cmd.Flags().IntVar(&opts.MaxResults, "max-results", opts.MaxResults, "Maximum number of items to return per page")
	return opts
}

func baseURL(url string) string {
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, "/api")
	url = strings.TrimSuffix(url, "/rest")
	return url
}

func handle[T any](ret T, resp *jira.Response, err error) T {
	closeBody(resp)
	internal.MustNoErr(err)
	return ret
}

func closeBody(resp *jira.Response) {
	if resp != nil {
		_ = resp.Body.Close()
	}
}
