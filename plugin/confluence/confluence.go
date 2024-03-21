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
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
	goconfluence "github.com/virtomize/confluence-go-api"
)

type confluenceCfg struct {
	console.OutCfg
	token   string
	baseURL string
	timeout time.Duration
	opts    *jira.SearchOptions
}

type Topic struct {
	Title     string
	Level     int
	SubTopics []Topic
}

func NewConfluenceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confluence",
		Short: "Query Confluence",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(
		NewCreateCmd(),
		NewUpdateCmd(),
		NewSearchCmd(),
	)

	return cmd
}

// newClient creates a new Confluence client.
func newClient(baseURL, token string) (*goconfluence.API, error) {
	if baseURL == "" || token == "" {
		return nil, fmt.Errorf("both, url and token must be defined")
	}
	return goconfluence.NewAPI(baseURL, "", token)
}

func addCommonFlags(cmd *cobra.Command, cfg *confluenceCfg) {
	cmd.Flags().DurationVarP(&cfg.timeout, "timeout", "T", cfg.timeout, "Set the network timeout in seconds")
	cmd.Flags().StringVarP(&cfg.baseURL, "url", "u", cfg.baseURL, "Define the Confluence base URL")
	cmd.Flags().StringVar(&cfg.token, "token", "", "Set the Confluence access token to use")
}

func readContent(name string) string {
	f := internal.Must(res.Open(name))
	defer func() { _ = f.Close() }()
	b := strings.Builder{}
	_ = internal.Must(io.Copy(&b, f))
	return b.String()
}
