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

//go:build !no_github

//go:generate go run github.com/abc-inc/heimdall/tools/cmddoc github.com/google/go-github/v69@v69.2.0/github/pulls\*.go ../../docs

package github

import (
	"reflect"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/google/go-github/v69/github"
	"github.com/spf13/cobra"
)

func NewPRCmd() *cobra.Command {
	var branch string
	var protected bool
	state := "all"
	cfg := newGHCfg()
	cfg.branch = &branch
	cfg.protected = &protected
	cfg.visibility = "all"
	cfg.affiliation = "owner,collaborator,organization_member"
	cfg.state = &state

	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Handles communication with the pull request related methods.",
		Example: heredoc.Doc(`
		`),
		Args: cobra.ExactArgs(0),
	}

	ops := []string{
		"comment",
		"comments",
		"commits",
		"files",
		"get",
		"list",
		"list-comments",
		"pull-requests-with-commit",
		"raw",
		"review",
		"reviews",
		"review-comments",
		"reviewers",
	}

	svcTyp := reflect.TypeOf(&github.PullRequestsService{})
	cmd.AddCommand(createCmds(cfg, svcTyp, execPR, ops)...)
	for _, sub := range cmd.Commands() {
		if sub.Name() != "all" {
			addRepoFlags(cfg, sub)
		}

		switch sub.Name() {
		case "comment":
			sub.Flags().Int64Var(&cfg.id, "number", cfg.id, "Pull request number.")
		case "comments":
			sub.Flags().Int64Var(&cfg.id, "number", cfg.id, "Pull request number.")
		case "commits":
			sub.Flags().Int64Var(&cfg.id, "number", cfg.id, "Pull request number.")
		case "files":
			sub.Flags().Int64Var(&cfg.id, "number", cfg.id, "Pull request number.")
		case "get":
			sub.Flags().Int64Var(&cfg.id, "number", cfg.id, "Pull request number.")
		case "list":
			sub.Flags().StringVar(cfg.state, "state", *cfg.state, "Either open, closed, or all to filter by state.")
			sub.Flags().StringVar(&cfg.head, "head", cfg.head, "Filter pulls by head user or head organization and branch name in the format of user:ref-name or organization:ref-name.")
			sub.Flags().StringVar(&cfg.base, "base", cfg.base, "Filter pulls by base branch name. Example: gh-pages.")
			sub.Flags().StringVar(cfg.sort, "sort", *cfg.sort, "What to sort results by. popularity will sort by the number of comments. long-running will sort by date created and will limit the results to pull requests that have been open for more than a month and have had activity within the past month.")
			sub.Flags().StringVar(cfg.direction, "direction", *cfg.direction, "The direction of the sort. Default: desc when sort is created or sort is not specified, otherwise asc.")
		case "list-comments":
			sub.Flags().StringVar(cfg.sort, "sort", *cfg.sort, "What to sort results by. popularity will sort by the number of comments. long-running will sort by date created and will limit the results to pull requests that have been open for more than a month and have had activity within the past month.")
			sub.Flags().StringVar(cfg.direction, "direction", *cfg.direction, "The direction of the sort. Default: desc when sort is created or sort is not specified, otherwise asc.")
		case "pull-requests-with-commit":
			sub.Flags().Int64Var(&cfg.id, "number", cfg.id, "Pull request number.")
			sub.Flags().StringVar(&cfg.sha, "sha", cfg.sha, "The SHA of the commit that needs a review.")
		case "raw":
			sub.Flags().Int64Var(&cfg.id, "number", cfg.id, "Pull request number.")
		case "review":
			sub.Flags().Int64Var(&cfg.id, "number", cfg.id, "Pull request number.")
			sub.Flags().Int64Var(&cfg.id, "review-id", cfg.reviewID, "The unique identifier of the review.")
		case "reviews":
			sub.Flags().Int64Var(&cfg.id, "number", cfg.id, "Pull request number.")
		case "review-comments":
			sub.Flags().Int64Var(&cfg.id, "number", cfg.id, "Pull request number.")
		case "reviewers":
			sub.Flags().Int64Var(&cfg.id, "number", cfg.id, "Pull request number.")
		}
	}

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func execPR(cfg *ghCfg, cmd *cobra.Command) (x any, err error) {
	setHostOwnerRepo(cfg, cfg.host, cfg.owner, cfg.repo)
	cfg.client = newClient()
	svc := cfg.client.PullRequests

	switch cmd.Name() {
	case "comment":
		x, _, err = svc.GetComment(getCtx(cfg), cfg.owner, cfg.repo, cfg.id)
	case "comments":
		x, _, err = svc.ListComments(getCtx(cfg), cfg.owner, cfg.repo, int(cfg.id), &github.PullRequestListCommentsOptions{
			Sort:        *cfg.sort,
			Direction:   *cfg.direction,
			Since:       cfg.sinceTime,
			ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "commits":
		x, _, err = svc.ListCommits(getCtx(cfg), cfg.owner, cfg.repo, int(cfg.id),
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "files":
		x, _, err = svc.ListFiles(getCtx(cfg), cfg.owner, cfg.repo, int(cfg.id),
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "get":
		x, _, err = svc.Get(getCtx(cfg), cfg.owner, cfg.repo, int(cfg.id))
	case "list":
		x, _, err = svc.List(getCtx(cfg), cfg.owner, cfg.repo, &github.PullRequestListOptions{
			State:       *cfg.state,
			Head:        cfg.head,
			Base:        cfg.base,
			Sort:        *cfg.sort,
			Direction:   *cfg.direction,
			ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "list-comments":
		x, _, err = svc.ListComments(getCtx(cfg), cfg.owner, cfg.repo, int(cfg.id), &github.PullRequestListCommentsOptions{
			Sort:        *cfg.sort,
			Direction:   *cfg.direction,
			Since:       cfg.sinceTime,
			ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "pull-requests-with-commit":
		x, _, err = svc.ListPullRequestsWithCommit(getCtx(cfg), cfg.owner, cfg.repo, cfg.sha,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "raw":
		x, _, err = svc.GetRaw(getCtx(cfg), cfg.owner, cfg.repo, int(cfg.id), github.RawOptions{Type: github.Diff})
	case "review":
		x, _, err = svc.GetReview(getCtx(cfg), cfg.owner, cfg.repo, int(cfg.id), cfg.reviewID)
	case "reviews":
		x, _, err = svc.ListReviews(getCtx(cfg), cfg.owner, cfg.repo, int(cfg.id),
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "review-comments":
		x, _, err = svc.ListReviewComments(getCtx(cfg), cfg.owner, cfg.repo, int(cfg.id), cfg.reviewID,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "reviewers":
		x, _, err = svc.ListReviewers(getCtx(cfg), cfg.owner, cfg.repo, int(cfg.id),
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	}
	return
}
